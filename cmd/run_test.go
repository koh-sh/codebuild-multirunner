package cmd

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
)

func Test_convertBuildConfigToStartBuildInput(t *testing.T) {
	type args struct {
		build Build
	}
	tests := []struct {
		name string
		args args
		want codebuild.StartBuildInput
	}{
		{name: "basic",
			args: args{Build{}},
			want: codebuild.StartBuildInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBuildConfigToStartBuildInput(tt.args.build); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertBuildConfigToStartBuildInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
