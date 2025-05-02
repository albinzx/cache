package cache

import (
	"reflect"
	"testing"
	"time"
)

func TestWithTTL(t *testing.T) {
	type args struct {
		ttl time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "test with 10s",
			args: args{ttl: 10 * time.Second},
			want: 10 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setConfig := &SetConfiguration{}
			WithTTL(tt.args.ttl)(setConfig)
			if got := setConfig.TTL; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}
