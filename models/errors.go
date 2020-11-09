package models

import (
	"log"

	"github.com/gofrs/uuid"
)

// CustomError serializes error messages
type CustomError struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
	Code    string    `json:"code"`
}

//NewCustomError returns an instance of a custom error
func NewCustomError(message string, code string, err error) CustomError {
	id, _ := uuid.NewV4()
	log.Println(id, code, message, "Err:", err)
	return CustomError{ID: id, Code: code, Message: message}
}
