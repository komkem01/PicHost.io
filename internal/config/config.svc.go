package config

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Service[T any] struct {
	hostname    string
	appName     string
	environment string
	debug       bool
	conf        *T
}

func newService[T any](dConf *T) *Service[T] {
	godotenv.Load()
	conf := configWithDefault(dConf)
	confRef := reflect.ValueOf(conf)
	appName := confRef.Elem().FieldByName("AppName").String()
	debug := confRef.Elem().FieldByName("Debug").Bool()
	Environment := confRef.Elem().FieldByName("Environment").String()
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = appName + "-" + hex.EncodeToString(big.NewInt(rand.Int63()).Bytes())
	}
	return &Service[T]{
		hostname:    hostname,
		appName:     appName,
		environment: Environment,
		debug:       debug,
		conf:        conf,
	}
}

// HostName returns the hostname of the service.
func (s *Service[T]) Hostname() string {
	return s.hostname
}

func (s *Service[T]) Config() *T {
	return s.conf
}

func (s *Service[T]) AppName() string {
	return s.appName
}

func (s *Service[T]) Version() string {
	return version
}

func (s *Service[T]) Environment() string {
	return s.environment
}

func (s *Service[T]) Debug() bool {
	return s.debug
}

func configWithDefault[T any](confDefault *T) *T {
	rConfig := reflect.ValueOf(confDefault).Elem()
	t := rConfig.Type()
	for i := 0; i < t.NumField(); i++ {
		key := stringToAllCapsCase(t.Field(i).Name)
		switch t.Field(i).Type.Kind() {
		case reflect.Struct:
			rConfig.Field(i).Set(configStruct(key, rConfig.Field(i), ""))
		default:
			defaultValue := rConfig.Field(i)
			tags := t.Field(i).Tag.Get("conf")
			conf(key, defaultValue, tags)
		}
	}
	return confDefault
}

func configStruct(prefix string, v reflect.Value, tags string) reflect.Value {
	switch v.Kind() {
	case reflect.Struct:
		configStructForStruct(prefix, v)
	case reflect.Map:
		configStructForMap(prefix, v)
	case reflect.Pointer:
		configStructForPtr(prefix, v)
	default:
		configStructForDefault(prefix, v, tags)
	}
	return v
}

func configStructForStruct(prefix string, v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		pf := fieldName[0]
		tags := v.Type().Field(i).Tag.Get("conf")
		if pf >= 'A' && pf <= 'Z' {
			key := fmt.Sprintf("%s_%s", prefix, stringToAllCapsCase(fieldName))
			v.Field(i).Set(configStruct(key, v.Field(i), tags))
		}
	}
}

func configStructForMap(prefix string, v reflect.Value) {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	mapKeyMap := map[string]bool{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			env := strings.Split(env, "=")[0]
			rmPrefix := strings.ReplaceAll(env, prefix+"_", "")
			mapKey := strings.ToLower(strings.Split(rmPrefix, "_")[0])
			mapKeyMap[mapKey] = true
		}
	}

	mapKeys := []string{}
	for k := range mapKeyMap {
		mapKeys = append(mapKeys, k)
	}

	for _, mapKey := range mapKeys {
		key := fmt.Sprintf("%s_%s", prefix, strings.ToUpper(mapKey))
		kv := v.MapIndex(reflect.ValueOf(mapKey))
		if kv.Kind() == reflect.Invalid {
			kv = reflect.New(v.Type().Elem().Elem())
		}
		v.SetMapIndex(reflect.ValueOf(mapKey), configStruct(key, kv, ""))
	}
}

func configStructForPtr(prefix string, v reflect.Value) {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	v.Elem().Set(configStruct(prefix, v.Elem(), ""))
}

func configStructForDefault(prefix string, defaultValue reflect.Value, tags string) {
	conf(prefix, defaultValue, tags)
}

func conf(key string, fallback reflect.Value, tags string) any {
	if value, ok := os.LookupEnv(key); ok {
		switch kind := fallback.Kind(); kind {
		case reflect.String:
			fallback.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				panic(err)
			}
			fallback.SetInt(i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				panic(err)
			}
			fallback.SetUint(i)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				panic(err)
			}
			fallback.SetFloat(f)
		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				panic(err)
			}
			fallback.SetBool(b)
		default:
			panic(fmt.Sprintf("Unsupported type %s", kind))
		}
	}
	for tag := range strings.SplitSeq(tags, ",") {
		if tag == "required" && fallback.IsZero() {
			panic(fmt.Sprintf("Required config %q is not set", key))
		}
	}
	return fallback
}

func stringToAllCapsCase(str string) string {
	allCapsBuilder := strings.Builder{}
	defer allCapsBuilder.Reset()
	allCapsBuilder.WriteByte(str[0])
	for _, c := range str[1:] {
		if c >= 'A' && c <= 'Z' {
			allCapsBuilder.WriteString("_")
		}
		allCapsBuilder.WriteRune(c)
	}
	return strings.ToUpper(allCapsBuilder.String())
}
