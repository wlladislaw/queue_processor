package main

import (
	"fmt"
	"net/http"
	"queue_processor/pkg/client"
	"queue_processor/pkg/helpers"
	"queue_processor/pkg/processor"
)

func main() {
	conf, errConf := helpers.LoadConf("./configs/config.yaml")
	helpers.FatalOnError("config read err ", errConf)

	//rabbitQ, err := processor.NewRabbitMQ(conf.URL)
	kafkaT, err := processor.NewKafka([]string{conf.URL}, conf.MaxRetryAtts)
	helpers.FatalOnError("cant connect to tube processor", err)

	h := &client.ClientHandler{
		// Qp: rabbitQ,
		Qp: kafkaT,
	}

	http.HandleFunc("/", h.MainPage)
	http.HandleFunc("/upload", h.UploadPage)

	fmt.Println("starting client server at :8089")

	errServer := http.ListenAndServe(":8089", nil)
	if errServer != nil {
		helpers.FatalOnError("init err clientapi ", errServer)
	}
}
