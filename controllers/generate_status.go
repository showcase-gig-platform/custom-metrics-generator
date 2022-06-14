package controllers

import (
	"fmt"
	k8sv1 "github.com/showcase-gig-platform/custom-metrics-generator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type metricTime struct {
	metric k8sv1.MetricsSourceSpecMetric
	time   time.Time
}

func generateStatus(metrics []k8sv1.MetricsSourceSpecMetric, refTime time.Time) k8sv1.MetricsSourceStatus {
	currentMetric := getMetricSpecificTime(metrics, refTime)
	// 該当するmetricがなかった場合は空の構造体が返ってくる
	prevEventTime := prevValidSchedule(metrics, currentMetric, refTime)
	nextEventTime := nextSchedule(metrics, currentMetric, refTime)
	nextMetric := getMetricSpecificTime(metrics, nextEventTime)

	return k8sv1.MetricsSourceStatus{
		CurrentValue: currentMetric.Value,
		Last: k8sv1.MetricsSourceStatusSchedule{
			Schedule: metav1.Time{Time: prevEventTime},
			Value:    currentMetric.Value,
		},
		Next: k8sv1.MetricsSourceStatusSchedule{
			Schedule: metav1.Time{Time: nextEventTime},
			Value:    nextMetric.Value,
		},
		LastRefreshTime: metav1.Time{Time: time.Now()},
	}
}

// nowの時点で参照するmetricを選ぶ
// 有効なものが複数ある場合、開始時刻がより近いもの
func getMetricSpecificTime(metrics []k8sv1.MetricsSourceSpecMetric, now time.Time) k8sv1.MetricsSourceSpecMetric {
	var current metricTime
	for _, m := range metrics {
		schedule, e := parse(m.Start)
		if e != nil {
			log.Log.Error(e, fmt.Sprintf("getMetricSpecificTime : Cron parse error, `%v`", m.Start))
			continue
		}
		start := schedule.Prev(now)
		end := start.Add(m.Duration.Duration)
		if end.Before(now) || end == now {
			// 前回のスケジュールがされてからdurationが既に経過している = メトリクスを出す時間範囲にないので無視
			// イコールを含めるのはスケジュール開始時と動作を統一させるため（ただしns単位の話なので実質テストの対応）
			continue
		}
		if start.After(current.time) {
			// 現在のものより近いので採用
			current.metric = m
			current.time = start
		}
	}
	return current.metric
}

// 最後に起きたイベントの時刻を出す
// 現時刻で有効なmetricがない場合（baseMetricsが空の場合）、すべてのmetricsから最後のイベントが起きた時刻
// そうでない場合、baseMetricsの開始時刻または。それより後に開始・終了の両方があったmetricの中で最後のイベントが起きた時刻
func prevValidSchedule(metrics []k8sv1.MetricsSourceSpecMetric, baseMetric k8sv1.MetricsSourceSpecMetric, now time.Time) time.Time {
	var nearly time.Time
	baseTime := time.Time{}
	if bs, e := parse(baseMetric.Start); e == nil {
		baseTime = bs.Prev(now)
		nearly = baseTime
	}
	for _, metric := range metrics {
		ps, e := parse(metric.Start)
		if e != nil {
			log.Log.Error(e, fmt.Sprintf("prevValidSchedule : Cron parse error, `%v`", metric.Start))
			continue
		}
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

// 次回のイベント予定時刻を出す
// 現時刻で有効なmetricがない場合（baseMetricsが空の場合）、すべてのmetricsの開始時刻のうち一番近いもの
// そうでない場合、baseMetricsの終了時刻または、それより前に開始があるmetricsの中で最初にイベントが起きる時刻
func nextSchedule(metrics []k8sv1.MetricsSourceSpecMetric, baseMetric k8sv1.MetricsSourceSpecMetric, now time.Time) time.Time {
	var baseTime time.Time
	var nearly time.Time
	if bs, e := parse(baseMetric.Start); e == nil {
		baseTime = bs.Prev(now).Add(baseMetric.Duration.Duration)
		nearly = baseTime
	}
	for _, metric := range metrics {
		ns, e := parse(metric.Start)
		if e != nil {
			log.Log.Error(e, fmt.Sprintf("nextSchedule : Cron parse error, `%v`", metric.Start))
			continue
		}
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
