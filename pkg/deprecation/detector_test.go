package deprecation

import (
	"net/http"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestDetector_DetectOperation(t *testing.T) {
	tests := []struct {
		name        string
		operation   *openapi3.Operation
		operationID string
		method      string
		path        string
		want        bool
	}{
		{
			name: "deprecated operation with standard field",
			operation: &openapi3.Operation{
				Deprecated: true,
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        true,
		},
		{
			name: "non-deprecated operation",
			operation: &openapi3.Operation{
				Deprecated: false,
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        false,
		},
		{
			name: "deprecated operation with x-cli-deprecation extension",
			operation: &openapi3.Operation{
				Deprecated: false,
				Extensions: map[string]interface{}{
					"x-cli-deprecation": map[string]interface{}{
						"sunset": "2025-12-31",
						"reason": "Use v2 API",
					},
				},
			},
			operationID: "listUsersV1",
			method:      "GET",
			path:        "/v1/users",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector()
			result := detector.DetectOperation(tt.operation, tt.operationID, tt.method, tt.path)

			if tt.want && result == nil {
				t.Errorf("DetectOperation() expected deprecation, got nil")
			}
			if !tt.want && result != nil {
				t.Errorf("DetectOperation() expected no deprecation, got %v", result)
			}

			if result != nil {
				if result.OperationID != tt.operationID {
					t.Errorf("OperationID = %v, want %v", result.OperationID, tt.operationID)
				}
				if result.Method != tt.method {
					t.Errorf("Method = %v, want %v", result.Method, tt.method)
				}
				if result.Path != tt.path {
					t.Errorf("Path = %v, want %v", result.Path, tt.path)
				}
			}
		})
	}
}

func TestDetector_DetectParameter(t *testing.T) {
	tests := []struct {
		name        string
		parameter   *openapi3.Parameter
		operationID string
		method      string
		path        string
		want        bool
	}{
		{
			name: "deprecated parameter",
			parameter: &openapi3.Parameter{
				Name:       "filter",
				Deprecated: true,
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        true,
		},
		{
			name: "non-deprecated parameter",
			parameter: &openapi3.Parameter{
				Name:       "limit",
				Deprecated: false,
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector()
			result := detector.DetectParameter(tt.parameter, tt.operationID, tt.method, tt.path)

			if tt.want && result == nil {
				t.Errorf("DetectParameter() expected deprecation, got nil")
			}
			if !tt.want && result != nil {
				t.Errorf("DetectParameter() expected no deprecation, got %v", result)
			}

			if result != nil {
				if result.Name != tt.parameter.Name {
					t.Errorf("Name = %v, want %v", result.Name, tt.parameter.Name)
				}
				if result.Type != DeprecationTypeParameter {
					t.Errorf("Type = %v, want %v", result.Type, DeprecationTypeParameter)
				}
			}
		})
	}
}

func TestDetector_parseDeprecationExtension(t *testing.T) {
	tests := []struct {
		name      string
		extension map[string]interface{}
		want      *DeprecationInfo
	}{
		{
			name: "complete extension",
			extension: map[string]interface{}{
				"sunset": "2025-12-31",
				"replacement": map[string]interface{}{
					"operation": "listUsersV2",
					"path":      "/v2/users",
					"command":   "users list-v2",
					"migration": "Use v2 API",
				},
				"reason":   "Performance issues",
				"docs_url": "https://docs.example.com/migration",
				"severity": "warning",
			},
			want: &DeprecationInfo{
				Reason:   "Performance issues",
				DocsURL:  "https://docs.example.com/migration",
				Severity: SeverityWarning,
			},
		},
		{
			name: "minimal extension",
			extension: map[string]interface{}{
				"sunset": "2025-12-31",
			},
			want: &DeprecationInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector()
			info := &DeprecationInfo{}
			detector.parseDeprecationExtension(tt.extension, info)

			if info.Reason != tt.want.Reason {
				t.Errorf("Reason = %v, want %v", info.Reason, tt.want.Reason)
			}
			if info.DocsURL != tt.want.DocsURL {
				t.Errorf("DocsURL = %v, want %v", info.DocsURL, tt.want.DocsURL)
			}
			if info.Severity != tt.want.Severity {
				t.Errorf("Severity = %v, want %v", info.Severity, tt.want.Severity)
			}

			if info.Sunset == nil && tt.extension["sunset"] != nil {
				t.Errorf("Sunset should be parsed")
			}
		})
	}
}

func TestDetector_DetectFromSunsetHeader(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		operationID string
		method      string
		path        string
		want        bool
	}{
		{
			name: "sunset header present",
			response: &http.Response{
				Header: http.Header{
					"Sunset":      []string{"Sat, 31 Dec 2025 23:59:59 GMT"},
					"Deprecation": []string{"true"},
				},
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        true,
		},
		{
			name: "no sunset header",
			response: &http.Response{
				Header: http.Header{},
			},
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        false,
		},
		{
			name:        "nil response",
			response:    nil,
			operationID: "listUsers",
			method:      "GET",
			path:        "/users",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector()
			result := detector.DetectFromSunsetHeader(tt.response, tt.operationID, tt.method, tt.path)

			if tt.want && result == nil {
				t.Errorf("DetectFromSunsetHeader() expected deprecation, got nil")
			}
			if !tt.want && result != nil {
				t.Errorf("DetectFromSunsetHeader() expected no deprecation, got %v", result)
			}

			if result != nil {
				if result.DetectedFrom != "sunset-header" {
					t.Errorf("DetectedFrom = %v, want sunset-header", result.DetectedFrom)
				}
			}
		})
	}
}

func TestCalculateDaysRemaining(t *testing.T) {
	tests := []struct {
		name   string
		sunset time.Time
		want   int
	}{
		{
			name:   "30 days in future",
			sunset: time.Now().Add(30 * 24 * time.Hour),
			want:   30,
		},
		{
			name:   "90 days in future",
			sunset: time.Now().Add(90 * 24 * time.Hour),
			want:   90,
		},
		{
			name:   "past sunset",
			sunset: time.Now().Add(-10 * 24 * time.Hour),
			want:   -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDaysRemaining(tt.sunset)
			// Allow 1 day tolerance for timing differences
			if got < tt.want-1 || got > tt.want+1 {
				t.Errorf("calculateDaysRemaining() = %v, want ~%v", got, tt.want)
			}
		})
	}
}

func TestCalculateWarningLevel(t *testing.T) {
	tests := []struct {
		name          string
		daysRemaining int
		expectedLevel WarningLevel
	}{
		{
			name:          "removed (past sunset)",
			daysRemaining: -5,
			expectedLevel: WarningLevelRemoved,
		},
		{
			name:          "critical (< 30 days)",
			daysRemaining: 15,
			expectedLevel: WarningLevelCritical,
		},
		{
			name:          "urgent (30-90 days)",
			daysRemaining: 60,
			expectedLevel: WarningLevelUrgent,
		},
		{
			name:          "warning (90-180 days)",
			daysRemaining: 120,
			expectedLevel: WarningLevelWarning,
		},
		{
			name:          "info (> 180 days)",
			daysRemaining: 200,
			expectedLevel: WarningLevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateWarningLevel(tt.daysRemaining)
			if got != tt.expectedLevel {
				t.Errorf("calculateWarningLevel(%d) = %v, want %v", tt.daysRemaining, got, tt.expectedLevel)
			}
		})
	}
}

func TestFilterByType(t *testing.T) {
	deprecations := []*DeprecationInfo{
		{Type: DeprecationTypeOperation, OperationID: "op1"},
		{Type: DeprecationTypeParameter, Name: "param1"},
		{Type: DeprecationTypeOperation, OperationID: "op2"},
	}

	result := FilterByType(deprecations, DeprecationTypeOperation)

	if len(result) != 2 {
		t.Errorf("FilterByType() returned %d items, want 2", len(result))
	}

	for _, dep := range result {
		if dep.Type != DeprecationTypeOperation {
			t.Errorf("FilterByType() returned wrong type: %v", dep.Type)
		}
	}
}

func TestFilterBySeverity(t *testing.T) {
	deprecations := []*DeprecationInfo{
		{Severity: SeverityWarning, OperationID: "op1"},
		{Severity: SeverityBreaking, OperationID: "op2"},
		{Severity: SeverityWarning, OperationID: "op3"},
	}

	result := FilterBySeverity(deprecations, SeverityWarning)

	if len(result) != 2 {
		t.Errorf("FilterBySeverity() returned %d items, want 2", len(result))
	}

	for _, dep := range result {
		if dep.Severity != SeverityWarning {
			t.Errorf("FilterBySeverity() returned wrong severity: %v", dep.Severity)
		}
	}
}

func TestGetMostUrgent(t *testing.T) {
	tests := []struct {
		name         string
		deprecations []*DeprecationInfo
		want         WarningLevel
	}{
		{
			name: "critical is most urgent",
			deprecations: []*DeprecationInfo{
				{Level: WarningLevelInfo},
				{Level: WarningLevelCritical},
				{Level: WarningLevelWarning},
			},
			want: WarningLevelCritical,
		},
		{
			name: "removed is most urgent",
			deprecations: []*DeprecationInfo{
				{Level: WarningLevelUrgent},
				{Level: WarningLevelRemoved},
				{Level: WarningLevelCritical},
			},
			want: WarningLevelRemoved,
		},
		{
			name:         "empty list",
			deprecations: []*DeprecationInfo{},
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMostUrgent(tt.deprecations)

			if tt.want == "" {
				if result != nil {
					t.Errorf("GetMostUrgent() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("GetMostUrgent() = nil, want %v", tt.want)
				} else if result.Level != tt.want {
					t.Errorf("GetMostUrgent().Level = %v, want %v", result.Level, tt.want)
				}
			}
		})
	}
}

func TestIsDeprecated(t *testing.T) {
	tests := []struct {
		name         string
		deprecations []*DeprecationInfo
		want         bool
	}{
		{
			name:         "has deprecations",
			deprecations: []*DeprecationInfo{{OperationID: "op1"}},
			want:         true,
		},
		{
			name:         "empty list",
			deprecations: []*DeprecationInfo{},
			want:         false,
		},
		{
			name:         "nil list",
			deprecations: nil,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDeprecated(tt.deprecations)
			if got != tt.want {
				t.Errorf("IsDeprecated() = %v, want %v", got, tt.want)
			}
		})
	}
}
