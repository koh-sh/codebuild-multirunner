package cmd

import "testing"

func Test_dumpConfig(t *testing.T) {
	type args struct {
		bc BuildConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "basic",
			args: args{BuildConfig{[]Build{
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
