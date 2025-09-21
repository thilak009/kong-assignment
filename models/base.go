package models

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	// gorm:"<-:create" only allows create and read but not update
	// this is avoid updating created_at with a zero value by mistake
	CreatedAt time.Time      `json:"createdAt" gorm:"<-:create"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

type BaseWithId struct {
	Base
	ID string `gorm:"primaryKey" json:"id"`
}

type ErrorResponse struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	TraceId string      `json:"traceId"`
	Details interface{} `json:"details,omitempty"`
}
