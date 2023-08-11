package dump

import (
	"testing"

	root "github.com/koh-sh/codebuild-multirunner/cmd"
)

func Test_dumpConfig(t *testing.T) {
	type args struct {
		bc root.BuildConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "basic",
			args: args{root.BuildConfig{Builds: []root.Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
			},
			want: `builds:
    - projectName: testproject
      sourceVersion: chore/test
    - projectName: testproject2
`,
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
