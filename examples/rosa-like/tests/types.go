package tests

import (
	"encoding/json"
	"time"
)

// Shared type definitions for all test files
// These match the API response structures

type Cluster struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	State      string    `json:"state"`
	Region     string    `json:"region"`
	MultiAZ    bool      `json:"multi_az"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MachinePool struct {
	ID                 string    `json:"id"`
	ClusterID          string    `json:"cluster_id"`
	InstanceType       string    `json:"instance_type"`
	Replicas           int       `json:"replicas"`
	AutoscalingEnabled bool      `json:"autoscaling_enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type IdentityProvider struct {
	ID            string                 `json:"id"`
	ClusterID     string                 `json:"cluster_id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	MappingMethod string                 `json:"mapping_method"`
	Config        map[string]interface{} `json:"config"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type AddonInstallation struct {
	ID         string                 `json:"id"`
	ClusterID  string                 `json:"cluster_id"`
	AddonID    string                 `json:"addon_id"`
	State      string                 `json:"state"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type ListResponse struct {
	Items []json.RawMessage `json:"items"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
	Total int               `json:"total"`
}

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
	Code        int    `json:"code,omitempty"`
}

// Request types for creating resources

type MachinePoolRequest struct {
	InstanceType       string `json:"instance_type"`
	Replicas           int    `json:"replicas"`
	AutoscalingEnabled bool   `json:"autoscaling_enabled"`
}

type IdentityProviderRequest struct {
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	MappingMethod string                 `json:"mapping_method"`
	Config        map[string]interface{} `json:"config"`
}
