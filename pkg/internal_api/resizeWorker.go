package internal_api

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"

	"github.com/nfnt/resize"
)

var (
	sizes = []uint{80, 160, 320}
)

type ImgResizeTask struct {
	Name string
	MD5  string
}

func ResizeWorker(task *ImgResizeTask) error {
	originalPath := fmt.Sprintf("./imgs/%s.jpg", task.MD5)
	for _, size := range sizes {
		resizedPath := fmt.Sprintf("./imgs/%s_%d.jpg", task.MD5, size)
		err := ResizeImage(originalPath, resizedPath, size)
		if err != nil {
			log.Println("resize failed", err)
			return err
		}
	}
	return nil
}

func ResizeImage(originalPath string, resizedPath string, size uint) error {
	file, err := os.Open(originalPath)
	if err != nil {
		return fmt.Errorf("cant open file %s: %s", originalPath, err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		return fmt.Errorf("cant jpeg decode file %s", err)
	}
	errFileClose := file.Close()
	if errFileClose != nil {
		return fmt.Errorf("cant close file %s", errFileClose)
	}

	resizeImage := resize.Resize(size, 0, img, resize.Lanczos3)
	out, err := os.Create(resizedPath)
	if err != nil {
		return fmt.Errorf("cant create file %s: %s", resizedPath, err)
	}

	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Printf("err in out for resizedPath closer defer : %v\n", closeErr)
		}
	}()

	if errEncode := jpeg.Encode(out, resizeImage, nil); errEncode != nil {
		return fmt.Errorf("cant encode img %s ", err)
	}

	return nil
}
