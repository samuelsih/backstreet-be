package repo

import (
	"backstreetlinkv2/api/model"
	"context"
	"testing"
)

func TestPGRepo_Insert(t *testing.T) {
	p := PGRepo{testDB}
	ctx := context.Background()

	type args struct {
		ctx        context.Context
		key        string
		dataSource any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert success",
			args: args{
				ctx: ctx,
				key: "some-key",
				dataSource: model.ShortenRequest{
					Alias:      "some-key",
					Type:       "LINK",
					RedirectTo: "https://google.com",
				},
			},
			wantErr: false,
		},

		{
			name: "insert duplicate above",
			args: args{
				ctx: ctx,
				key: "some-key",
				dataSource: model.ShortenRequest{
					Alias:      "some-key",
					Type:       "LINK",
					RedirectTo: "https://google.com",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := p.Insert(tt.args.ctx, tt.args.key, tt.args.dataSource); (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
