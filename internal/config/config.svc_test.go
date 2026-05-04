package config

import "testing"

func Test_stringToAllCapsCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test App", args{"App"}, "APP"},
		{"Test AppConfig", args{"AppConfig"}, "APP_CONFIG"},
		{"Test OtelCollectorEndpoint", args{"OtelCollectorEndpoint"}, "OTEL_COLLECTOR_ENDPOINT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringToAllCapsCase(tt.args.str); got != tt.want {
				t.Errorf("stringToAllCapsCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
