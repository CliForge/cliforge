package secrets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/CliForge/cliforge/pkg/cli"
)

// MaskValue masks a sensitive value using the specified masking strategy.
func MaskValue(value string, config *cli.SecretsMasking) string {
	if config == nil {
		return defaultMask(value)
	}

	switch config.Style {
	case "full":
		return fullMask(config.Replacement)
	case "hash":
		return hashMask(value)
	case "partial":
		return partialMask(value, config.PartialShowChars, config.Replacement)
	default:
		return partialMask(value, config.PartialShowChars, config.Replacement)
	}
}

// defaultMask applies default masking (partial with 6 chars).
func defaultMask(value string) string {
	return partialMask(value, 6, "***")
}

// fullMask completely masks the value.
func fullMask(replacement string) string {
	if replacement == "" {
		return "***"
	}
	return replacement
}

// partialMask shows the first N characters and masks the rest.
func partialMask(value string, showChars int, replacement string) string {
	if replacement == "" {
		replacement = "***"
	}

	// If value is too short, fully mask it
	if len(value) <= showChars {
		return replacement
	}

	// Show first N chars, mask the rest
	return value[:showChars] + replacement
}

// hashMask creates a SHA256 hash of the value for audit purposes.
func hashMask(value string) string {
	hash := sha256.Sum256([]byte(value))
	hashStr := hex.EncodeToString(hash[:])
	// Return first 16 chars of hash with prefix
	return "sha256:" + hashStr[:16]
}

// MaskingWriter wraps an io.Writer and masks secrets before writing.
type MaskingWriter struct {
	detector *Detector
	context  string
	delegate interface {
		Write(p []byte) (n int, err error)
	}
}

// NewMaskingWriter creates a new MaskingWriter.
func NewMaskingWriter(detector *Detector, context string, delegate interface {
	Write(p []byte) (n int, err error)
}) *MaskingWriter {
	return &MaskingWriter{
		detector: detector,
		context:  context,
		delegate: delegate,
	}
}

// Write implements io.Writer, masking secrets before writing to the delegate.
func (w *MaskingWriter) Write(p []byte) (n int, err error) {
	// Check if masking should be applied in this context
	if !w.detector.ShouldMaskInContext(w.context) {
		return w.delegate.Write(p)
	}

	// Mask the content
	masked := w.detector.MaskString(string(p))

	// Write masked content
	_, err = w.delegate.Write([]byte(masked))
	if err != nil {
		return 0, err
	}

	// Return original byte count to maintain io.Writer contract
	return len(p), nil
}

// MaskFormatter provides formatting-aware masking for structured output.
type MaskFormatter struct {
	detector *Detector
}

// NewMaskFormatter creates a new MaskFormatter.
func NewMaskFormatter(detector *Detector) *MaskFormatter {
	return &MaskFormatter{
		detector: detector,
	}
}

// FormatJSON formats and masks JSON data.
func (f *MaskFormatter) FormatJSON(data interface{}) string {
	masked := f.detector.MaskJSON(data)
	// Note: Actual JSON formatting would be done by the output package
	return fmt.Sprintf("%v", masked)
}

// FormatTable formats and masks table data.
// Each row is a map of column name to value.
func (f *MaskFormatter) FormatTable(rows []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rows))

	for i, row := range rows {
		maskedRow := make(map[string]interface{}, len(row))

		for key, value := range row {
			// Detect and mask if necessary
			detection := f.detector.DetectInField(key, value)
			if detection.IsSecret {
				maskedRow[key] = MaskValue(fmt.Sprint(value), f.detector.config.Masking)
			} else {
				maskedRow[key] = value
			}
		}

		result[i] = maskedRow
	}

	return result
}

// MaskStrategy defines the interface for custom masking strategies.
type MaskStrategy interface {
	// Mask takes a value and returns the masked version
	Mask(value string) string

	// Name returns the name of the masking strategy
	Name() string
}

// PartialMaskStrategy implements partial masking.
type PartialMaskStrategy struct {
	showChars   int
	replacement string
}

// NewPartialMaskStrategy creates a new partial masking strategy.
func NewPartialMaskStrategy(showChars int, replacement string) *PartialMaskStrategy {
	return &PartialMaskStrategy{
		showChars:   showChars,
		replacement: replacement,
	}
}

func (s *PartialMaskStrategy) Mask(value string) string {
	return partialMask(value, s.showChars, s.replacement)
}

func (s *PartialMaskStrategy) Name() string {
	return "partial"
}

// FullMaskStrategy implements full masking.
type FullMaskStrategy struct {
	replacement string
}

// NewFullMaskStrategy creates a new full masking strategy.
func NewFullMaskStrategy(replacement string) *FullMaskStrategy {
	return &FullMaskStrategy{
		replacement: replacement,
	}
}

func (s *FullMaskStrategy) Mask(value string) string {
	return fullMask(s.replacement)
}

func (s *FullMaskStrategy) Name() string {
	return "full"
}

// HashMaskStrategy implements hash masking.
type HashMaskStrategy struct{}

// NewHashMaskStrategy creates a new hash masking strategy.
func NewHashMaskStrategy() *HashMaskStrategy {
	return &HashMaskStrategy{}
}

func (s *HashMaskStrategy) Mask(value string) string {
	return hashMask(value)
}

func (s *HashMaskStrategy) Name() string {
	return "hash"
}

// CreateMaskStrategy creates a MaskStrategy from configuration.
func CreateMaskStrategy(config *cli.SecretsMasking) MaskStrategy {
	if config == nil {
		return NewPartialMaskStrategy(6, "***")
	}

	switch config.Style {
	case "full":
		return NewFullMaskStrategy(config.Replacement)
	case "hash":
		return NewHashMaskStrategy()
	case "partial":
		return NewPartialMaskStrategy(config.PartialShowChars, config.Replacement)
	default:
		return NewPartialMaskStrategy(config.PartialShowChars, config.Replacement)
	}
}

// StructurePreservingMask masks a value while preserving its structure.
// This is useful for debugging where you want to see the shape but not the content.
func StructurePreservingMask(value string, config *cli.SecretsMasking) string {
	if config == nil || config.Style != "partial" {
		return MaskValue(value, config)
	}

	// Preserve structure indicators like prefixes and separators
	// Example: "sk_live_abc123def456" -> "sk_liv***"
	// Example: "Bearer eyJhbGc..." -> "Bearer eyJ***"

	// Common separators to preserve
	separators := []string{"_", "-", ".", " "}

	for _, sep := range separators {
		if len(value) > 0 && containsSeparator(value, sep) {
			parts := splitBySeparator(value, sep)
			maskedParts := make([]string, len(parts))

			for i, part := range parts {
				if i == 0 && len(part) <= config.PartialShowChars {
					// Keep first part if it's short (likely a prefix)
					maskedParts[i] = part
				} else if i == 0 {
					maskedParts[i] = partialMask(part, config.PartialShowChars, "")
				} else {
					// Mask subsequent parts
					maskedParts[i] = config.Replacement
					break
				}
			}

			return joinParts(maskedParts, sep) + config.Replacement
		}
	}

	// No separator found, use standard partial mask
	return partialMask(value, config.PartialShowChars, config.Replacement)
}

// Helper functions

func containsSeparator(value, sep string) bool {
	for i := 0; i < len(value); i++ {
		if string(value[i]) == sep {
			return true
		}
	}
	return false
}

func splitBySeparator(value, sep string) []string {
	var parts []string
	current := ""

	for i := 0; i < len(value); i++ {
		if string(value[i]) == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(value[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func joinParts(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}

	return result
}
