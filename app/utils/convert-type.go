package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/jinzhu/copier"
)

func ConvertToType[T any](input any) (T, error) {
	var result T

	// Marshal struct เป็น JSON
	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("marshal error: %w", err)
	}

	// Unmarshal data เป็น Struct T ที่ระบุไว้
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("unmarshal error: %w", err)
	}

	return result, nil
}

func ToReader(data any) (io.Reader, error) {
	switch v := reflect.ValueOf(data); v.Kind() {
	case reflect.String:
		return bytes.NewBufferString(v.String()), nil
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Ptr:
		jsonStrBody, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(jsonStrBody), nil
	}
	return nil, fmt.Errorf("unsupported type: %s", reflect.TypeOf(data).String())
}

func CopyNTimeToUnix(to any, from any) error {
	return copier.CopyWithOption(to, from, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
		Converters: []copier.TypeConverter{
			{
				SrcType: time.Time{},
				DstType: int64(0),
				Fn: func(src any) (any, error) {
					return src.(time.Time).Unix(), nil
				},
			},
		},
	})
}
