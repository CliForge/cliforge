package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/spf13/cobra"
)

// VersionInfo contains version information about the CLI and API.
type VersionInfo struct {
	ClientVersion string    `json:"client_version"`
	ServerVersion string    `json:"server_version,omitempty"`
	APITitle      string    `json:"api_title,omitempty"`
	OpenAPISpec   string    `json:"openapi_spec,omitempty"`
	Built         time.Time `json:"built,omitempty"`
	GoVersion     string    `json:"go_version"`
	Platform      string    `json:"platform"`
	Compiler      string    `json:"compiler"`
}

// VersionOptions configures the version command behavior.
type VersionOptions struct {
	Config         *cli.Config
	BuildTime      time.Time
	ShowAPIVersion bool
	OutputFormat   string
	Output         io.Writer
}

// NewVersionCommand creates a new version command.
func NewVersionCommand(opts *VersionOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Display version information for both the CLI binary and the API server.

The version command shows:
- CLI binary version (client version)
- API server version (if available)
- Build information (build time, Go version, platform)
- OpenAPI specification version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "Output format (text|json|yaml)")

	return cmd
}

// runVersion executes the version command.
func runVersion(opts *VersionOptions) error {
	info := &VersionInfo{
		ClientVersion: opts.Config.Metadata.Version,
		GoVersion:     runtime.Version(),
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Compiler:      runtime.Compiler,
		Built:         opts.BuildTime,
	}

	// Add API version if enabled
	if opts.ShowAPIVersion && opts.Config.API.Version != "" {
		info.ServerVersion = opts.Config.API.Version
		info.OpenAPISpec = opts.Config.API.Version
	}

	// Add API title if available
	if opts.Config.Metadata.Description != "" {
		info.APITitle = opts.Config.Metadata.Description
	}

	// Format output
	switch opts.OutputFormat {
	case "json":
		return formatVersionJSON(info, opts.Output)
	case "yaml":
		return formatVersionYAML(info, opts.Output)
	default:
		return formatVersionText(info, opts.Output)
	}
}

// formatVersionText formats version info as human-readable text.
func formatVersionText(info *VersionInfo, w io.Writer) error {
	_, _ = fmt.Fprintf(w, "Client Version: %s\n", info.ClientVersion)

	if info.ServerVersion != "" {
		_, _ = fmt.Fprintf(w, "Server Version: %s", info.ServerVersion)
		if info.APITitle != "" {
			_, _ = fmt.Fprintf(w, " (%s)", info.APITitle)
		}
		_, _ = fmt.Fprintln(w)
	}

	if info.OpenAPISpec != "" {
		_, _ = fmt.Fprintf(w, "OpenAPI Spec: %s\n", info.OpenAPISpec)
	}

	if !info.Built.IsZero() {
		_, _ = fmt.Fprintf(w, "Built: %s\n", info.Built.Format(time.RFC3339))
	}

	_, _ = fmt.Fprintf(w, "Go: %s\n", info.GoVersion)
	_, _ = fmt.Fprintf(w, "Platform: %s\n", info.Platform)
	_, _ = fmt.Fprintf(w, "Compiler: %s\n", info.Compiler)

	return nil
}

// formatVersionJSON formats version info as JSON.
func formatVersionJSON(info *VersionInfo, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

// formatVersionYAML formats version info as YAML.
func formatVersionYAML(info *VersionInfo, w io.Writer) error {
	_, _ = fmt.Fprintf(w, "client_version: %s\n", info.ClientVersion)

	if info.ServerVersion != "" {
		_, _ = fmt.Fprintf(w, "server_version: %s\n", info.ServerVersion)
	}

	if info.APITitle != "" {
		_, _ = fmt.Fprintf(w, "api_title: %s\n", info.APITitle)
	}

	if info.OpenAPISpec != "" {
		_, _ = fmt.Fprintf(w, "openapi_spec: %s\n", info.OpenAPISpec)
	}

	if !info.Built.IsZero() {
		_, _ = fmt.Fprintf(w, "built: %s\n", info.Built.Format(time.RFC3339))
	}

	_, _ = fmt.Fprintf(w, "go_version: %s\n", info.GoVersion)
	_, _ = fmt.Fprintf(w, "platform: %s\n", info.Platform)
	_, _ = fmt.Fprintf(w, "compiler: %s\n", info.Compiler)

	return nil
}

// GetVersionShort returns a short version string suitable for --version flag.
func GetVersionShort(config *cli.Config) string {
	clientVer := config.Metadata.Version
	serverVer := config.API.Version

	if serverVer != "" {
		return fmt.Sprintf("Client: %s  Server: %s", clientVer, serverVer)
	}

	return fmt.Sprintf("v%s", clientVer)
}
