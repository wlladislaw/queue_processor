package helpers

import "log"

func FatalOnError(msg string, err error) {
	if err != nil {
		log.Fatalf(msg+" %s", err)
	}
}
