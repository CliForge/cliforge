package builtin

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/plugin"
	"gopkg.in/yaml.v3"
)

// FileOpsPlugin handles file operations: read, parse, validate, transform.
type FileOpsPlugin struct {
	allowedPaths []string
	maxFileSize  int64
}

// NewFileOpsPlugin creates a new file operations plugin.
func NewFileOpsPlugin(allowedPaths []string, maxFileSize int64) *FileOpsPlugin {
	if maxFileSize == 0 {
		maxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	return &FileOpsPlugin{
		allowedPaths: allowedPaths,
		maxFileSize:  maxFileSize,
	}
}

// Execute performs file operations.
func (p *FileOpsPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	operation, ok := input.Data["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation is required")
	}

	startTime := time.Now()

	var result map[string]interface{}
	var err error

	switch operation {
	case "read":
		result, err = p.readFile(input)
	case "parse":
		result, err = p.parseFile(input)
	case "validate":
		result, err = p.validateFile(input)
	case "transform":
		result, err = p.transformFile(input)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	duration := time.Since(startTime)

	if err != nil {
		return &plugin.PluginOutput{
			ExitCode: 1,
			Error:    err.Error(),
			Duration: duration,
		}, nil
	}

	return &plugin.PluginOutput{
		ExitCode: 0,
		Data:     result,
		Duration: duration,
	}, nil
}

// Validate checks if the plugin is properly configured.
func (p *FileOpsPlugin) Validate() error {
	return nil
}

// Describe returns plugin metadata.
func (p *FileOpsPlugin) Describe() *plugin.PluginInfo {
	manifest := plugin.PluginManifest{
		Name:        "file-ops",
		Version:     "1.0.0",
		Type:        plugin.PluginTypeBuiltin,
		Description: "Read, parse, validate, and transform files",
		Author:      "CliForge",
		Permissions: []plugin.Permission{
			{
				Type:        plugin.PermissionReadFile,
				Resource:    "*",
				Description: "Read files from the filesystem",
			},
		},
	}

	return &plugin.PluginInfo{
		Manifest: manifest,
		Capabilities: []string{
			"read", "parse", "validate", "transform",
			"formats:json", "formats:yaml", "formats:pem",
			"formats:htpasswd", "formats:x509",
		},
		Status: plugin.PluginStatusReady,
	}
}

// readFile reads a file and returns its content.
func (p *FileOpsPlugin) readFile(input *plugin.PluginInput) (map[string]interface{}, error) {
	filePath, ok := input.Data["file"].(string)
	if !ok {
		return nil, fmt.Errorf("file path is required")
	}

	// Validate path
	if err := p.validatePath(filePath); err != nil {
		return nil, err
	}

	// Check file size
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > p.maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), p.maxFileSize)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content": string(data),
		"size":    info.Size(),
		"path":    filePath,
	}, nil
}

// parseFile parses a file in various formats.
func (p *FileOpsPlugin) parseFile(input *plugin.PluginInput) (map[string]interface{}, error) {
	filePath, ok := input.Data["file"].(string)
	if !ok {
		return nil, fmt.Errorf("file path is required")
	}
	_ = filePath

	format, ok := input.Data["format"].(string)
	if !ok {
		return nil, fmt.Errorf("format is required")
	}

	// Read file first
	result, err := p.readFile(input)
	if err != nil {
		return nil, err
	}

	content := result["content"].(string)

	// Parse based on format
	switch format {
	case "json":
		return p.parseJSON(content)
	case "yaml":
		return p.parseYAML(content)
	case "pem":
		return p.parsePEM(content, input)
	case "htpasswd":
		return p.parseHTPasswd(content)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// validateFile validates a file's format and content.
func (p *FileOpsPlugin) validateFile(input *plugin.PluginInput) (map[string]interface{}, error) {
	// Parse the file first
	result, err := p.parseFile(input)
	if err != nil {
		return map[string]interface{}{
			"valid":   false,
			"message": err.Error(),
		}, nil
	}

	// If parsing succeeded, file is valid
	result["valid"] = true
	return result, nil
}

// transformFile transforms file content.
func (p *FileOpsPlugin) transformFile(input *plugin.PluginInput) (map[string]interface{}, error) {
	transformation, ok := input.Data["transformation"].(string)
	if !ok {
		return nil, fmt.Errorf("transformation is required")
	}

	switch transformation {
	case "base64-encode":
		return p.base64Encode(input)
	case "base64-decode":
		return p.base64Decode(input)
	default:
		return nil, fmt.Errorf("unsupported transformation: %s", transformation)
	}
}

// parseJSON parses JSON content.
func (p *FileOpsPlugin) parseJSON(content string) (map[string]interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return map[string]interface{}{
		"parsed": data,
		"format": "json",
	}, nil
}

// parseYAML parses YAML content.
func (p *FileOpsPlugin) parseYAML(content string) (map[string]interface{}, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	return map[string]interface{}{
		"parsed": data,
		"format": "yaml",
	}, nil
}

// parsePEM parses PEM-encoded content.
func (p *FileOpsPlugin) parsePEM(content string, input *plugin.PluginInput) (map[string]interface{}, error) {
	block, _ := pem.Decode([]byte(content))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	result := map[string]interface{}{
		"type":   block.Type,
		"format": "pem",
		"base64": base64.StdEncoding.EncodeToString(block.Bytes),
	}

	// If it's a certificate, parse additional details
	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}

		result["subject"] = cert.Subject.String()
		result["issuer"] = cert.Issuer.String()
		result["not_before"] = cert.NotBefore.Format(time.RFC3339)
		result["not_after"] = cert.NotAfter.Format(time.RFC3339)
		result["serial_number"] = cert.SerialNumber.String()

		if len(cert.DNSNames) > 0 {
			result["dns_names"] = cert.DNSNames
		}
	}

	return result, nil
}

// parseHTPasswd parses htpasswd file format.
func (p *FileOpsPlugin) parseHTPasswd(content string) (map[string]interface{}, error) {
	lines := strings.Split(content, "\n")
	users := make([]map[string]string, 0)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid htpasswd format at line %d", i+1)
		}

		users = append(users, map[string]string{
			"username": parts[0],
			"password": parts[1], // hashed password
		})
	}

	return map[string]interface{}{
		"users":  users,
		"format": "htpasswd",
		"count":  len(users),
	}, nil
}

// base64Encode encodes file content to base64.
func (p *FileOpsPlugin) base64Encode(input *plugin.PluginInput) (map[string]interface{}, error) {
	result, err := p.readFile(input)
	if err != nil {
		return nil, err
	}

	content := result["content"].(string)
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	return map[string]interface{}{
		"encoded": encoded,
		"size":    len(encoded),
	}, nil
}

// base64Decode decodes base64 content.
func (p *FileOpsPlugin) base64Decode(input *plugin.PluginInput) (map[string]interface{}, error) {
	content, ok := input.Data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	return map[string]interface{}{
		"decoded": string(decoded),
		"size":    len(decoded),
	}, nil
}

// validatePath validates that a file path is allowed.
func (p *FileOpsPlugin) validatePath(path string) error {
	// Clean and resolve the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// If allowed paths are specified, check if path is in allowed list
	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range p.allowedPaths {
			absAllowed, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}

			// Check if path is within allowed path
			if strings.HasPrefix(absPath, absAllowed) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("path not in allowed list: %s", absPath)
		}
	}

	return nil
}
