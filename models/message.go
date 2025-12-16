package models

import (
	"sync"
	"time"
)

// MessageDirection indicates if message was sent or received
type MessageDirection string

const (
	MessageSent     MessageDirection = "Send"
	MessageReceived MessageDirection = "Recv"
)

// MQTTMessage represents a logged MQTT message
type MQTTMessage struct {
	Direction MessageDirection `json:"direction"`
	Topic     string           `json:"topic"`
	Payload   string           `json:"payload"`
	Timestamp time.Time        `json:"timestamp"`
}

// MessageLog stores MQTT messages with a maximum size limit
type MessageLog struct {
	mu       sync.RWMutex
	messages []MQTTMessage
	maxSize  int
}

// NewMessageLog creates a new message log with a maximum size
func NewMessageLog(maxSize int) *MessageLog {
	if maxSize <= 0 {
		maxSize = 1000 // Default max size
	}
	return &MessageLog{
		messages: make([]MQTTMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

// AddMessage adds a message to the log (newest at front)
func (l *MessageLog) AddMessage(direction MessageDirection, topic, payload string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := MQTTMessage{
		Direction: direction,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Insert at beginning (newest first)
	l.messages = append([]MQTTMessage{msg}, l.messages...)

	// Trim to max size
	if len(l.messages) > l.maxSize {
		l.messages = l.messages[:l.maxSize]
	}
}

// GetRecent returns the n most recent messages
func (l *MessageLog) GetRecent(n int) []MQTTMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n <= 0 || n > len(l.messages) {
		n = len(l.messages)
	}

	result := make([]MQTTMessage, n)
	copy(result, l.messages[:n])
	return result
}

// GetAll returns all messages
func (l *MessageLog) GetAll() []MQTTMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]MQTTMessage, len(l.messages))
	copy(result, l.messages)
	return result
}

// Clear removes all messages from the log
func (l *MessageLog) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = make([]MQTTMessage, 0, l.maxSize)
}

// Count returns the number of messages in the log
func (l *MessageLog) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.messages)
}
