package config

import (
	"google.golang.org/grpc/resolver"
)

func Init(conf *Config) {
	resolver.SetDefaultScheme("dns")
}
