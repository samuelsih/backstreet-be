package repo

import (
	"reflect"
	"testing"
)

func TestCache_Set(t *testing.T) {
	t.Parallel()

	type args struct {
		key   string
		value []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test set",
			args: args{
				key:   "foo",
				value: []byte("bar"),
			},
			wantErr: false,
		},

		{
			name: "test duplicate",
			args: args{
				key:   "foo",
				value: []byte("baz"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testCache.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	t.Parallel()

	if err := testCache.Set("somekey", []byte("test")); err != nil {
		t.Errorf("Get() error = %v", err)
		return
	}

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				key: "somekey",
			},
			want:    []byte("test"),
			wantErr: false,
		},

		{
			name: "test failed",
			args: args{
				key: "somekey123",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testCache.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
