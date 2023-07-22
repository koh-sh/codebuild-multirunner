package cmd

import (
	"reflect"
	"testing"
)

func Test_readConfigFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want BuildConfig
	}{
		{name: "basic",
			args: args{"testfiles/_test.yaml"},
			want: BuildConfig{[]Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
		},
		{name: "environment variable",
			args: args{"testfiles/_test2.yaml"},
			want: BuildConfig{[]Build{
				{ProjectName: "testproject3", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
		},
	}
	t.Setenv("TEST_ENV", "testproject3") // setting environment variable for test case 2
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readConfigFile(tt.args.filepath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
