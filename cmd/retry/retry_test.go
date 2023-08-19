package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/koh-sh/codebuild-multirunner/common"
)

func Test_retryCodeBuild(t *testing.T) {
	id1 := "project:12345678"
	type args struct {
		client func(t *testing.T) common.CodeBuildAPI
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{client: common.ReturnRetryBuildMockAPI(types.Build{Id: &id1}), id: id1},
			want:    id1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retryCodeBuild(tt.args.client(t), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("retryCodeBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("retryCodeBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}
