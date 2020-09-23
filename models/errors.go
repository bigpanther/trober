package models

import "github.com/gofrs/uuid"

// CustomError
type CustomError struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
	Code    string    `json:"code"`
}
