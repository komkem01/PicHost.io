package modules

import (
	"fmt"
	"reflect"
)

func (m *Modules) Map() map[string]any {
	if m == nil {
		return nil
	}
	ret := make(map[string]any)
	for i, v := range reflect.VisibleFields(reflect.TypeOf(m).Elem()) {
		intf := reflect.ValueOf(m).Elem().FieldByName(v.Name).Interface()
		if !reflect.ValueOf(intf).IsNil() {
			ret[fmt.Sprintf("%d.%s", i, v.Name)] = reflect.ValueOf(m).Elem().FieldByName(v.Name).Interface()
		}
	}
	return ret
}

func Map() map[string]any {
	return mod.Map()
}
