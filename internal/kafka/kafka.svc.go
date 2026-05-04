package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"mcop/app/utils/syncx"
	"mcop/internal/config"
	kafkainf "mcop/internal/kafka/inf"
	"mcop/internal/log"
	"mcop/internal/provider"

	"github.com/IBM/sarama"
)

type SyncProducerWithErr struct {
	sarama.SyncProducer
	Error error
}

func (s *SyncProducerWithErr) Sync() (sarama.SyncProducer, error) {
	return s.SyncProducer, s.Error
}

var _ provider.Close = (*Service)(nil)

type Service struct {
	cli              sarama.Client
	broker           *sarama.Broker
	syncProducerPool *syncx.Pool[SyncProducerWithErr]
}
type Config struct {
	CaPath   string
	CertPath string
	KeyPath  string
	Brokers  string
	Section  string
}

func newService(conf *config.Config[Config]) *Service {
	log := log.Default()
	confVal := conf.Val

	if confVal.CaPath == `` || confVal.CertPath == `` || confVal.KeyPath == `` || confVal.Brokers == `` {
		panic("Kafka configuration is not set properly. Please check your environment variables or configuration file.")
	}

	bCa, err := os.ReadFile(confVal.CaPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read CA file: %s", err))
	}

	cert, err := tls.LoadX509KeyPair(confVal.CertPath, confVal.KeyPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load key pair: %s", err))
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(bCa)

	clientTLS := tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
	}
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = conf.AppName()
	kafkaConfig.Net.TLS.Enable = true
	kafkaConfig.Net.TLS.Config = &clientTLS
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Offsets.AutoCommit.Enable = false

	addrs := strings.Split(confVal.Brokers, " ")
	client, err := sarama.NewClient(addrs, kafkaConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Kafka client: %s", err))
	}

	tmp := *kafkaConfig
	kafkaProducerConfig := &tmp
	kafkaProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaProducerConfig.Producer.Return.Successes = true

	spool := syncx.NewPool(func() *SyncProducerWithErr {
		producer, err := sarama.NewSyncProducer(addrs, kafkaProducerConfig)
		if err != nil {
			log.Errf("Failed to create producer: %s", err)
			return &SyncProducerWithErr{Error: err}
		}
		return &SyncProducerWithErr{producer, nil}
	})
	producerWaitErr := spool.Get()
	if producerWaitErr.Error != nil {
		log.Errf("Failed to create producer: %s", producerWaitErr.Error)
		panic(producerWaitErr.Error)
	}

	spool.Put(producerWaitErr)

	broker, err := client.Controller()
	if err != nil {
		log.Errf("Failed to get controller: %s", err)
		panic(err)
	}
	svc := &Service{
		cli:              client,
		broker:           broker,
		syncProducerPool: spool,
	}

	return svc
}

func (s *Service) CreateTopic(ctx context.Context, topic string) error {
	log := log.WithCtx(ctx)
	broker, err := s.cli.RefreshController()
	if err != nil {
		log.Errf("Failed to get controller: %s", err)
		return err
	}
	req := &sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     3,
				ReplicationFactor: 2,
			},
		},
		Timeout: 30 * time.Second,
	}
	ctr, err := broker.CreateTopics(req)
	if err != nil {
		log.Errf("Failed to create topic: %s", err)
		return err
	}
	tErr, ok := ctr.TopicErrors[topic]
	if ok && tErr.Err != sarama.ErrNoError {
		log.Errf("Failed to create topic(%s): %s", topic, tErr.Err)
	}
	return nil
}

func (s *Service) DeleteTopic(ctx context.Context, topic string) error {
	log := log.WithCtx(ctx)
	broker, err := s.cli.RefreshController()
	if err != nil {
		log.Errf("Failed to get controller: %s", err)
		return err
	}
	req := &sarama.DeleteTopicsRequest{
		Topics:  []string{topic},
		Timeout: 1 * time.Second,
	}
	dtr, err := broker.DeleteTopics(req)
	if err != nil {
		log.Errf("Failed to delete topic: %s", err)
		return err
	}
	tErr, ok := dtr.TopicErrorCodes[topic]
	if ok && tErr != sarama.ErrNoError {
		log.Errf("Failed to delete topic(%s): %s", topic, tErr.Error())
	}
	return nil
}

func (s *Service) ExistsTopic(ctx context.Context, topic string) (bool, error) {
	log := log.WithCtx(ctx)
	topics, err := s.cli.Topics()
	if err != nil {
		log.Errf("Failed to get topics: %s", err)
		return false, err
	}
	slices.Sort(topics)
	_, ok := slices.BinarySearch(topics, topic)
	return ok, nil
}

func (s *Service) Producer(ctx context.Context, topic, key string, value []byte) error {
	log := log.WithCtx(ctx)
	var sKey sarama.Encoder
	if key != "" {
		sKey = sarama.StringEncoder(key)
	}

	pwe := s.syncProducerPool.Get()
	provider, err := pwe.Sync()
	if err != nil {
		log.Errf("Failed to get producer: %s", err)
		return err
	}
	defer s.syncProducerPool.Put(pwe)

	err = provider.SendMessages([]*sarama.ProducerMessage{{
		Topic: topic,
		Headers: []sarama.RecordHeader{{
			Key:   []byte("content-type"),
			Value: []byte("application/json"),
		}},
		Key:   sKey,
		Value: sarama.ByteEncoder(value),
	}})

	return err
}

func (s *Service) ProduceJSON(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Producer(ctx, topic, key, data)
}

func (s *Service) ConsumerGroup(ctx context.Context, groupID string, topics []string, handler kafkainf.ConsumerGroupHandler) (func(context.Context) error, error) {
	log := log.WithCtx(ctx)

	cg, err := sarama.NewConsumerGroupFromClient(groupID, s.cli)
	if err != nil {
		log.Errf("Failed to create consumer group: %s", err)
		return nil, err
	}

	if err := cg.Consume(ctx, topics, handler); err != nil {
		log.Errf("Failed to consume: %s", err)
		return nil, err
	}
	return func(_ context.Context) error {
		if err := cg.Close(); err != nil {
			log.Errf("Failed to close consumer group: %s", err)
			return err
		}
		return nil
	}, nil
}

// Close implements provider.Close.
func (s *Service) Close(_ context.Context) error {
	return s.cli.Close()
}
