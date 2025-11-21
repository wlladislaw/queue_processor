package internal_api

import (
	"encoding/json"
	"log"
	"net/http"
)

func ResizeHandler(w http.ResponseWriter, r *http.Request) {
	task := &ImgResizeTask{}
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "cant unpack json", http.StatusBadRequest)
		return
	}

	log.Printf("intapi handle task: %+v", task)
	errRes := ResizeWorker(task)
	if errRes != nil {
		http.Error(w, "resize internal err", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
