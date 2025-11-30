package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Server represents the mock API server.
type Server struct {
	store      *Store
	authServer *AuthServer
	mux        *http.ServeMux
	port       string
}

// NewServer creates a new Server instance.
func NewServer(port string) *Server {
	s := &Server{
		store:      NewStore(),
		authServer: NewAuthServer(),
		mux:        http.NewServeMux(),
		port:       port,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// OAuth2 endpoints
	s.mux.HandleFunc("/auth/token", s.authServer.HandleToken)

	// API metadata endpoint (public)
	s.mux.HandleFunc("/api/v1", s.handleMetadata)

	// Cluster endpoints (protected)
	s.mux.HandleFunc("/api/v1/clusters", s.authServer.AuthMiddleware(s.handleClusters))
	s.mux.HandleFunc("/api/v1/clusters/", s.authServer.AuthMiddleware(s.handleClusterResource))

	// Reference data endpoints (protected)
	s.mux.HandleFunc("/api/v1/versions", s.authServer.AuthMiddleware(s.handleVersions))
	s.mux.HandleFunc("/api/v1/regions", s.authServer.AuthMiddleware(s.handleRegions))

	// AWS verification endpoints (protected)
	s.mux.HandleFunc("/api/v1/aws/credentials/verify", s.authServer.AuthMiddleware(s.handleVerifyCredentials))
	s.mux.HandleFunc("/api/v1/aws/quotas/verify", s.authServer.AuthMiddleware(s.handleVerifyQuotas))

	// Health check (public)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// handleMetadata returns API metadata.
func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metadata := Metadata{
		Version:    "v1",
		ServerTime: time.Now().Format(time.RFC3339),
		DocsURL:    "https://api.openshift.com/docs",
	}

	s.respondJSON(w, http.StatusOK, metadata)
}

// handleHealth returns health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleClusters handles cluster list and create operations.
func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listClusters(w, r)
	case http.MethodPost:
		s.createCluster(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleClusterResource handles individual cluster operations.
func (s *Server) handleClusterResource(w http.ResponseWriter, r *http.Request) {
	// Extract cluster ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Cluster ID required", http.StatusBadRequest)
		return
	}

	clusterID := parts[0]

	// Handle nested resources
	if len(parts) > 1 {
		resource := parts[1]
		switch resource {
		case "machine_pools":
			s.handleMachinePools(w, r, clusterID, parts[2:])
		case "identity_providers":
			s.handleIdentityProviders(w, r, clusterID, parts[2:])
		case "addons":
			s.handleAddons(w, r, clusterID, parts[2:])
		case "upgrade_policies":
			s.handleUpgradePolicies(w, r, clusterID, parts[2:])
		default:
			http.Error(w, "Unknown resource", http.StatusNotFound)
		}
		return
	}

	// Handle cluster-level operations
	switch r.Method {
	case http.MethodGet:
		s.getCluster(w, r, clusterID)
	case http.MethodPatch:
		s.updateCluster(w, r, clusterID)
	case http.MethodDelete:
		s.deleteCluster(w, r, clusterID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listClusters returns all clusters.
func (s *Server) listClusters(w http.ResponseWriter, r *http.Request) {
	clusters := s.store.ListClusters()

	// Convert to interface slice for ListResponse
	items := make([]interface{}, len(clusters))
	for i, cluster := range clusters {
		items[i] = cluster
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// createCluster creates a new cluster.
func (s *Server) createCluster(w http.ResponseWriter, r *http.Request) {
	var cluster Cluster
	if err := json.NewDecoder(r.Body).Decode(&cluster); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	// Generate ID if not provided
	if cluster.ID == "" {
		cluster.ID = uuid.New().String()
	}

	// Set default state
	if cluster.State == "" {
		cluster.State = "pending"
	}

	// Validate required fields
	if cluster.Name == "" {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Cluster name is required")
		return
	}

	if cluster.Region == "" {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Region is required")
		return
	}

	// Set defaults
	if cluster.CloudProvider == "" {
		cluster.CloudProvider = "aws"
	}

	if cluster.Version == "" {
		cluster.Version = "4.14.0"
	}

	// Generate URLs
	cluster.ConsoleURL = fmt.Sprintf("https://console-%s.example.com", cluster.ID[:8])
	cluster.APIURL = fmt.Sprintf("https://api-%s.example.com:6443", cluster.ID[:8])

	// Create cluster
	if err := s.store.CreateCluster(&cluster); err != nil {
		s.respondError(w, http.StatusConflict, "conflict", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, cluster)
}

// getCluster returns a specific cluster.
func (s *Server) getCluster(w http.ResponseWriter, r *http.Request, clusterID string) {
	cluster, err := s.store.GetCluster(clusterID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, cluster)
}

// updateCluster updates a cluster.
func (s *Server) updateCluster(w http.ResponseWriter, r *http.Request, clusterID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	err := s.store.UpdateCluster(clusterID, func(cluster *Cluster) error {
		// Apply updates
		if name, ok := updates["name"].(string); ok {
			cluster.Name = name
		}
		if version, ok := updates["version"].(string); ok {
			cluster.Version = version
		}
		if properties, ok := updates["properties"].(map[string]interface{}); ok {
			if cluster.Properties == nil {
				cluster.Properties = make(map[string]string)
			}
			for k, v := range properties {
				if strVal, ok := v.(string); ok {
					cluster.Properties[k] = strVal
				}
			}
		}
		return nil
	})

	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	// Retrieve updated cluster
	cluster, _ := s.store.GetCluster(clusterID)
	s.respondJSON(w, http.StatusOK, cluster)
}

// deleteCluster deletes a cluster.
func (s *Server) deleteCluster(w http.ResponseWriter, r *http.Request, clusterID string) {
	if err := s.store.DeleteCluster(clusterID); err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleMachinePools handles machine pool operations.
func (s *Server) handleMachinePools(w http.ResponseWriter, r *http.Request, clusterID string, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		// List or create
		switch r.Method {
		case http.MethodGet:
			s.listMachinePools(w, r, clusterID)
		case http.MethodPost:
			s.createMachinePool(w, r, clusterID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	poolID := parts[0]

	// Individual machine pool operations
	switch r.Method {
	case http.MethodGet:
		s.getMachinePool(w, r, poolID)
	case http.MethodPatch:
		s.updateMachinePool(w, r, poolID)
	case http.MethodDelete:
		s.deleteMachinePool(w, r, poolID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listMachinePools lists machine pools for a cluster.
func (s *Server) listMachinePools(w http.ResponseWriter, r *http.Request, clusterID string) {
	pools := s.store.ListMachinePools(clusterID)

	items := make([]interface{}, len(pools))
	for i, pool := range pools {
		items[i] = pool
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// createMachinePool creates a machine pool.
func (s *Server) createMachinePool(w http.ResponseWriter, r *http.Request, clusterID string) {
	var pool MachinePool
	if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	pool.ClusterID = clusterID
	if pool.ID == "" {
		pool.ID = uuid.New().String()
	}

	if err := s.store.CreateMachinePool(&pool); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, pool)
}

// getMachinePool returns a specific machine pool.
func (s *Server) getMachinePool(w http.ResponseWriter, r *http.Request, poolID string) {
	pool, err := s.store.GetMachinePool(poolID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, pool)
}

// updateMachinePool updates a machine pool.
func (s *Server) updateMachinePool(w http.ResponseWriter, r *http.Request, poolID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	err := s.store.UpdateMachinePool(poolID, func(pool *MachinePool) error {
		if replicas, ok := updates["replicas"].(float64); ok {
			pool.Replicas = int(replicas)
		}
		if instanceType, ok := updates["instance_type"].(string); ok {
			pool.InstanceType = instanceType
		}
		return nil
	})

	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	pool, _ := s.store.GetMachinePool(poolID)
	s.respondJSON(w, http.StatusOK, pool)
}

// deleteMachinePool deletes a machine pool.
func (s *Server) deleteMachinePool(w http.ResponseWriter, r *http.Request, poolID string) {
	if err := s.store.DeleteMachinePool(poolID); err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleIdentityProviders handles identity provider operations.
func (s *Server) handleIdentityProviders(w http.ResponseWriter, r *http.Request, clusterID string, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listIdentityProviders(w, r, clusterID)
		case http.MethodPost:
			s.createIdentityProvider(w, r, clusterID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	idpID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getIdentityProvider(w, r, idpID)
	case http.MethodDelete:
		s.deleteIdentityProvider(w, r, idpID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listIdentityProviders lists identity providers for a cluster.
func (s *Server) listIdentityProviders(w http.ResponseWriter, r *http.Request, clusterID string) {
	idps := s.store.ListIdentityProviders(clusterID)

	items := make([]interface{}, len(idps))
	for i, idp := range idps {
		items[i] = idp
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// createIdentityProvider creates an identity provider.
func (s *Server) createIdentityProvider(w http.ResponseWriter, r *http.Request, clusterID string) {
	var idp IdentityProvider
	if err := json.NewDecoder(r.Body).Decode(&idp); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	idp.ClusterID = clusterID
	if idp.ID == "" {
		idp.ID = uuid.New().String()
	}

	if err := s.store.CreateIdentityProvider(&idp); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, idp)
}

// getIdentityProvider returns a specific identity provider.
func (s *Server) getIdentityProvider(w http.ResponseWriter, r *http.Request, idpID string) {
	idp, err := s.store.GetIdentityProvider(idpID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, idp)
}

// deleteIdentityProvider deletes an identity provider.
func (s *Server) deleteIdentityProvider(w http.ResponseWriter, r *http.Request, idpID string) {
	if err := s.store.DeleteIdentityProvider(idpID); err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleAddons handles addon operations.
func (s *Server) handleAddons(w http.ResponseWriter, r *http.Request, clusterID string, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listAddons(w, r, clusterID)
		case http.MethodPost:
			s.createAddon(w, r, clusterID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	addonID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getAddon(w, r, addonID)
	case http.MethodDelete:
		s.deleteAddon(w, r, addonID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listAddons lists addons for a cluster.
func (s *Server) listAddons(w http.ResponseWriter, r *http.Request, clusterID string) {
	addons := s.store.ListAddonInstallations(clusterID)

	items := make([]interface{}, len(addons))
	for i, addon := range addons {
		items[i] = addon
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// createAddon creates an addon installation.
func (s *Server) createAddon(w http.ResponseWriter, r *http.Request, clusterID string) {
	var addon AddonInstallation
	if err := json.NewDecoder(r.Body).Decode(&addon); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	addon.ClusterID = clusterID
	if addon.ID == "" {
		addon.ID = uuid.New().String()
	}
	if addon.State == "" {
		addon.State = "installing"
	}

	if err := s.store.CreateAddonInstallation(&addon); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, addon)
}

// getAddon returns a specific addon installation.
func (s *Server) getAddon(w http.ResponseWriter, r *http.Request, addonID string) {
	addon, err := s.store.GetAddonInstallation(addonID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, addon)
}

// deleteAddon deletes an addon installation.
func (s *Server) deleteAddon(w http.ResponseWriter, r *http.Request, addonID string) {
	if err := s.store.DeleteAddonInstallation(addonID); err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpgradePolicies handles upgrade policy operations.
func (s *Server) handleUpgradePolicies(w http.ResponseWriter, r *http.Request, clusterID string, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		switch r.Method {
		case http.MethodGet:
			s.listUpgradePolicies(w, r, clusterID)
		case http.MethodPost:
			s.createUpgradePolicy(w, r, clusterID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	policyID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getUpgradePolicy(w, r, policyID)
	case http.MethodDelete:
		s.deleteUpgradePolicy(w, r, policyID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listUpgradePolicies lists upgrade policies for a cluster.
func (s *Server) listUpgradePolicies(w http.ResponseWriter, r *http.Request, clusterID string) {
	policies := s.store.ListUpgradePolicies(clusterID)

	items := make([]interface{}, len(policies))
	for i, policy := range policies {
		items[i] = policy
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// createUpgradePolicy creates an upgrade policy.
func (s *Server) createUpgradePolicy(w http.ResponseWriter, r *http.Request, clusterID string) {
	var policy UpgradePolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	policy.ClusterID = clusterID
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	if policy.State == "" {
		policy.State = "scheduled"
	}

	if err := s.store.CreateUpgradePolicy(&policy); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, policy)
}

// getUpgradePolicy returns a specific upgrade policy.
func (s *Server) getUpgradePolicy(w http.ResponseWriter, r *http.Request, policyID string) {
	policy, err := s.store.GetUpgradePolicy(policyID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, policy)
}

// deleteUpgradePolicy deletes an upgrade policy.
func (s *Server) deleteUpgradePolicy(w http.ResponseWriter, r *http.Request, policyID string) {
	if err := s.store.DeleteUpgradePolicy(policyID); err != nil {
		s.respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON sends a JSON response.
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// respondError sends an error response.
func (s *Server) respondError(w http.ResponseWriter, status int, errorCode, description string) {
	s.respondJSON(w, status, ErrorResponse{
		Error:       errorCode,
		Description: description,
		Code:        status,
	})
}

// handleVersions returns available OpenShift versions.
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	versions := []map[string]interface{}{
		{"id": "openshift-v4.14.5", "raw_id": "4.14.5", "channel_group": "stable", "enabled": true, "default": true},
		{"id": "openshift-v4.14.4", "raw_id": "4.14.4", "channel_group": "stable", "enabled": true, "default": false},
		{"id": "openshift-v4.13.20", "raw_id": "4.13.20", "channel_group": "stable", "enabled": true, "default": false},
	}

	items := make([]interface{}, len(versions))
	for i, v := range versions {
		items[i] = v
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleRegions returns available AWS regions.
func (s *Server) handleRegions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	regions := []map[string]interface{}{
		{"id": "us-east-1", "display_name": "US East (N. Virginia)", "enabled": true, "supports_multi_az": true},
		{"id": "us-east-2", "display_name": "US East (Ohio)", "enabled": true, "supports_multi_az": true},
		{"id": "us-west-1", "display_name": "US West (N. California)", "enabled": true, "supports_multi_az": false},
		{"id": "us-west-2", "display_name": "US West (Oregon)", "enabled": true, "supports_multi_az": true},
		{"id": "eu-west-1", "display_name": "Europe (Ireland)", "enabled": true, "supports_multi_az": true},
		{"id": "eu-central-1", "display_name": "Europe (Frankfurt)", "enabled": true, "supports_multi_az": true},
		{"id": "ap-southeast-1", "display_name": "Asia Pacific (Singapore)", "enabled": true, "supports_multi_az": true},
	}

	items := make([]interface{}, len(regions))
	for i, v := range regions {
		items[i] = v
	}

	response := ListResponse{
		Items: items,
		Page:  1,
		Size:  len(items),
		Total: len(items),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleVerifyCredentials mocks AWS credential verification.
func (s *Server) handleVerifyCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock successful verification
	result := map[string]interface{}{
		"valid":      true,
		"account_id": "123456789012",
		"arn":        "arn:aws:iam::123456789012:user/rosa-admin",
	}

	s.respondJSON(w, http.StatusOK, result)
}

// handleVerifyQuotas mocks AWS quota verification.
func (s *Server) handleVerifyQuotas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock successful quota check
	result := map[string]interface{}{
		"sufficient": true,
		"quotas": map[string]interface{}{
			"vpc":             map[string]int{"limit": 5, "used": 2, "available": 3},
			"elastic_ip":      map[string]int{"limit": 5, "used": 0, "available": 5},
			"nat_gateway":     map[string]int{"limit": 5, "used": 1, "available": 4},
			"instance_m5":     map[string]int{"limit": 20, "used": 0, "available": 20},
		},
	}

	s.respondJSON(w, http.StatusOK, result)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	log.Printf("Starting mock API server on :%s", s.port)
	log.Printf("OAuth2 token endpoint: http://localhost:%s/auth/token", s.port)
	log.Printf("API endpoint: http://localhost:%s/api/v1", s.port)
	log.Printf("Health check: http://localhost:%s/health", s.port)
	return http.ListenAndServe(":"+s.port, s.mux)
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer(port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
