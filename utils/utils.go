package utils

import (
	"log"
)

func HandleError(err error, message string) {
	log.Fatalf("%s: %v", message, err)
}
