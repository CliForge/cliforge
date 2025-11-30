package main

import "time"

// Cluster represents a ROSA cluster.
type Cluster struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	State           string            `json:"state"`
	Region          string            `json:"region"`
	MultiAZ         bool              `json:"multi_az"`
	Version         string            `json:"version"`
	CloudProvider   string            `json:"cloud_provider"`
	ConsoleURL      string            `json:"console_url,omitempty"`
	APIURL          string            `json:"api_url,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	ExpirationTime  *time.Time        `json:"expiration_time,omitempty"`
	Properties      map[string]string `json:"properties,omitempty"`
	MachinePools    []string          `json:"machine_pools,omitempty"`
	IdentityProviders []string        `json:"identity_providers,omitempty"`
	Addons          []string          `json:"addons,omitempty"`
}

// MachinePool represents a machine pool in a cluster.
type MachinePool struct {
	ID              string            `json:"id"`
	ClusterID       string            `json:"cluster_id"`
	InstanceType    string            `json:"instance_type"`
	Replicas        int               `json:"replicas"`
	AutoscalingEnabled bool           `json:"autoscaling_enabled"`
	MinReplicas     int               `json:"min_replicas,omitempty"`
	MaxReplicas     int               `json:"max_replicas,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Taints          []Taint           `json:"taints,omitempty"`
	AvailabilityZone string           `json:"availability_zone,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Taint represents a node taint.
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// IdentityProvider represents an identity provider configuration.
type IdentityProvider struct {
	ID        string                 `json:"id"`
	ClusterID string                 `json:"cluster_id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	MappingMethod string              `json:"mapping_method"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// AddonInstallation represents an installed addon.
type AddonInstallation struct {
	ID        string                 `json:"id"`
	ClusterID string                 `json:"cluster_id"`
	AddonID   string                 `json:"addon_id"`
	State     string                 `json:"state"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// UpgradePolicy represents a cluster upgrade policy.
type UpgradePolicy struct {
	ID                string     `json:"id"`
	ClusterID         string     `json:"cluster_id"`
	Version           string     `json:"version"`
	ScheduleType      string     `json:"schedule_type"`
	NextRun           *time.Time `json:"next_run,omitempty"`
	State             string     `json:"state"`
	EnableMinorVersionUpgrades bool `json:"enable_minor_version_upgrades"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ListResponse wraps paginated list responses.
type ListResponse struct {
	Items []interface{} `json:"items"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
	Total int           `json:"total"`
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
	Code        int    `json:"code,omitempty"`
}

// TokenResponse represents an OAuth2 token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Metadata represents API metadata.
type Metadata struct {
	Version     string `json:"version"`
	ServerTime  string `json:"server_time"`
	DocsURL     string `json:"docs_url"`
}
