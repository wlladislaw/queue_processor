package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	MsgInQueue    *prometheus.GaugeVec
	TasksByStatus *prometheus.CounterVec
	TaskDuration  *prometheus.HistogramVec
	TakingTaskDur *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		MsgInQueue: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "messages_in_tube",
			Help: "Number of messages in tube now.",
		},
			[]string{"tube"},
		),
		TasksByStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tasks_by_status",
				Help: "Number of tasks per second by status.",
			},
			[]string{"tube", "status"},
		),
		TaskDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "task_time_duration",
				Help:    "Task processing duration per status",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"tube", "status"},
		),
		TakingTaskDur: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "taking_task_duration_seconds",
				Help:    "Time between taking from tube and processing",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"tube"},
		),
	}
	reg.MustRegister(m.MsgInQueue)
	reg.MustRegister(m.TaskDuration)
	reg.MustRegister(m.TasksByStatus)
	reg.MustRegister(m.TakingTaskDur)
	return m
}
