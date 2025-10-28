package logger

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

// APILogEvent represents an API request/response log event
type APILogEvent struct {
	RequestID    null.String `json:"request_id"`
	Service      string      `json:"service"`
	URL          string      `json:"url"`
	Method       string      `json:"method"`
	ResponseCode int         `json:"response_code"`
	ResponseBody null.String `json:"response_body,omitempty"`
	RequestBody  null.String `json:"request_body,omitempty"`
	UserID       null.String `json:"user_id,omitempty"`
	Duration     float64     `json:"duration"`
	Version      string      `json:"version"`
	Name         string      `json:"name"`
	CreatedAt    time.Time   `json:"created_at"`
}
