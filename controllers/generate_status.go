package controllers

import (
	k8sv1 "github.com/showcase-gig-platform/custom-metrics-generator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type metricTime struct {
	metric k8sv1.MetricsSourceSpecMetric
	time   time.Time
}

func generateStatus(metrics []k8sv1.MetricsSourceSpecMetric, now time.Time) k8sv1.MetricsSourceStatus {
	currentMetric := getMetricSpecificTime(metrics, now)
	// 該当するmetricがなかった場合は空の構造体が返ってくる
	prevEventTime := prevValidSchedule(metrics, currentMetric, now)
	nextEventTime := nextSchedule(metrics, currentMetric, now)
	next := getMetricSpecificTime(metrics, nextEventTime.Add(1*time.Second))

	return k8sv1.MetricsSourceStatus{
		CurrentValue: currentMetric.Value,
		Last: k8sv1.MetricsSourceStatusSchedule{
			Schedule: metav1.Time{Time: prevEventTime},
			Value:    currentMetric.Value,
		},
		Next: k8sv1.MetricsSourceStatusSchedule{
			Schedule: metav1.Time{Time: nextEventTime},
			Value:    next.Value,
		},
	}
}

// nowになりきって今の値を
func getMetricSpecificTime(metrics []k8sv1.MetricsSourceSpecMetric, now time.Time) k8sv1.MetricsSourceSpecMetric {
	var current metricTime
	for _, m := range metrics {
		schedule, _ := parse(m.Start)
		s := schedule.Prev(now)
		e := s.Add(m.Duration.Duration)
		if e.Before(now) {
			// log.Log.Info("out of range")
			continue
		}
		if s.After(current.time) {
			// log.Log.Info("after")
			current.metric = m
			current.time = s
		} else {
			// log.Log.Info("before")
		}
	}
	return current.metric
}

// baseScheduleより後にscheduleがあったmetricsのうち、一番近いendを出す
func prevValidSchedule(metrics []k8sv1.MetricsSourceSpecMetric, baseMetric k8sv1.MetricsSourceSpecMetric, now time.Time) time.Time {
	var nearly time.Time
	baseTime := time.Time{}
	if bs, e := parse(baseMetric.Start); e == nil {
		baseTime = bs.Prev(now)
		nearly = baseTime
	}
	for _, metric := range metrics {
		ps, _ := parse(metric.Start)
		prev := ps.Prev(now)
		if prev.After(baseTime) {
			end := prev.Add(metric.Duration.Duration)
			if end.After(nearly) {
				nearly = end
			}
		}
	}

	return nearly
}

func nextSchedule(metrics []k8sv1.MetricsSourceSpecMetric, baseMetric k8sv1.MetricsSourceSpecMetric, now time.Time) time.Time {
	var baseTime time.Time
	var nearly time.Time
	if bs, e := parse(baseMetric.Start); e == nil {
		baseTime = bs.Prev(now).Add(baseMetric.Duration.Duration)
		nearly = baseTime
	}
	for _, metric := range metrics {
		ns, _ := parse(metric.Start)
		next := ns.Next(now)
		if !baseTime.IsZero() && next.After(baseTime) {
			continue
		}
		if nearly.IsZero() || next.Before(nearly) {
			nearly = next
		}
	}

	return nearly
}
