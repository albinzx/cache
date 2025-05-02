package redis

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/albinzx/cache"
	"github.com/albinzx/cache/internal"
	"github.com/go-redis/redismock/v9"
	goredis "github.com/redis/go-redis/v9"
)

func Test_defaults(t *testing.T) {
	type args struct {
		cache *Cacher
	}
	tests := []struct {
		name          string
		args          args
		wantClientNil bool
		wantNameNil   bool
	}{
		{
			name:          "test with nil client",
			args:          args{cache: &Cacher{client: nil, prefix: &internal.NoPrefix{}}},
			wantClientNil: false,
			wantNameNil:   false,
		},
		{
			name:          "test with nil name",
			args:          args{cache: &Cacher{client: &goredis.Client{}, prefix: nil}},
			wantClientNil: false,
			wantNameNil:   false,
		},
		{
			name:          "test with nil client and name",
			args:          args{cache: &Cacher{client: nil, prefix: nil}},
			wantClientNil: false,
			wantNameNil:   false,
		},
		{
			name:          "test with client and name",
			args:          args{cache: &Cacher{client: &goredis.Client{}, prefix: &internal.NoPrefix{}}},
			wantClientNil: false,
			wantNameNil:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults(tt.args.cache)
			if got := tt.args.cache.client; (got == nil) != tt.wantClientNil {
				t.Errorf("cache.client = %v, want nil %v", got, tt.wantClientNil)
			}
			if got := tt.args.cache.prefix; (got == nil) != tt.wantNameNil {
				t.Errorf("cache.name = %v, want nil %v", got, tt.wantNameNil)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name:    "test with nil options",
			args:    args{options: nil},
			wantNil: false,
		},
		{
			name:    "test with empty options",
			args:    args{options: []Option{}},
			wantNil: false,
		},
		{
			name:    "test with options",
			args:    args{options: []Option{WithTTL(10)}},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.options...); (got == nil) != tt.wantNil {
				t.Errorf("New() = %v, want nil %v", got, tt.wantNil)
			}
		})
	}
}

func TestCacher_Set(t *testing.T) {
	type args struct {
		ctx        context.Context
		key        string
		value      []byte
		setOptions []cache.SetOption
		init       func(string, []byte) (*Cacher, redismock.ClientMock)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test set with no expiration",
			args: args{
				ctx:        context.Background(),
				key:        "key",
				value:      []byte("value"),
				setOptions: []cache.SetOption{},
				init: func(key string, value []byte) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectSet(key, value, 0).SetVal("OK")
					return &Cacher{
						client: client,
						prefix: &internal.NoPrefix{},
					}, mock
				},
			},
			wantErr: false,
		},
		{
			name: "test set with global ttl",
			args: args{
				ctx:        context.Background(),
				key:        "key",
				value:      []byte("value"),
				setOptions: []cache.SetOption{},
				init: func(key string, value []byte) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectSet(key, value, time.Second).SetVal("OK")
					return &Cacher{
						client: client,
						ttl:    time.Second,
						prefix: &internal.NoPrefix{},
					}, mock
				},
			},
			wantErr: false,
		},
		{
			name: "test set with ttl option",
			args: args{
				ctx:        context.Background(),
				key:        "key",
				value:      []byte("value"),
				setOptions: []cache.SetOption{cache.WithTTL(5 * time.Second)},
				init: func(key string, value []byte) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectSet("test."+key, value, 5*time.Second).SetVal("OK")
					return &Cacher{
						client: client,
						prefix: &internal.WithPrefix{Name: "test"},
					}, mock
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, mock := tt.args.init(tt.args.key, tt.args.value)
			if err := c.Set(tt.args.ctx, tt.args.key, tt.args.value, tt.args.setOptions...); (err != nil) != tt.wantErr {
				t.Errorf("Cacher.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Set() expectation were not met, %v", err)
			}
		})
	}
}

func TestCacher_Get(t *testing.T) {
	type args struct {
		ctx  context.Context
		key  string
		init func(string) (*Cacher, redismock.ClientMock)
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test get empty value",
			args: args{
				ctx: context.Background(),
				key: "key",
				init: func(key string) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectGet(key).SetErr(goredis.Nil)
					return &Cacher{
						client: client,
						ttl:    time.Second,
						prefix: &internal.NoPrefix{},
					}, mock
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test get with value",
			args: args{
				ctx: context.Background(),
				key: "key",
				init: func(key string) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectSet(key, []byte("value"), 5*time.Second).SetVal("OK")
					mock.ExpectGet(key).SetVal("value")
					cacher := &Cacher{
						client: client,
						ttl:    5 * time.Second,
						prefix: &internal.NoPrefix{},
					}
					cacher.Set(context.Background(), key, []byte("value"))

					return cacher, mock
				},
			},
			want:    []byte("value"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, mock := tt.args.init(tt.args.key)
			got, err := c.Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cacher.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.Get() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Set() expectation were not met, %v", err)
			}
		})
	}
}

func TestCacher_Delete(t *testing.T) {
	type args struct {
		ctx  context.Context
		key  string
		init func(string) (*Cacher, redismock.ClientMock)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test delete with no error",
			args: args{
				ctx: context.Background(),
				key: "key",
				init: func(key string) (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectDel(key).SetVal(1)
					return &Cacher{
						client: client,
						prefix: &internal.NoPrefix{},
					}, mock
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, mock := tt.args.init(tt.args.key)
			if err := c.Delete(tt.args.ctx, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Cacher.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Set() expectation were not met, %v", err)
			}
		})
	}
}

func TestCacher_Load(t *testing.T) {
	type args struct {
		ctx  context.Context
		data map[string][]byte
		init func() (*Cacher, redismock.ClientMock)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test load with no error",
			args: args{
				ctx: context.Background(),
				data: map[string][]byte{
					"key1": []byte("value1"),
					"key2": []byte("value2"),
				},
				init: func() (*Cacher, redismock.ClientMock) {
					client, mock := redismock.NewClientMock()
					mock.ExpectSet("key1", []byte("value1"), time.Second).SetVal("OK")
					mock.ExpectSet("key2", []byte("value2"), time.Second).SetVal("OK")
					return &Cacher{
						client: client,
						ttl:    time.Second,
						prefix: &internal.NoPrefix{},
					}, mock
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, mock := tt.args.init()
			if err := c.Load(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Cacher.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Set() expectation were not met, %v", err)
			}
		})
	}
}

func TestCacher_Close(t *testing.T) {
	type fields struct {
		client      goredis.UniversalClient
		closeClient bool
	}
	client, _ := redismock.NewClientMock()
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test close with close client true",
			fields: fields{
				client:      client,
				closeClient: true,
			},
			wantErr: false,
		},
		{
			name: "test close with close client false",
			fields: fields{
				client:      client,
				closeClient: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				client:      tt.fields.client,
				closeClient: tt.fields.closeClient,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Cacher.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithRedisClient(t *testing.T) {
	type args struct {
		client goredis.UniversalClient
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name:    "test with client",
			args:    args{client: &goredis.Client{}},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{}
			WithRedisClient(tt.args.client)(c)
			if got := c.client; (got == nil) != tt.wantNil {
				t.Errorf("WithRedisClient() = %v, want nil %v", got, tt.wantNil)
			}
		})
	}
}

func TestWithSharedRedisClient(t *testing.T) {
	type args struct {
		client      goredis.UniversalClient
		closeClient bool
	}
	tests := []struct {
		name      string
		args      args
		wantNil   bool
		wantClose bool
	}{
		{
			name:      "test with client and close client true",
			args:      args{client: &goredis.Client{}, closeClient: true},
			wantNil:   false,
			wantClose: true,
		},
		{
			name:      "test with client and close client false",
			args:      args{client: &goredis.Client{}, closeClient: false},
			wantNil:   false,
			wantClose: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{}
			WithSharedRedisClient(tt.args.client, tt.args.closeClient)(c)
			if got := c.client; (got == nil) != tt.wantNil {
				t.Errorf("WithSharedRedisClient() = %v, want nil %v", got, tt.wantNil)
			}
			if got := c.closeClient; got != tt.wantClose {
				t.Errorf("WithSharedRedisClient() = %v, want close %v", got, tt.wantClose)
			}
		})
	}
}

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
			c := &Cacher{}
			WithTTL(tt.args.ttl)(c)
			if got := c.ttl; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want internal.KeyPrefix
	}{
		{
			name: "test with name",
			args: args{name: "test"},
			want: &internal.WithPrefix{Name: "test"},
		},
		{
			name: "test with empty name",
			args: args{name: ""},
			want: &internal.NoPrefix{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{}
			WithName(tt.args.name)(c)
			if got := c.prefix; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithName() = %v, want %v", got, tt.want)
			}
		})
	}
}
