package rules

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"Float64", 42.5, 42.5, true},
		{"Float32", float32(3.14), 3.14, true},
		{"Int", 42, 42.0, true},
		{"Int64", int64(100), 100.0, true},
		{"Uint", uint(50), 50.0, true},
		{"JSON Number", json.Number("2.718"), 2.718, true},
		{"String number", "123.45", 123.45, true},
		{"Invalid string", "not a number", 0, false},
		{"Invalid JSON Number", json.Number("invalid"), 0, false},
		{"Nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toFloat64(tt.input)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		ok       bool
	}{
		{"Int", 42, 42, true},
		{"Float64 integer", 42.0, 42, true},
		{"Float64 non-integer", 42.5, 0, false},
		{"String integer", "123", 123, true},
		{"String non-integer", "123.45", 0, false},
		{"JSON Number", json.Number("100"), 100, true},
		{"Invalid input", "abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toInt(tt.input)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		ok       bool
	}{
		{"String", "hello", "hello", true},
		{"Byte slice", []byte("world"), "world", true},
		{"Int", 42, "42", true},
		{"Custom stringer", fmt.Errorf("error"), "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toString(tt.input)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
		ok       bool
	}{
		{"Bool true", true, true, true},
		{"Bool false", false, false, true},
		{"String true", "true", true, true},
		{"String yes", "yes", true, true},
		{"String 1", "1", true, true},
		{"String false", "false", false, true},
		{"Int non-zero", 42, true, true},
		{"Int zero", 0, false, true},
		{"Float non-zero", 3.14, true, true},
		{"Invalid", struct{}{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toBool(tt.input)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatValidators(t *testing.T) {
	tests := []struct {
		name     string
		format   func(string2 string) bool
		input    string
		expected bool
	}{
		{"Email valid", validateEmail, "test@example.com", true},
		{"Email invalid", validateEmail, "invalid", false},
		{"DateTime valid", validateDateTime, "2023-10-05T14:30:00Z", true},
		{"DateTime invalid", validateDateTime, "2023-13-01", false},
		{"Date valid", validateDate, "2023-10-05", true},
		{"Date invalid", validateDate, "2023-10-32", false},
		{"Time valid", validateTime, "14:30:00", true},
		{"Time invalid", validateTime, "25:00:00", false},
		{"URI valid", validateURI, "https://example.com", true},
		{"URI invalid", validateURI, "://invalid", false},
		{"Hostname valid", validateHostname, "example.com", true},
		{"Hostname invalid", validateHostname, "invalid..com", false},
		{"IPv4 valid", validateIPv4, "192.168.1.1", true},
		{"IPv4 invalid", validateIPv4, "256.1.2.3", false},
		{"IPv6 valid", validateIPv6, "2001:db8::1", true},
		{"IPv6 invalid", validateIPv6, "2001::db8::1", false},
		{"UUID valid", validateUUID, "123e4567-e89b-12d3-a456-426614174000", true},
		{"UUID invalid", validateUUID, "invalid-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.format(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
