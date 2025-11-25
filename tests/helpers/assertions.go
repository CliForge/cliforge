package helpers

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// AssertJSONEqual asserts that two JSON strings are equal.
func AssertJSONEqual(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedJSON, actualJSON interface{}

	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	expectedStr := formatJSON(expectedJSON)
	actualStr := formatJSON(actualJSON)

	if expectedStr != actualStr {
		t.Fatalf("JSON mismatch\nExpected:\n%s\n\nActual:\n%s", expectedStr, actualStr)
	}
}

// AssertFileExists asserts that a file exists.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if !FileExists(path) {
		t.Fatalf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if FileExists(path) {
		t.Fatalf("Expected file not to exist: %s", path)
	}
}

// AssertStringContains asserts that a string contains a substring.
func AssertStringContains(t *testing.T, str, substr string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Fatalf("Expected string to contain %q\nString: %s", substr, str)
	}
}

// AssertStringNotContains asserts that a string does not contain a substring.
func AssertStringNotContains(t *testing.T, str, substr string) {
	t.Helper()
	if strings.Contains(str, substr) {
		t.Fatalf("Expected string not to contain %q\nString: %s", substr, str)
	}
}

// AssertEqual asserts that two values are equal.
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal.
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Fatalf("Expected values to be different, but both are %v", expected)
	}
}

// AssertNil asserts that a value is nil.
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Fatalf("Expected nil, got %v", value)
	}
}

// AssertNotNil asserts that a value is not nil.
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatalf("Expected non-nil value")
	}
}

// AssertNoError asserts that an error is nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError asserts that an error is not nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

// AssertErrorContains asserts that an error contains a substring.
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected an error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("Expected error to contain %q\nError: %v", substr, err)
	}
}

// AssertTrue asserts that a value is true.
func AssertTrue(t *testing.T, value bool, msg string) {
	t.Helper()
	if !value {
		t.Fatalf("Expected true: %s", msg)
	}
}

// AssertFalse asserts that a value is false.
func AssertFalse(t *testing.T, value bool, msg string) {
	t.Helper()
	if value {
		t.Fatalf("Expected false: %s", msg)
	}
}

// AssertGreaterThan asserts that a value is greater than another.
func AssertGreaterThan(t *testing.T, value, threshold int, msg string) {
	t.Helper()
	if value <= threshold {
		t.Fatalf("Expected %d > %d: %s", value, threshold, msg)
	}
}

// AssertLessThan asserts that a value is less than another.
func AssertLessThan(t *testing.T, value, threshold int, msg string) {
	t.Helper()
	if value >= threshold {
		t.Fatalf("Expected %d < %d: %s", value, threshold, msg)
	}
}

// AssertDurationLessThan asserts that a duration is less than a threshold.
func AssertDurationLessThan(t *testing.T, duration, threshold time.Duration, msg string) {
	t.Helper()
	if duration >= threshold {
		t.Fatalf("Expected duration %s < %s: %s", duration, threshold, msg)
	}
}

// AssertEventually retries an assertion until it passes or times out.
func AssertEventually(t *testing.T, assertion func() bool, timeout, interval time.Duration, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if assertion() {
			return
		}
		time.Sleep(interval)
	}

	t.Fatalf("Assertion timed out after %s: %s", timeout, msg)
}

// AssertSliceContains asserts that a slice contains a value.
func AssertSliceContains(t *testing.T, slice []string, value string) {
	t.Helper()
	for _, item := range slice {
		if item == value {
			return
		}
	}
	t.Fatalf("Expected slice to contain %q\nSlice: %v", value, slice)
}

// AssertSliceNotContains asserts that a slice does not contain a value.
func AssertSliceNotContains(t *testing.T, slice []string, value string) {
	t.Helper()
	for _, item := range slice {
		if item == value {
			t.Fatalf("Expected slice not to contain %q\nSlice: %v", value, slice)
		}
	}
}

// AssertMapContains asserts that a map contains a key.
func AssertMapContains(t *testing.T, m map[string]interface{}, key string) {
	t.Helper()
	if _, exists := m[key]; !exists {
		t.Fatalf("Expected map to contain key %q\nMap: %v", key, m)
	}
}

// formatJSON formats a JSON object for comparison.
func formatJSON(data interface{}) string {
	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}

// ParseJSON parses a JSON string into a map.
func ParseJSON(t *testing.T, data string) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nData: %s", err, data)
	}
	return result
}

// ParseJSONArray parses a JSON string into a slice.
func ParseJSONArray(t *testing.T, data string) []interface{} {
	t.Helper()

	var result []interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nData: %s", err, data)
	}
	return result
}

// MustMarshalJSON marshals data to JSON or fails the test.
func MustMarshalJSON(t *testing.T, data interface{}) string {
	t.Helper()

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return string(b)
}

// Retry retries a function until it succeeds or times out.
func Retry(t *testing.T, fn func() error, maxAttempts int, delay time.Duration) error {
	t.Helper()

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < maxAttempts-1 {
				time.Sleep(delay)
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
