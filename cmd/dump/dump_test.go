package dump

import (
	"testing"

	"github.com/koh-sh/codebuild-multirunner/common"
)

func Test_dumpConfig(t *testing.T) {
	wantyaml := `builds:
    - projectName: testproject
      sourceVersion: chore/test
    - projectName: testproject2
`
	type args struct {
		bc common.BuildConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "basic",
			args: args{common.BuildConfig{Builds: []common.Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
			},
			want: wantyaml,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dumpConfig(tt.args.bc); got != tt.want {
				t.Errorf("dumpConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
