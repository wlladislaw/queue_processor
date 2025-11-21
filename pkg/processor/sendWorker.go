package processor

import (
	"bytes"
	"log"
	"net/http"
	"queue_processor/pkg/metrics"
	"time"
)

type SendWorker struct {
	URL           string
	MaxRetryAtts  int
	InitTimeRetry int
	Metrics       *metrics.Metrics
	Tube          Tuber
}

func NewSendWorker(url string, maxRetryAtts int, initretry int,
	metrics *metrics.Metrics, tube Tuber) *SendWorker {
	return &SendWorker{URL: url, MaxRetryAtts: maxRetryAtts,
		InitTimeRetry: initretry, Metrics: metrics, Tube: tube}
}

func (sw *SendWorker) Send(task Task) {
	var delay = sw.InitTimeRetry
	for i := 0; ; i++ {
		msg := bytes.NewReader(task.Body())

		timeStart := time.Now()
		sw.Metrics.TakingTaskDur.
			WithLabelValues(sw.Tube.GetTubeName()).Observe(time.Since(task.GetCreatedAt()).Seconds())

		res, err := http.Post(sw.URL, "application/json", msg)
		if err != nil {
			log.Println("err at post request to intapi", err)
		}

		taskDur := time.Since(timeStart)
		defer func() {
			if closeErr := res.Body.Close(); closeErr != nil {
				log.Printf("err close res body at defer: %v\n", closeErr)
			}
		}()

		sw.Metrics.TasksByStatus.WithLabelValues(sw.Tube.GetTubeName(),res.Status).Inc()
		sw.Metrics.TaskDuration.WithLabelValues(sw.Tube.GetTubeName(),res.Status).Observe(taskDur.Seconds())

		switch res.StatusCode {
		case 200:
			log.Printf("worker did task: %v", task)
			if err := task.Ack(false); err != nil {
				log.Printf("err Ack: %s at task %s", err, task)
			}
			return
		case 400:
			log.Printf("bad task req: %v", task)
			if err := task.Ack(false); err != nil {
				log.Printf("err Ack: %s at task %s", err, task)
			}
			return
		case 429:
			time.Sleep(10 * time.Second)
			sw.Send(task)
		case 500:
			log.Println("Server err 500, trying again ")

			if i >= (sw.MaxRetryAtts - 1) {
				log.Printf("after %d attempts, last error: %s", sw.MaxRetryAtts, err)
				errNack := task.Nack(false, false)
				if errNack != nil {
					log.Printf("err Nack %s , task - %v", errNack, task)
				}
				return
			}

			time.Sleep(time.Second * time.Duration(delay))
			delay = delay * 2
		}
	}
}
