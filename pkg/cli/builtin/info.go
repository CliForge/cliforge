package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/CliForge/cliforge/pkg/cache"
	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

// InfoOptions configures the info command behavior.
type InfoOptions struct {
	Config          *cli.CLIConfig
	CheckAPIHealth  bool
	ShowConfigPaths bool
	OutputFormat    string
	Output          io.Writer
	HTTPClient      *http.Client
}

// CLIInfo contains information about the CLI and API.
type CLIInfo struct {
	CLI    CLIDetails    `json:"cli"`
	API    APIDetails    `json:"api"`
	Config ConfigDetails `json:"config"`
	Status StatusDetails `json:"status,omitempty"`
}

// CLIDetails contains CLI metadata.
type CLIDetails struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Homepage    string `json:"homepage,omitempty"`
	DocsURL     string `json:"docs_url,omitempty"`
}

// APIDetails contains API information.
type APIDetails struct {
	Title      string `json:"title,omitempty"`
	Version    string `json:"version,omitempty"`
	BaseURL    string `json:"base_url"`
	SpecURL    string `json:"spec_url"`
	Endpoints  int    `json:"endpoints,omitempty"`
}

// ConfigDetails contains configuration paths.
type ConfigDetails struct {
	ConfigFile string `json:"config_file"`
	CacheDir   string `json:"cache_dir"`
	DataDir    string `json:"data_dir"`
	StateDir   string `json:"state_dir"`
	Auth       string `json:"auth"`
}

// StatusDetails contains API health status.
type StatusDetails struct {
	APIReachable      bool      `json:"api_reachable"`
	AuthValid         bool      `json:"auth_valid,omitempty"`
	SpecCached        bool      `json:"spec_cached"`
	SpecAge           string    `json:"spec_age,omitempty"`
	LastChecked       time.Time `json:"last_checked"`
}

// NewInfoCommand creates a new info command.
func NewInfoCommand(opts *InfoOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show CLI and API information",
		Long: `Display detailed information about the CLI tool and API.

The info command shows:
- CLI metadata (name, version, description)
- API information (title, version, base URL, spec URL)
- Configuration paths (config file, cache, data, state directories)
- API health status (if --check-health is enabled)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.CheckAPIHealth, "check-health", false, "Check API health and connectivity")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "Output format (text|json|yaml)")

	return cmd
}

// runInfo executes the info command.
func runInfo(opts *InfoOptions) error {
	info := buildCLIInfo(opts)

	// Check API health if requested
	if opts.CheckAPIHealth {
		status := checkAPIHealth(opts)
		info.Status = status
	} else {
		// Just check spec cache
		info.Status = checkSpecCache(opts)
	}

	// Format output
	switch opts.OutputFormat {
	case "json":
		return formatInfoJSON(info, opts.Output)
	case "yaml":
		return formatInfoYAML(info, opts.Output)
	default:
		return formatInfoText(info, opts.Output)
	}
}

// buildCLIInfo builds the CLI info structure.
func buildCLIInfo(opts *InfoOptions) *CLIInfo {
	cliName := opts.Config.Metadata.Name

	info := &CLIInfo{
		CLI: CLIDetails{
			Name:        cliName,
			Version:     opts.Config.Metadata.Version,
			Description: opts.Config.Metadata.Description,
			Homepage:    opts.Config.Metadata.Homepage,
			DocsURL:     opts.Config.Metadata.DocsURL,
		},
		API: APIDetails{
			Title:   opts.Config.Metadata.Description,
			Version: opts.Config.API.Version,
			BaseURL: opts.Config.API.BaseURL,
			SpecURL: opts.Config.API.OpenAPIURL,
		},
		Config: ConfigDetails{
			ConfigFile: filepath.Join(xdg.ConfigHome, cliName, "config.yaml"),
			CacheDir:   filepath.Join(xdg.CacheHome, cliName),
			DataDir:    filepath.Join(xdg.DataHome, cliName),
			StateDir:   filepath.Join(xdg.StateHome, cliName),
			Auth:       "Not configured",
		},
	}

	// Check if auth is configured
	if _, err := os.Stat(filepath.Join(xdg.DataHome, cliName, "credentials")); err == nil {
		info.Config.Auth = "Configured"
	}

	return info
}

// checkAPIHealth checks if the API is reachable.
func checkAPIHealth(opts *InfoOptions) StatusDetails {
	status := StatusDetails{
		LastChecked: time.Now(),
	}

	// Check spec cache first
	cacheStatus := checkSpecCache(opts)
	status.SpecCached = cacheStatus.SpecCached
	status.SpecAge = cacheStatus.SpecAge

	// Try to reach the API
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", opts.Config.API.BaseURL, nil)
	if err != nil {
		return status
	}

	resp, err := opts.HTTPClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		status.APIReachable = resp.StatusCode < 500
	}

	return status
}

// checkSpecCache checks the spec cache status.
func checkSpecCache(opts *InfoOptions) StatusDetails {
	status := StatusDetails{
		LastChecked: time.Now(),
	}

	cliName := opts.Config.Metadata.Name
	specCache, err := cache.NewSpecCache(cliName)
	if err != nil {
		return status
	}

	// Check if spec is cached
	ctx := context.Background()
	cached, err := specCache.Get(ctx, opts.Config.API.OpenAPIURL)
	if err == nil && cached != nil {
		status.SpecCached = true
		age := time.Since(cached.FetchedAt)
		status.SpecAge = formatDuration(age)
	}

	return status
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// formatInfoText formats info as human-readable text.
func formatInfoText(info *CLIInfo, w io.Writer) error {
	fmt.Fprintf(w, "\n%s v%s\n", info.CLI.Name, info.CLI.Version)
	if info.CLI.Description != "" {
		fmt.Fprintf(w, "%s\n", info.CLI.Description)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "API Information:")
	if info.API.Title != "" {
		fmt.Fprintf(w, "  Title: %s\n", info.API.Title)
	}
	if info.API.Version != "" {
		fmt.Fprintf(w, "  Version: %s\n", info.API.Version)
	}
	fmt.Fprintf(w, "  Base URL: %s\n", info.API.BaseURL)
	fmt.Fprintf(w, "  Spec URL: %s\n", info.API.SpecURL)
	if info.API.Endpoints > 0 {
		fmt.Fprintf(w, "  Endpoints: %d\n", info.API.Endpoints)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Configuration:")
	fmt.Fprintf(w, "  Config file: %s\n", info.Config.ConfigFile)
	fmt.Fprintf(w, "  Cache dir: %s\n", info.Config.CacheDir)
	fmt.Fprintf(w, "  Data dir: %s\n", info.Config.DataDir)
	fmt.Fprintf(w, "  State dir: %s\n", info.Config.StateDir)
	fmt.Fprintf(w, "  Auth: %s\n", info.Config.Auth)
	fmt.Fprintln(w)

	// Show status if available
	if info.Status.LastChecked.IsZero() {
		return nil
	}

	fmt.Fprintln(w, "Status:")
	if info.Status.APIReachable {
		fmt.Fprintln(w, "  ✓ API reachable")
	} else {
		fmt.Fprintln(w, "  ✗ API unreachable")
	}

	if info.Status.SpecCached {
		fmt.Fprintf(w, "  ✓ Spec cached (age: %s)\n", info.Status.SpecAge)
	} else {
		fmt.Fprintln(w, "  ✗ Spec not cached")
	}

	return nil
}

// formatInfoJSON formats info as JSON.
func formatInfoJSON(info *CLIInfo, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

// formatInfoYAML formats info as YAML.
func formatInfoYAML(info *CLIInfo, w io.Writer) error {
	fmt.Fprintln(w, "cli:")
	fmt.Fprintf(w, "  name: %s\n", info.CLI.Name)
	fmt.Fprintf(w, "  version: %s\n", info.CLI.Version)
	if info.CLI.Description != "" {
		fmt.Fprintf(w, "  description: %s\n", info.CLI.Description)
	}
	if info.CLI.Homepage != "" {
		fmt.Fprintf(w, "  homepage: %s\n", info.CLI.Homepage)
	}

	fmt.Fprintln(w, "api:")
	if info.API.Title != "" {
		fmt.Fprintf(w, "  title: %s\n", info.API.Title)
	}
	if info.API.Version != "" {
		fmt.Fprintf(w, "  version: %s\n", info.API.Version)
	}
	fmt.Fprintf(w, "  base_url: %s\n", info.API.BaseURL)
	fmt.Fprintf(w, "  spec_url: %s\n", info.API.SpecURL)

	fmt.Fprintln(w, "config:")
	fmt.Fprintf(w, "  config_file: %s\n", info.Config.ConfigFile)
	fmt.Fprintf(w, "  cache_dir: %s\n", info.Config.CacheDir)
	fmt.Fprintf(w, "  data_dir: %s\n", info.Config.DataDir)
	fmt.Fprintf(w, "  state_dir: %s\n", info.Config.StateDir)
	fmt.Fprintf(w, "  auth: %s\n", info.Config.Auth)

	if !info.Status.LastChecked.IsZero() {
		fmt.Fprintln(w, "status:")
		fmt.Fprintf(w, "  api_reachable: %t\n", info.Status.APIReachable)
		fmt.Fprintf(w, "  spec_cached: %t\n", info.Status.SpecCached)
		if info.Status.SpecAge != "" {
			fmt.Fprintf(w, "  spec_age: %s\n", info.Status.SpecAge)
		}
	}

	return nil
}
