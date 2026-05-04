package storage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer trace.Tracer
	store  entitiesinf.StorageEntity
	conf   *config.Config[Config]
}

type Options struct {
	*config.Config[Config]
	tracer trace.Tracer
	store  entitiesinf.StorageEntity
}

func newService(opt *Options) *Service {
	return &Service{
		tracer: opt.tracer,
		store:  opt.store,
		conf:   opt.Config,
	}
}

type uploadSource struct {
	Reader      io.ReadCloser
	Size        int64
	ContentType string
	Filename    string
}

const storageShortCodeLength = 8

type s3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	PublicBaseURL   string
}

func (s *Service) getS3Config() (*s3Config, error) {
	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
	publicBaseURL := strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL"))

	useSSL := false
	if raw := strings.TrimSpace(os.Getenv("S3_USE_SSL")); raw != "" {
		useSSL = strings.EqualFold(raw, "true")
	}
	if raw := strings.TrimSpace(os.Getenv("S3_USE_S_S_L")); raw != "" {
		useSSL = strings.EqualFold(raw, "true")
	}

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("missing S3 config: require S3_ENDPOINT, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY, S3_BUCKET_NAME")
	}

	if publicBaseURL == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		publicBaseURL = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return &s3Config{
		Endpoint:        endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		BucketName:      bucket,
		UseSSL:          useSSL,
		PublicBaseURL:   strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

func (s *Service) openSource(ctx context.Context, req UploadRequestService) (*uploadSource, error) {
	if req.Path != nil && strings.TrimSpace(*req.Path) != "" {
		path := strings.TrimSpace(*req.Path)
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open file path: %w", err)
		}
		stat, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("stat file path: %w", err)
		}

		contentType := "application/octet-stream"
		if req.MIMEType != nil && strings.TrimSpace(*req.MIMEType) != "" {
			contentType = strings.TrimSpace(*req.MIMEType)
		}

		return &uploadSource{
			Reader:      f,
			Size:        stat.Size(),
			ContentType: contentType,
			Filename:    filepath.Base(path),
		}, nil
	}

	if req.URL != nil && strings.TrimSpace(*req.URL) != "" {
		url := strings.TrimSpace(*req.URL)
		hReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("build source url request: %w", err)
		}

		resp, err := http.DefaultClient.Do(hReq)
		if err != nil {
			return nil, fmt.Errorf("download source url: %w", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("download source url returned status %d", resp.StatusCode)
		}

		contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
		if req.MIMEType != nil && strings.TrimSpace(*req.MIMEType) != "" {
			contentType = strings.TrimSpace(*req.MIMEType)
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		filename := filepath.Base(url)
		if filename == "." || filename == "/" || filename == "" {
			filename = "upload.bin"
		}

		return &uploadSource{
			Reader:      resp.Body,
			Size:        resp.ContentLength,
			ContentType: contentType,
			Filename:    filename,
		}, nil
	}

	return nil, fmt.Errorf("either path or url is required for upload")
}

func (s *Service) uploadToS3(ctx context.Context, src *uploadSource) (objectPath string, objectURL string, size int64, mimeType string, err error) {
	conf, err := s.getS3Config()
	if err != nil {
		return "", "", 0, "", err
	}

	client, err := s.newMinioClient(conf)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("create s3 client: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(src.Filename))
	if ext == "" {
		ext = ".bin"
	}

	objectPath = fmt.Sprintf("uploads/%s/%s%s", time.Now().Format("2006/01/02"), uuid.NewString(), ext)

	uploadInfo, err := client.PutObject(
		ctx,
		conf.BucketName,
		objectPath,
		src.Reader,
		src.Size,
		minio.PutObjectOptions{ContentType: src.ContentType},
	)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("put object: %w", err)
	}

	objectURL = fmt.Sprintf("%s/%s/%s", conf.PublicBaseURL, conf.BucketName, objectPath)
	return objectPath, objectURL, uploadInfo.Size, src.ContentType, nil
}

func (s *Service) newMinioClient(conf *s3Config) (*minio.Client, error) {
	return minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, ""),
		Secure: conf.UseSSL,
	})
}

func (s *Service) presignExpiry() time.Duration {
	const defaultSeconds = 900
	seconds := defaultSeconds
	if raw := strings.TrimSpace(os.Getenv("S3_PRESIGN_EXPIRE_SECONDS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			seconds = v
		}
	}
	return time.Duration(seconds) * time.Second
}

func (s *Service) objectPathFromStorage(item *ent.StorageEntity) (string, error) {
	if item.Path != nil && strings.TrimSpace(*item.Path) != "" {
		return strings.TrimPrefix(strings.TrimSpace(*item.Path), "/"), nil
	}
	if item.URL == nil || strings.TrimSpace(*item.URL) == "" {
		return "", fmt.Errorf("storage has no path or url")
	}

	conf, err := s.getS3Config()
	if err != nil {
		return "", err
	}

	u, err := url.Parse(strings.TrimSpace(*item.URL))
	if err != nil {
		return "", fmt.Errorf("parse storage url: %w", err)
	}

	prefix := "/" + strings.Trim(conf.BucketName, "/") + "/"
	if !strings.HasPrefix(u.Path, prefix) {
		return "", fmt.Errorf("storage url does not contain bucket path")
	}

	return strings.TrimPrefix(u.Path, prefix), nil
}

func (s *Service) signObjectPath(ctx context.Context, objectPath string) (string, error) {
	conf, err := s.getS3Config()
	if err != nil {
		return "", err
	}

	client, err := s.newMinioClient(conf)
	if err != nil {
		return "", fmt.Errorf("create s3 client: %w", err)
	}

	presigned, err := client.PresignedGetObject(
		ctx,
		conf.BucketName,
		strings.TrimPrefix(strings.TrimSpace(objectPath), "/"),
		s.presignExpiry(),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("presign object: %w", err)
	}

	return presigned.String(), nil
}

func (s *Service) PresignStorage(ctx context.Context, item *ent.StorageEntity) (string, error) {
	objectPath, err := s.objectPathFromStorage(item)
	if err != nil {
		return "", err
	}
	return s.signObjectPath(ctx, objectPath)
}

func (s *Service) generateShortCode(length int) (string, error) {
	if length < 6 || length > 8 {
		return "", fmt.Errorf("invalid short code length: %d", length)
	}

	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	b := make([]byte, length)
	max := byte(len(alphabet))

	for i := 0; i < length; i++ {
		buf := []byte{0}
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		b[i] = alphabet[buf[0]%max]
	}

	return string(b), nil
}

func (s *Service) isShortCodeUniqueConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}

	constraint := strings.ToLower(pgErr.ConstraintName)
	message := strings.ToLower(pgErr.Message)
	detail := strings.ToLower(pgErr.Detail)

	return strings.Contains(constraint, "short_code") || strings.Contains(message, "short_code") || strings.Contains(detail, "short_code")
}
