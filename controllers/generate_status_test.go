package controllers

import (
	k8sv1 "github.com/showcase-gig-platform/custom-metrics-generator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
	"time"
)

func Test_generateStatus(t *testing.T) {
	type args struct {
		metrics []k8sv1.MetricsSourceSpecMetric
		now     time.Time
	}
	tests := []struct {
		name string
		args args
		want k8sv1.MetricsSourceStatus
	}{
		{
			name: "1 metric 1",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 11, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "1 metric 2",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "1 metric 3",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "1 metric 4",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "1 metric 5",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 13, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 1 - 1",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("40m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("20m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 11, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 13, 30, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 1 - 2",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("40m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("20m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 12, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 40, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "2 metrics 1 - 3",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("40m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("20m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 12, 50, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 40, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "2 metrics 1 - 4",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("40m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("20m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 20, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 30, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "2 metrics 1 - 5",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("40m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("20m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 30, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 2 - 1",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 11, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 15, 10, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 2 - 2",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 12, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "2 metrics 2 - 3",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 14, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 15, 10, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "2 metrics 2 - 4",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 15, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 15, 10, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 3 - 1",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 11, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 14, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 3 - 2",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "2 metrics 3 - 3",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "2 metrics 3 - 4",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 3 - 5",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 10, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 3 - 6",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "2 metrics 3 - 7",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 13, 50, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "2 metrics 3 - 8",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "2 metrics 3 - 9",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "10 13 * * *",
						Duration: metav1.Duration{Duration: duration("30m")},
						Value:    5,
					},
				},
				now: time.Date(2022, 1, 5, 15, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "3 metrics 1 - 1",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 11, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 15, 20, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "3 metrics 1 - 2",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 20, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "3 metrics 1 - 3",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 13, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 20, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    20,
				},
			},
		},
		{
			name: "3 metrics 1 - 4",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 13, 50, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 20,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    20,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 40, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "3 metrics 1 - 5",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 14, 10, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 20,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 40, 0, 0, time.UTC)},
					Value:    20,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 40, 0, 0, time.UTC)},
					Value:    5,
				},
			},
		},
		{
			name: "3 metrics 1 - 6",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 14, 50, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 5,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 14, 40, 0, 0, time.UTC)},
					Value:    5,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 15, 20, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "3 metrics 1 - 7",
			args: args{
				metrics: []k8sv1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    10,
					},
					{
						Start:    "20 13 * * *",
						Duration: metav1.Duration{Duration: duration("120m")},
						Value:    5,
					},
					{
						Start:    "40 13 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    20,
					},
				},
				now: time.Date(2022, 1, 5, 15, 30, 0, 0, time.UTC),
			},
			want: k8sv1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 15, 20, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: k8sv1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateStatus(tt.args.metrics, tt.args.now)
			if !reflect.DeepEqual(got.CurrentValue, tt.want.CurrentValue) {
				t.Errorf("CurrentValue = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Last, tt.want.Last) {
				t.Errorf("LastSchedule = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Next, tt.want.Next) {
				t.Errorf("NextSchedule = %v, want %v", got, tt.want)
			}
		})
	}
}
