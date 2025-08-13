package protocol

import (
	"time"
)

// MessageType represents the type of IPC message
type MessageType string

const (
	// Health check message
	MessageTypeHealthCheck MessageType = "health_check"
	// Health check response
	MessageTypeHealthResponse MessageType = "health_response"
	// Database connection request
	MessageTypeDBConnect MessageType = "db_connect"
	// Database connection response
	MessageTypeDBConnectResponse MessageType = "db_connect_response"
	// Query execution request
	MessageTypeQuery MessageType = "query"
	// Query execution response
	MessageTypeQueryResponse MessageType = "query_response"
	// Error message
	MessageTypeError MessageType = "error"
)

// Message represents a generic IPC message
type Message struct {
	ID        string                 `msgpack:"id"`
	Type      MessageType            `msgpack:"type"`
	Timestamp time.Time              `msgpack:"timestamp"`
	Data      map[string]interface{} `msgpack:"data"`
}

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	Message
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Message
	Status    string `msgpack:"status"`
	Timestamp int64  `msgpack:"timestamp"`
	Version   string `msgpack:"version"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message
	Error   string `msgpack:"error"`
	Code    int    `msgpack:"code"`
	Details string `msgpack:"details"`
}

// NewMessage creates a new message with the given type and data
func NewMessage(msgType MessageType, data map[string]interface{}) *Message {
	return &Message{
		ID:        generateID(),
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewHealthCheckRequest creates a new health check request
func NewHealthCheckRequest() *HealthCheckRequest {
	return &HealthCheckRequest{
		Message: *NewMessage(MessageTypeHealthCheck, nil),
	}
}

// NewHealthCheckResponse creates a new health check response
func NewHealthCheckResponse(status, version string) *HealthCheckResponse {
	return &HealthCheckResponse{
		Message:   *NewMessage(MessageTypeHealthResponse, nil),
		Status:    status,
		Timestamp: time.Now().Unix(),
		Version:   version,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, code int, details string) *ErrorResponse {
	return &ErrorResponse{
		Message: *NewMessage(MessageTypeError, nil),
		Error:   err.Error(),
		Code:    code,
		Details: details,
	}
}

// generateID generates a simple unique ID for messages
func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
