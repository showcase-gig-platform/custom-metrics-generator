/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"strings"
	"time"

	k8sv1 "github.com/showcase-gig-platform/custom-metrics-generator/api/v1"

	"github.com/showcase-gig-platform/cron/v3"
)

// MetricsSourceReconciler reconciles a MetricsSource object
type MetricsSourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// 定期更新やunregisterする時のためにグローバルで持っておく
var collectors = map[string]*prometheus.GaugeVec{}

var (
	interval            int
	offset              int
	timezone            string
	prefix              string
	flagIntervalDefault = 60
	flagOffsetDefault   = 0
	flagTimezoneDefault = "UTC"
	flagPrefixDefault   = ""
)

func init() {
	flag.IntVar(&interval, "interval-seconds", flagIntervalDefault, "interval seconds to fetch metrics")
	flag.IntVar(&offset, "offset-seconds", flagOffsetDefault, "offset seconds to generate metrics")
	flag.StringVar(&timezone, "timezone", flagTimezoneDefault, "set timezone")
	flag.StringVar(&prefix, "metrics-prefix", flagPrefixDefault, "set prefix for metrics name")
}

func initGaugeVec(name string, labelKeys []string, constLabels map[string]string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        name,
			Help:        "auto generateted metrics for " + name,
			ConstLabels: constLabels,
		},
		labelKeys,
	)
}

func parse(cs string) (cron.Schedule, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if c, e := p.Parse(cs); e != nil {
		return nil, e
	} else {
		return c, nil
	}
}

func resumeNamespacedName(namespacednamestring string) (types.NamespacedName, error) {
	split := strings.Split(namespacednamestring, "/")
	if len(split) != 2 {
		return types.NamespacedName{}, apierrors.NewBadRequest("cannot generate NamespacedName from string")
	}

	return types.NamespacedName{
		Namespace: split[0],
		Name:      split[1],
	}, nil
}

//+kubebuilder:rbac:groups=k8s.oder.com,resources=metricssources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.oder.com,resources=metricssources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.oder.com,resources=metricssources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MetricsSource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MetricsSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// ctrl.Requestにはリソースのnamespaceとnameが入ってるだけで
	// そのリソースに対するアクションがcreate/edit/deleteどれなのかの情報はない
	// deleteはリソースをGetしてnotFoundなら削除されたと判断する

	log.Log.Info("triggered reconcile", "request", req)

	_ = log.FromContext(ctx)

	key := req.String()

	var resource k8sv1.MetricsSource
	if e := r.Get(ctx, req.NamespacedName, &resource); e != nil {
		if apierrors.IsNotFound(e) {
			// リソースが削除された場合の処理
			deleteMetrics(key)
			return ctrl.Result{}, nil
		} else {
			// それ以外のエラー
			return ctrl.Result{}, fmt.Errorf("reconcile - failed to get resource : %w", e)
		}
	}

	// cron形式が正しいかチェック
	// 他のチェックも入れて総合的なvalidateで切り出しても
	for _, item := range resource.Spec.Metrics {
		cs := item.Start
		if _, e := parse(cs); e != nil {
			return ctrl.Result{}, fmt.Errorf("reconcile - failed to parse cron : %w", e)
		}
	}

	status := generateStatus(resource.Spec.Metrics, time.Now())
	resource.Status = status
	r.Status().Update(ctx, &resource)

	metricsName := convertPromFormatName(prefix + resource.Spec.MetricsName)
	labels := formatAllLabels(resource.Spec.Labels)
	var keys []string
	for k, _ := range labels {
		keys = append(keys, k)
	}
	gauge := initGaugeVec(metricsName, keys, prometheus.Labels{"origin": key})
	// レジストリからいったん削除してから再登録する
	// metricsSource変更時にメトリクス名やlabelのkeyに変更があった場合の古いメトリクスが残るのを防ぎたいが
	// 差分あるなしをちゃんとハンドリングしようとすると結構面倒だったので一律で削除してから登録としておく
	deleteMetrics(key)
	if e := metrics.Registry.Register(gauge); e != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile - failed to register metrics : %w", e)
	}
	gauge.With(labels).Set(float64(status.CurrentValue))
	collectors[key] = gauge

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetricsSourceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	log.Log.Info("setupwithmanager")

	// TODO: エラーハンドリング
	go func() {
		for {
			log.Log.Info("periodic update start")

			ls, _ := metrics.Registry.Gather()
			for _, l := range ls {
				if *l.Name == "sample_metrics" {
					log.Log.Info("debug", "gather", l)
				}
			}

			r.updateAllStatusAndMetrics(ctx)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1.MetricsSource{}).
		Complete(r)
}

func (r *MetricsSourceReconciler) updateAllStatusAndMetrics(ctx context.Context) {
	for key, collector := range collectors {
		nn, err := resumeNamespacedName(key)
		if err != nil {
			log.Log.Error(err, "failed to resume namespaced-name.")
			continue
		}

		var resource k8sv1.MetricsSource
		if e := r.Get(ctx, nn, &resource); e != nil {
			// これがよく出るようだとcollectorsの中身と登録済みresourceが何らかの原因でずれている可能性が
			log.Log.Error(err, fmt.Sprintf("failed to get resource : %s", nn.String()))
			continue
		}

		o := getOffset(resource.Spec.OffsetSeconds)
		l := getLocation(resource.Spec.Timezone)
		refTime := time.Now().In(l).Add(o)
		status := generateStatus(resource.Spec.Metrics, refTime)
		resource.Status = status
		r.Status().Update(ctx, &resource)

		labels := formatAllLabels(resource.Spec.Labels)
		collector.With(labels).Set(float64(status.CurrentValue))
	}
}

// prometheusのメトリクス名とlabel名に使用できる文字列に変換
// [a-zA-Z_][a-zA-Z0-9_]*
// 不正な文字種は _ に置換

const replace = "_"

func convertPromFormatName(str string) string {
	first := regexp.MustCompile(`^[^a-zA-Z_:]`)
	other := regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	return other.ReplaceAllString(first.ReplaceAllString(str, replace), replace)
}

func convertPromFormatLabelKey(str string) string {
	first := regexp.MustCompile(`^[^a-zA-Z_]`)
	other := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	finish := regexp.MustCompile(`^_{2,}`)
	return finish.ReplaceAllString(other.ReplaceAllString(first.ReplaceAllString(str, replace), replace), replace)
}

func formatAllLabels(labels map[string]string) map[string]string {
	var ret = map[string]string{}
	for k, v := range labels {
		key := convertPromFormatLabelKey(k)
		ret[key] = v
	}
	return ret
}

func getLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(timezone)
	if tz != "" {
		loc, err = time.LoadLocation(tz)
	}
	if err != nil {
		log.Log.Error(err, "failed to load timezone, use UTC.")
		loc = time.FixedZone("UTC", 0)
	}
	return loc
}

func getOffset(o *int) time.Duration {
	if o != nil {
		return time.Duration(*o) * time.Second
	}
	return time.Duration(offset) * time.Second
}

func deleteMetrics(key string) {
	if old, ok := collectors[key]; ok {
		if b := metrics.Registry.Unregister(old); b {
			// collectorsからdeleteするのはUnregisterの成否を問わないほうがいいかも？
			// 何かしらの理由でレジストリとcollectorsに差分が発生する可能性と、Unregisterが失敗する原因パターン次第
			delete(collectors, key)
		} else {
			log.Log.Info("deleteMetrics: could not remove target metrics")
		}
	}
}
