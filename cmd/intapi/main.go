package main

import (
	"log"
	"net/http"
	"queue_processor/pkg/helpers"
	"queue_processor/pkg/internal_api"
)

func main() {
	http.HandleFunc("/intapi/v1/form/resize_worker", internal_api.ResizeHandler)
	
	log.Println("starting intapi at 8081")

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		helpers.FatalOnError("intapi err at start ", err)
	}

}
