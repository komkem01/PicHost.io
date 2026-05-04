package config

import (
	"fmt"
	"reflect"
)

type Config[C any] struct {
	App
	Val *C
}

type App interface {
	Hostname() string
	AppName() string
	Version() string
	Environment() string
	Debug() bool
}

func Conf[C any](app App) *Config[C] {
	var c *C
	conf := reflect.ValueOf(app).MethodByName("Config").Call(nil)[0].Elem()
	exists := false
	for i := range conf.NumField() {
		elem := conf.Field(i)
		if elem.Addr().Type() == reflect.TypeOf(c) {
			c = elem.Addr().Interface().(*C)
			exists = true
			break
		}
	}
	if !exists {
		panic(fmt.Sprintf("Config not found for %q", reflect.TypeOf(c)))
	}
	return &Config[C]{
		App: app,
		Val: c,
	}
}
