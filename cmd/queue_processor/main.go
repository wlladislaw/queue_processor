package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"queue_processor/pkg/helpers"
	"queue_processor/pkg/metrics"
	"queue_processor/pkg/processor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	conf, errConf := helpers.LoadConf("./configs/config.yaml")
	helpers.FatalOnError("err config read ", errConf)

	queueName := conf.TubeName
	tubeAddr := conf.URL
	kafkaT, err := processor.NewKafka([]string{tubeAddr}, conf.MaxRetryAtts)
	//rabitQueue, err := processor.NewRabbitMQ(amqpAddr)
	helpers.FatalOnError("cant start processor ", err)

	defer func() {
		if closeErr := kafkaT.Close(); closeErr != nil {
			log.Printf("err at defer closer chan processor: %v\n", closeErr)
		}
	}()

	taskDest, ok := conf.Tasks[queueName]
	if !ok {
		log.Fatal("url for task intapi not found")
	}

	tube, err := kafkaT.Consume(queueName)
	helpers.FatalOnError("cant register consumer", err)

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	worker := processor.NewSendWorker(taskDest, conf.MaxRetryAtts, conf.InitialTimeRetry, m, tube)
	go func() {
		processor.Pool(tube.GetTasks(), conf.WorkersCount, worker)
	}()
	go func() {
		for{
			m.MsgInQueue.WithLabelValues(tube.GetTubeName()).Set(float64(tube.GetMesgsCount()))
			fmt.Println("metric[messges in tube now]: ", tube.GetMesgsCount())
			time.Sleep(1 * time.Minute)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	fmt.Printf("starting processor at :%s, metrics at /metrics\n", conf.Port)
	errServer := http.ListenAndServe(":"+conf.Port, nil)
	if errServer != nil {
		helpers.FatalOnError("init err processor ", errServer)
	}
}

