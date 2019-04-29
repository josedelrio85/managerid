package idbsc

import (
	"log"

	uuid "github.com/satori/go.uuid"
)

func getUUID() (uuid.UUID, error) {
	// or error handling
	u2, err := uuid.NewV4()
	if err != nil {
		log.Printf("Something went wrong: %s", err)
		return uuid.Nil, err
	}
	return u2, nil
}
