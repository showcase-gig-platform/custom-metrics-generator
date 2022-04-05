package controllers

import (
	"flag"
	v1 "github.com/showcase-gig-platform/custom-metrics-generator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
	"time"
)

var (
	zero = 0
	ten  = 10
)

func intPtr(val int) *int {
	return &val
}

func Test_getLocationSpec(t *testing.T) {
	type args struct {
		spec v1.MetricsSourceSpec
	}
	tests := []struct {
		name string
		args args
		want *time.Location
	}{
		{
			"empty spec",
			args{
				v1.MetricsSourceSpec{},
			},
			time.UTC,
		},
		{
			"with spec",
			args{
				v1.MetricsSourceSpec{
					Timezone: "Asia/Tokyo",
				},
			},
			jst,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine.Set("timezone", "UTC")
			if got := getLocation(tt.args.spec.Timezone); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLocationFlag(t *testing.T) {
	type args struct {
		spec v1.MetricsSourceSpec
	}
	tests := []struct {
		name string
		args args
		want *time.Location
	}{
		{
			"empty spec",
			args{
				v1.MetricsSourceSpec{},
			},
			jst,
		},
		{
			"with spec",
			args{
				v1.MetricsSourceSpec{
					Timezone: "UTC",
				},
			},
			time.UTC,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine.Set("timezone", "Asia/Tokyo")
			if got := getLocation(tt.args.spec.Timezone); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOffsetSpec(t *testing.T) {
	type args struct {
		spec v1.MetricsSourceSpec
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			"empty spec",
			args{
				v1.MetricsSourceSpec{},
			},
			0 * time.Second,
		},
		{
			"with spec",
			args{
				v1.MetricsSourceSpec{
					OffsetSeconds: intPtr(10),
				},
			},
			10 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOffset(tt.args.spec.OffsetSeconds); got != tt.want {
				t.Errorf("getOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOffsetFlag(t *testing.T) {
	type args struct {
		spec v1.MetricsSourceSpec
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			"empty spec",
			args{
				v1.MetricsSourceSpec{},
			},
			30 * time.Second,
		},
		{
			"with spec",
			args{
				v1.MetricsSourceSpec{
					OffsetSeconds: intPtr(10),
				},
			},
			10 * time.Second,
		},
		{
			"with spec zero",
			args{
				v1.MetricsSourceSpec{
					OffsetSeconds: intPtr(0),
				},
			},
			0 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine.Set("offset-seconds", "30")
			if got := getOffset(tt.args.spec.OffsetSeconds); got != tt.want {
				t.Errorf("getOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertPromFormat(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "replace 1",
			arg:  "aaaaa",
			want: "aaaaa",
		},
		{
			name: "replace 2",
			arg:  "00000",
			want: "_0000",
		},
		{
			name: "replace 3",
			arg:  "*****",
			want: "_____",
		},
		{
			name: "replace 4",
			arg:  "a*a*a",
			want: "a_a_a",
		},
		{
			name: "replace 5",
			arg:  "0*a0*a",
			want: "__a0_a",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := convertPromFormat(test.arg); !reflect.DeepEqual(got, test.want) {
				t.Errorf("convertPromFormat() = %v, want %v", got, test.want)
			}
		})
	}
}

func Test_offset(t *testing.T) {
	type args struct {
		flagOffset string
		specOffset *int
		metrics    []v1.MetricsSourceSpecMetric
		now        time.Time
	}
	tests := []struct {
		name string
		args args
		want v1.MetricsSourceStatus
	}{
		{
			name: "before 1",
			args: args{
				flagOffset: "600",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 11, 49, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 4, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "before 2",
			args: args{
				flagOffset: "600",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 11, 50, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "middle 1",
			args: args{
				flagOffset: "600",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 49, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "middle 2",
			args: args{
				flagOffset: "600",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 50, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "after 1",
			args: args{
				flagOffset: "600",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("60m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 13, 1, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 0,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 13, 0, 0, 0, time.UTC)},
					Value:    0,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 6, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
			},
		},
		{
			name: "spec",
			args: args{
				flagOffset: "0",
				specOffset: intPtr(600),
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("10m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 11, 55, 0, 0, time.UTC),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 10, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.flagOffset != "" {
				flag.CommandLine.Set("offset-seconds", tt.args.flagOffset)
			}
			now := tt.args.now.Add(getOffset(tt.args.specOffset))
			if got := generateStatus(tt.args.metrics, now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timezone(t *testing.T) {
	type args struct {
		flagTimezone string
		specTimezone string
		metrics      []v1.MetricsSourceSpecMetric
		now          time.Time
	}
	tests := []struct {
		name string
		args args
		want v1.MetricsSourceStatus
	}{
		{
			name: "default",
			args: args{
				flagTimezone: "",
				specTimezone: "",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 3 * * *",
						Duration: metav1.Duration{Duration: duration("10m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 5, 0, 0, jst),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 3, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 3, 10, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "jst to utc",
			args: args{
				flagTimezone: "Asia/Tokyo",
				specTimezone: "UTC",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 3 * * *",
						Duration: metav1.Duration{Duration: duration("10m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 5, 0, 0, jst),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 3, 0, 0, 0, time.UTC)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 3, 10, 0, 0, time.UTC)},
					Value:    0,
				},
			},
		},
		{
			name: "jst",
			args: args{
				flagTimezone: "",
				specTimezone: "Asia/Tokyo",
				metrics: []v1.MetricsSourceSpecMetric{
					{
						Start:    "0 12 * * *",
						Duration: metav1.Duration{Duration: duration("10m")},
						Value:    10,
					},
				},
				now: time.Date(2022, 1, 5, 12, 5, 0, 0, jst),
			},
			want: v1.MetricsSourceStatus{
				CurrentValue: 10,
				Last: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 0, 0, 0, jst)},
					Value:    10,
				},
				Next: v1.MetricsSourceStatusSchedule{
					Schedule: metav1.Time{Time: time.Date(2022, 1, 5, 12, 10, 0, 0, jst)},
					Value:    0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine.Set("timezone", "UTC")
			if tt.args.flagTimezone != "" {
				flag.CommandLine.Set("timezone", tt.args.flagTimezone)
			}
			now := tt.args.now.In(getLocation(tt.args.specTimezone))
			if got := generateStatus(tt.args.metrics, now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
