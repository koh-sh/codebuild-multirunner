package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
)

func Test_retryCodeBuild(t *testing.T) {
	id1 := "project:12345678"
	type args struct {
		client func(t *testing.T) mr.CodeBuildAPI
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
			args:    args{client: mr.ReturnRetryBuildMockAPI(types.Build{Id: &id1}), id: id1},
			want:    id1,
			wantErr: false,
		},
		{
			name:    "api error",
			args:    args{client: mr.ReturnRetryBuildMockAPI(types.Build{}), id: "error"},
			want:    "",
			wantErr: true,
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
