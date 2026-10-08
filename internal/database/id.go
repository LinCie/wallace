package database

import "github.com/gofrs/uuid/v5"

func GenerateID() (uuid.UUID, error) {
	return uuid.NewV7()
}
