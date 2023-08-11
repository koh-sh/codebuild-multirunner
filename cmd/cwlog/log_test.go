package cwlog

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
)

func Test_getCloudWatchLogSetting(t *testing.T) {
	var id = "project:12345678"
	var group = "/aws/codebuild/project"
	var stream = "12345678"
	var enabledClient = root.ReturnBatchGetBuildsMockAPI([]types.Build{
		{Logs: &types.LogsLocation{CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "ENABLED"},
			GroupName:  &group,
			StreamName: &stream},
		},
	})
	var disabledClient = root.ReturnBatchGetBuildsMockAPI([]types.Build{
		{Logs: &types.LogsLocation{CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "DISABLED"},
			GroupName:  &group,
			StreamName: &stream},
		},
	})
	type args struct {
		client root.CodeBuildAPI
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{name: "enabled",
			args:    args{enabledClient(t), id},
			want:    "/aws/codebuild/project",
			want1:   "12345678",
			wantErr: false,
		},
		{name: "disabled",
			args:    args{disabledClient(t), id},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getCloudWatchLogSetting(tt.args.client, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCloudWatchLogSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCloudWatchLogSetting() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCloudWatchLogSetting() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getCloudWatchLogEvents(t *testing.T) {
	var lines = []string{"first line", "second line", "third line"}
	var wantOutput = []cwltypes.OutputLogEvent{
		{
			IngestionTime: new(int64),
			Message:       &lines[0],
			Timestamp:     new(int64),
		},
		{
			IngestionTime: new(int64),
			Message:       &lines[1],
			Timestamp:     new(int64),
		},
		{
			IngestionTime: new(int64),
			Message:       &lines[2],
			Timestamp:     new(int64),
		},
	}
	var successClient = root.ReturnGetLogEventsMockAPI(wantOutput)
	var failClient = root.ReturnGetLogEventsMockAPI([]cwltypes.OutputLogEvent{})
	type args struct {
		client root.CWLGetLogEventsAPI
		group  string
		stream string
	}
	tests := []struct {
		name    string
		args    args
		want    cloudwatchlogs.GetLogEventsOutput
		wantErr bool
	}{
		{name: "success",
			args: args{successClient(t), "/aws/codebuild/project", "12345678"},
			want: cloudwatchlogs.GetLogEventsOutput{
				Events:            wantOutput,
				NextBackwardToken: nil,
				NextForwardToken:  nil,
			},
			wantErr: false},
		{name: "fail",
			args:    args{failClient(t), "", ""},
			want:    cloudwatchlogs.GetLogEventsOutput{},
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCloudWatchLogEvents(tt.args.client, tt.args.group, tt.args.stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCloudWatchLogEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCloudWatchLogEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}
