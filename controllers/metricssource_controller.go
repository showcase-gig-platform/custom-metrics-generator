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
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
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

func initGaugeVec(name string, labels map[string]string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        name,
			Help:        "auto generateted metrics for " + name,
			ConstLabels: labels,
		},
		[]string{},
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
		return types.NamespacedName{}, errors.NewBadRequest("cannot generate NamespacedName from string")
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
	log.Log.Info("triggered reconcile", "request", req)

	_ = log.FromContext(ctx)

	key := req.String()

	var resource k8sv1.MetricsSource
	deleted := false
	if e := r.Get(ctx, req.NamespacedName, &resource); e != nil {
		if errors.IsNotFound(e) {
			// リソースが削除された場合の処理
			deleted = true
		} else {
			// それ以外のエラー
		}
	}

	// cron形式が正しいかチェック
	// 他のも入れてvalidateで切り出しても
	// log.Log.Info("result", "resources", resources)
	for _, item := range resource.Spec.Metrics {
		// log.Log.Info(item.Duration.String())
		cs := item.Start
		// log.Log.Info(cs)
		_, e := parse(cs)
		if e != nil {
			// log.Log.Error(e, "cron parse error")
			return ctrl.Result{}, e
		} else {
			// log.Log.Info(c.Prev(time.Now()).String())
		}
	}

	status := generateStatus(resource.Spec.Metrics, time.Now())
	resource.Status = status
	// log.Log.Info("resource", "status", fmt.Sprintf("%#v\n", resource.Status))
	r.Status().Update(ctx, &resource)

	// リクエストのresourceのメトリクスをレジストリからいったん削除
	if old, ok := collectors[key]; ok {
		if b := metrics.Registry.Unregister(old); b {
			// log.Log.Info("metrics unregisted")
			delete(collectors, resource.ObjectMeta.Name)
		} else {
			// log.Log.Info("metrics not exist")
		}
	}

	if deleted {
		// log.Log.Info("finish delete")
		// log.Log.Info("--------------------")
		return ctrl.Result{}, nil
	}

	metricsName := convertPromFormat(prefix + resource.Spec.MetricsName)
	var labels = map[string]string{}
	for k, v := range resource.Spec.Labels {
		key := convertPromFormat(k)
		labels[key] = v
	}
	gauge := initGaugeVec(metricsName, labels)
	if e := metrics.Registry.Register(gauge); e != nil {
		return ctrl.Result{}, e
	}
	gauge.With(prometheus.Labels{}).Set(float64(status.CurrentValue))
	// 定期更新やunregisterする時のためにグローバルで持っておく
	collectors[key] = gauge

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetricsSourceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	log.Log.Info("setupwithmanager")

	// TODO: エラーハンドリング
	go func() {
		for {
			r.updateAllStatusAndMetrics(ctx)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1.MetricsSource{}).
		Complete(r)
}

func (r *MetricsSourceReconciler) updateAllStatusAndMetrics(ctx context.Context) {
	log.Log.Info("periodic update start")

	for key, collector := range collectors {
		nn, err := resumeNamespacedName(key)
		if err != nil {
			log.Log.Error(err, "failed to resume namespaced-name.")
			continue
		}

		var resource k8sv1.MetricsSource
		if e := r.Get(ctx, nn, &resource); e != nil {
			// それ以外のエラー
		}

		o := getOffset(resource.Spec.OffsetSeconds)
		l := getLocation(resource.Spec.Timezone)
		refTime := time.Now().In(l).Add(o)
		status := generateStatus(resource.Spec.Metrics, refTime)
		// log.Log.Info("diff", "old", resource.Status, "new", status)
		resource.Status = status
		r.Status().Update(ctx, &resource)

		collector.With(prometheus.Labels{}).Set(float64(status.CurrentValue))
	}
	// log.Log.Info("updateAllStatusAndMetrics end")
}

// prometheusのメトリクス名とlabel名に使用できる文字列に変換
// [a-zA-Z_][a-zA-Z0-9_]*
// 不正な文字種は _ に置換
func convertPromFormat(str string) string {
	first := regexp.MustCompile(`^[^a-zA-Z_]`)
	other := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return other.ReplaceAllString(first.ReplaceAllString(str, "_"), "_")
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
