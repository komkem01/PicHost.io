package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Close interface {
	Close(ctx context.Context) error // Before app close
}

type Config map[string]any

var done map[string]any

type byID []string

// Len implements sort.Interface.
func (b byID) Len() int {
	return len(b)
}

// Less implements sort.Interface.
func (b byID) Less(i int, j int) bool {
	ii := b[i]
	jj := b[j]
	iSTR := strings.Split(ii, ".")[0]
	jSTR := strings.Split(jj, ".")[0]
	in, err := strconv.Atoi(iSTR)
	if err != nil {
		panic(err)
	}
	jn, err := strconv.Atoi(jSTR)
	if err != nil {
		panic(err)
	}
	return in < jn
}

// Swap implements sort.Interface.
func (b byID) Swap(i int, j int) {
	b[i], b[j] = b[j], b[i]
}

var _ sort.Interface = (*byID)(nil)

func close(ctx context.Context, in Config) error {
	var err error
	arr := make([]string, 0, len(in))
	for name := range in {
		arr = append(arr, name)
	}
	arrIDs := byID(arr)
	sort.Sort(sort.Reverse(arrIDs))
	for _, name := range arr {
		module := in[name]
		if module == nil {
			continue
		}
		eModule := reflect.ValueOf(module).Elem()
		field := eModule.FieldByName("Svc")
		if !field.IsValid() {
			continue
		}

		vvv, ok := field.Elem().Interface().(Close)
		if ok && vvv != nil {
			if errClose := vvv.Close(ctx); errClose != nil {
				fmt.Println(err)
				err = errors.Join(err, errClose)
			}
		}
		done[name] = module
	}
	return err
}

func (conf *Config) Close(ctx context.Context) error {
	if conf == nil {
		return nil
	}
	done = map[string]any{}
	if err := close(ctx, *conf); err != nil {
		return err
	}
	done = nil
	return nil
}
