package client

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"queue_processor/pkg/helpers"
	"queue_processor/pkg/templates"
)

type QueuePublisher interface {
	Publish(qName string, data []byte) error
}

type ClientHandler struct {
	Qp QueuePublisher
}

type ImgResizeTask struct {
	Name string
	MD5  string
}

const (
	ImageResizeQueueName = "image_resize"
)

func (ch *ClientHandler) MainPage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(templates.UploadFormTmpl)
	if err != nil {
		log.Println(err)
		http.Error(w, "cant upload template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ch *ClientHandler) UploadPage(w http.ResponseWriter, r *http.Request) {
	uploadData, handler, err := r.FormFile("my_file")
	if err != nil {
		log.Println(err)
		http.Error(w, "err in form file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if closeErr := uploadData.Close(); closeErr != nil {
			fmt.Printf("err at close file: %s\n", closeErr)
		}
	}()

	_, errFilename := fmt.Fprintf(w, "handler.Filename %v\n", handler.Filename)
	if errFilename != nil {
		log.Println("handler.Filename err ", errFilename)
	}
	_, errHeader := fmt.Fprintf(w, "handler.Header %#v\n", handler.Header)
	if errHeader != nil {
		log.Println("handler.Header err", errHeader)
	}
	tmpName := helpers.RandStringRunes(32)

	tmpFile := "./imgs/" + tmpName + ".jpg"
	newFile, err := os.Create(tmpFile)
	if err != nil {
		http.Error(w, "cant open file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	hasher := md5.New()
	writtenBytes, err := io.Copy(newFile, io.TeeReader(uploadData, hasher))
	if err != nil {
		http.Error(w, "cant save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := newFile.Sync(); err != nil {
		http.Error(w, "err sync file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	errNewFileClose := newFile.Close()
	if errNewFileClose != nil {
		http.Error(w, "cant close file: "+errNewFileClose.Error(), http.StatusInternalServerError)
		return
	}

	md5Sum := hex.EncodeToString(hasher.Sum(nil))

	realFile := "./imgs/" + md5Sum + ".jpg"
	err = os.Rename(tmpFile, realFile)
	if err != nil {
		http.Error(w, "cant raname file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(ImgResizeTask{handler.Filename, md5Sum})
	fmt.Println("put task ", string(data))

	err = ch.Qp.Publish(ImageResizeQueueName, data)
	if err != nil {
		http.Error(w, "cant publish task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, errUploadWriten := fmt.Fprintf(w, "Upload %d bytes successful\n Image processing", writtenBytes)
	if errUploadWriten != nil {
		log.Println(errUploadWriten)
	}
}
