package dump

import (
	"testing"

	"github.com/koh-sh/codebuild-multirunner/internal/types"
)

func Test_dumpConfig(t *testing.T) {
	wantyaml := `builds:
    - projectName: testproject
      sourceVersion: chore/test
    - projectName: testproject2
`
	type args struct {
		bc types.BuildConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				types.BuildConfig{
					Builds: []types.Build{
						{ProjectName: "testproject", SourceVersion: "chore/test"},
						{ProjectName: "testproject2"},
					},
				},
			},
			want:    wantyaml,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dumpConfig(tt.args.bc)
			if (err != nil) != tt.wantErr {
				t.Errorf("dumpConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("dumpConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
