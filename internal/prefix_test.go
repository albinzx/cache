package internal

import (
	"testing"
)

func TestWithPrefix_Prefix(t *testing.T) {
	type fields struct {
		prefix string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "test with prefix",
			fields: fields{prefix: "test"},
			args:   args{key: "key"},
			want:   "test.key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WithPrefix{
				Name: tt.fields.prefix,
			}
			if got := wp.Prefix(tt.args.key); got != tt.want {
				t.Errorf("WithPrefix.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoPrefix_Prefix(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		np   *NoPrefix
		args args
		want string
	}{
		{
			name: "test without prefix",
			np:   &NoPrefix{},
			args: args{key: "key"},
			want: "key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			np := &NoPrefix{}
			if got := np.Prefix(tt.args.key); got != tt.want {
				t.Errorf("NoPrefix.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
