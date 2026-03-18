package internal

import (
	"encoding/json"
	"fmt"
)

type GELFMessage struct {
	Version      string                 `json:"version"`
	Host         string                 `json:"host"`
	ShortMessage string                 `json:"short_message"`
	FullMessage  string                 `json:"full_message,omitempty"`
	Timestamp    float64                `json:"timestamp"`
	Level        int                    `json:"level,omitempty"`
	Facility     string                 `json:"facility,omitempty"`
	Line         int                    `json:"line,omitempty"`
	File         string                 `json:"file,omitempty"`
	ExtraFields  map[string]interface{} `json:"-"`
}

// validateGELF checks if the GELF message has required fields
func (g *GELFMessage) validate() error {
	if g.Version == "" {
		return fmt.Errorf("missing required field: version")
	}
	if g.Host == "" {
		return fmt.Errorf("missing required field: host")
	}
	if g.ShortMessage == "" {
		return fmt.Errorf("missing required field: short_message")
	}
	if g.Timestamp == 0 {
		return fmt.Errorf("missing required field: timestamp")
	}
	return nil
}
type OTLPLogRecord struct {
	TimeUnixNano         string            `json:"timeUnixNano"`
	ObservedTimeUnixNano string            `json:"observedTimeUnixNano,omitempty"`
	SeverityNumber       int               `json:"severityNumber,omitempty"`
	SeverityText         string            `json:"severityText,omitempty"`
	TraceID              string            `json:"traceId,omitempty"`
	SpanID               string            `json:"spanId,omitempty"`
	Body                 map[string]string `json:"body,omitempty"`
	Attributes           []OTLPAttribute   `json:"attributes"`
}

type OTLPAttribute struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type OTLPScopeLog struct {
	Scope struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"scope"`
	LogRecords []OTLPLogRecord `json:"logRecords"`
}

type OTLPResourceLog struct {
	Resource struct {
		Attributes []OTLPAttribute `json:"attributes"`
	} `json:"resource"`
	ScopeLogs []OTLPScopeLog `json:"scopeLogs"`
}

func TransformToOTLP(gelfMessage []byte) ([]byte, error) {
	// Parse GELF message with extra fields support
	var gelfRaw map[string]interface{}
	if err := json.Unmarshal(gelfMessage, &gelfRaw); err != nil {
		return nil, fmt.Errorf("failed to parse GELF message: %w", err)
	}

	// Convert to strongly-typed GELFMessage
	var gelf GELFMessage
	if err := json.Unmarshal(gelfMessage, &gelf); err != nil {
		return nil, fmt.Errorf("failed to parse GELF message: %w", err)
	}

	// Extract extra fields (anything starting with _)
	for key, value := range gelfRaw {
		if len(key) > 0 && key[0] == '_' {
			if gelf.ExtraFields == nil {
				gelf.ExtraFields = make(map[string]interface{})
			}
			gelf.ExtraFields[key] = value
		}
	}

	// Validate required fields
	if err := gelf.validate(); err != nil {
		return nil, fmt.Errorf("invalid GELF message: %w", err)
	}

	// Check if "message" field should be used as "short_message"
	if gelf.ShortMessage == "" {
		if message, exists := gelfRaw["message"]; exists {
			if messageStr, ok := message.(string); ok {
				gelf.ShortMessage = messageStr
			}
		}
	}

	// Convert Level to Severity
	severityMap := map[int]string{
		0: "EMERGENCY",
		1: "ALERT",
		2: "CRITICAL",
		3: "ERROR",
		4: "WARNING",
		5: "NOTICE",
		6: "INFO",
		7: "DEBUG",
	}
	severity := severityMap[gelf.Level]

	// Prepare OTLP resource attributes
	resourceAttributes := []OTLPAttribute{
		{"host.name", gelf.Host},
	}

	// Prepare OTLP log record attributes
	attributes := []OTLPAttribute{
		{"full_message", gelf.FullMessage},
		{"facility", gelf.Facility},
		{"file", gelf.File},
		{"line", fmt.Sprintf("%d", gelf.Line)},
	}
	for key, value := range gelf.ExtraFields {
		if key == "_id" {
			continue // Ignore _id field
		}
		attributes = append(attributes, OTLPAttribute{Key: key, Value: value})
	}

	// Create OTLP log record
	logRecord := OTLPLogRecord{
		TimeUnixNano:   fmt.Sprintf("%d000000000", int64(gelf.Timestamp)), // Convert seconds to nanoseconds
		SeverityNumber: gelf.Level,
		SeverityText:   severity,
		Body: map[string]string{
			"stringValue": gelf.ShortMessage,
		},
		Attributes: attributes,
	}

	// Create OTLP scope logs
	scopeLog := OTLPScopeLog{
		Scope: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "gelf-forwarder",
			Version: "1.0.0",
		},
		LogRecords: []OTLPLogRecord{logRecord},
	}

	// Create OTLP resource log
	resourceLog := OTLPResourceLog{
		Resource: struct {
			Attributes []OTLPAttribute `json:"attributes"`
		}{
			Attributes: resourceAttributes,
		},
		ScopeLogs: []OTLPScopeLog{scopeLog},
	}

	// Convert to JSON
	return json.Marshal(map[string]interface{}{
		"resourceLogs": []OTLPResourceLog{resourceLog},
	})
}
