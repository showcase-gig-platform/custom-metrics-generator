package controllers

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type storage struct {
	metrics map[string]*dto.MetricFamily
}

type metric struct {
	name  string
	label map[string]string
	value int
}

var (
	listen            string
	path              string
	flagListenDefault = ":8082"
	flagPathDefault   = "/metrics"
)

func init() {
	flag.StringVar(&listen, "generate-metrics-bind-address", flagListenDefault, "Generated metrics endpoint addr.")
	flag.StringVar(&path, "generate-metrics-path", flagPathDefault, "Generated metrics path.")
}

func NewStorage() *storage {
	return &storage{
		metrics: map[string]*dto.MetricFamily{},
	}
}

func (s *storage) write(k string, m metric) {
	s.metrics[k] = &dto.MetricFamily{
		Name: proto.String(m.name),
		Help: proto.String("auto generateted metrics by " + k),
		Type: dto.MetricType_GAUGE.Enum(),
		Metric: []*dto.Metric{
			{
				Gauge: &dto.Gauge{
					Value: proto.Float64(float64(m.value)),
				},
				Label: genLabel(m.label),
			},
		},
	}
}

func (s *storage) delete(k string) {
	delete(s.metrics, k)
}

func (s *storage) gather() ([]*dto.MetricFamily, error) {
	var result []*dto.MetricFamily
	for _, metrics := range s.metrics {
		result = append(result, metrics)
	}
	return result, nil
}

func (s *storage) updateValue(k string, v int) {
	target := s.metrics[k]
	metric := target.Metric[0]
	metric.Gauge.Value = proto.Float64(float64(v))
}

func (s *storage) serve() {
	g := prometheus.GathererFunc(s.gather)
	http.Handle(path, promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	log.Log.Error(http.ListenAndServe(listen, nil), "Metrics server ended.")
}

func genLabel(source map[string]string) []*dto.LabelPair {
	var result []*dto.LabelPair
	for key, value := range source {
		lp := &dto.LabelPair{
			Name:  proto.String(key),
			Value: proto.String(value),
		}
		result = append(result, lp)
	}
	return result
}
