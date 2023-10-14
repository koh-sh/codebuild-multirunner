package dump

import (
	"testing"

	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
)

func Test_dumpConfig(t *testing.T) {
	wantyaml := `builds:
    - projectName: testproject
      sourceVersion: chore/test
    - projectName: testproject2
`
	type args struct {
		bc mr.BuildConfig
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
				mr.BuildConfig{
					Builds: []mr.Build{
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
