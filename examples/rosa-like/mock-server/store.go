package main

import (
	"fmt"
	"sync"
	"time"
)

// Store provides thread-safe in-memory storage for all resource types.
type Store struct {
	mu                sync.RWMutex
	clusters          map[string]*Cluster
	machinePools      map[string]*MachinePool
	identityProviders map[string]*IdentityProvider
	addonInstallations map[string]*AddonInstallation
	upgradePolicies   map[string]*UpgradePolicy

	// Index for efficient lookups
	machinePoolsByCluster map[string][]string
	idpsByCluster         map[string][]string
	addonsByCluster       map[string][]string
	upgradesByCluster     map[string][]string
}

// NewStore creates a new Store instance.
func NewStore() *Store {
	return &Store{
		clusters:               make(map[string]*Cluster),
		machinePools:           make(map[string]*MachinePool),
		identityProviders:      make(map[string]*IdentityProvider),
		addonInstallations:     make(map[string]*AddonInstallation),
		upgradePolicies:        make(map[string]*UpgradePolicy),
		machinePoolsByCluster:  make(map[string][]string),
		idpsByCluster:          make(map[string][]string),
		addonsByCluster:        make(map[string][]string),
		upgradesByCluster:      make(map[string][]string),
	}
}

// Cluster operations

func (s *Store) CreateCluster(cluster *Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[cluster.ID]; exists {
		return fmt.Errorf("cluster %s already exists", cluster.ID)
	}

	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	s.clusters[cluster.ID] = cluster
	return nil
}

func (s *Store) GetCluster(id string) (*Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cluster, exists := s.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", id)
	}

	// Return a copy to prevent external modification
	clusterCopy := *cluster
	return &clusterCopy, nil
}

func (s *Store) ListClusters() []*Cluster {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clusters := make([]*Cluster, 0, len(s.clusters))
	for _, cluster := range s.clusters {
		clusterCopy := *cluster
		clusters = append(clusters, &clusterCopy)
	}

	return clusters
}

func (s *Store) UpdateCluster(id string, updateFn func(*Cluster) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster, exists := s.clusters[id]
	if !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	if err := updateFn(cluster); err != nil {
		return err
	}

	cluster.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteCluster(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[id]; !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	// Clean up related resources
	delete(s.clusters, id)
	delete(s.machinePoolsByCluster, id)
	delete(s.idpsByCluster, id)
	delete(s.addonsByCluster, id)
	delete(s.upgradesByCluster, id)

	return nil
}

// MachinePool operations

func (s *Store) CreateMachinePool(pool *MachinePool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.machinePools[pool.ID]; exists {
		return fmt.Errorf("machine pool %s already exists", pool.ID)
	}

	// Verify cluster exists
	if _, exists := s.clusters[pool.ClusterID]; !exists {
		return fmt.Errorf("cluster %s not found", pool.ClusterID)
	}

	now := time.Now()
	pool.CreatedAt = now
	pool.UpdatedAt = now

	s.machinePools[pool.ID] = pool
	s.machinePoolsByCluster[pool.ClusterID] = append(s.machinePoolsByCluster[pool.ClusterID], pool.ID)

	return nil
}

func (s *Store) GetMachinePool(id string) (*MachinePool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pool, exists := s.machinePools[id]
	if !exists {
		return nil, fmt.Errorf("machine pool %s not found", id)
	}

	poolCopy := *pool
	return &poolCopy, nil
}

func (s *Store) ListMachinePools(clusterID string) []*MachinePool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	poolIDs := s.machinePoolsByCluster[clusterID]
	pools := make([]*MachinePool, 0, len(poolIDs))

	for _, poolID := range poolIDs {
		if pool, exists := s.machinePools[poolID]; exists {
			poolCopy := *pool
			pools = append(pools, &poolCopy)
		}
	}

	return pools
}

func (s *Store) UpdateMachinePool(id string, updateFn func(*MachinePool) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.machinePools[id]
	if !exists {
		return fmt.Errorf("machine pool %s not found", id)
	}

	if err := updateFn(pool); err != nil {
		return err
	}

	pool.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteMachinePool(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.machinePools[id]
	if !exists {
		return fmt.Errorf("machine pool %s not found", id)
	}

	// Remove from index
	clusterPools := s.machinePoolsByCluster[pool.ClusterID]
	for i, poolID := range clusterPools {
		if poolID == id {
			s.machinePoolsByCluster[pool.ClusterID] = append(clusterPools[:i], clusterPools[i+1:]...)
			break
		}
	}

	delete(s.machinePools, id)
	return nil
}

// IdentityProvider operations

func (s *Store) CreateIdentityProvider(idp *IdentityProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.identityProviders[idp.ID]; exists {
		return fmt.Errorf("identity provider %s already exists", idp.ID)
	}

	// Verify cluster exists
	if _, exists := s.clusters[idp.ClusterID]; !exists {
		return fmt.Errorf("cluster %s not found", idp.ClusterID)
	}

	now := time.Now()
	idp.CreatedAt = now
	idp.UpdatedAt = now

	s.identityProviders[idp.ID] = idp
	s.idpsByCluster[idp.ClusterID] = append(s.idpsByCluster[idp.ClusterID], idp.ID)

	return nil
}

func (s *Store) GetIdentityProvider(id string) (*IdentityProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idp, exists := s.identityProviders[id]
	if !exists {
		return nil, fmt.Errorf("identity provider %s not found", id)
	}

	idpCopy := *idp
	return &idpCopy, nil
}

func (s *Store) ListIdentityProviders(clusterID string) []*IdentityProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idpIDs := s.idpsByCluster[clusterID]
	idps := make([]*IdentityProvider, 0, len(idpIDs))

	for _, idpID := range idpIDs {
		if idp, exists := s.identityProviders[idpID]; exists {
			idpCopy := *idp
			idps = append(idps, &idpCopy)
		}
	}

	return idps
}

func (s *Store) UpdateIdentityProvider(id string, updateFn func(*IdentityProvider) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idp, exists := s.identityProviders[id]
	if !exists {
		return fmt.Errorf("identity provider %s not found", id)
	}

	if err := updateFn(idp); err != nil {
		return err
	}

	idp.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteIdentityProvider(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idp, exists := s.identityProviders[id]
	if !exists {
		return fmt.Errorf("identity provider %s not found", id)
	}

	// Remove from index
	clusterIDPs := s.idpsByCluster[idp.ClusterID]
	for i, idpID := range clusterIDPs {
		if idpID == id {
			s.idpsByCluster[idp.ClusterID] = append(clusterIDPs[:i], clusterIDPs[i+1:]...)
			break
		}
	}

	delete(s.identityProviders, id)
	return nil
}

// AddonInstallation operations

func (s *Store) CreateAddonInstallation(addon *AddonInstallation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.addonInstallations[addon.ID]; exists {
		return fmt.Errorf("addon installation %s already exists", addon.ID)
	}

	// Verify cluster exists
	if _, exists := s.clusters[addon.ClusterID]; !exists {
		return fmt.Errorf("cluster %s not found", addon.ClusterID)
	}

	now := time.Now()
	addon.CreatedAt = now
	addon.UpdatedAt = now

	s.addonInstallations[addon.ID] = addon
	s.addonsByCluster[addon.ClusterID] = append(s.addonsByCluster[addon.ClusterID], addon.ID)

	return nil
}

func (s *Store) GetAddonInstallation(id string) (*AddonInstallation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addon, exists := s.addonInstallations[id]
	if !exists {
		return nil, fmt.Errorf("addon installation %s not found", id)
	}

	addonCopy := *addon
	return &addonCopy, nil
}

func (s *Store) ListAddonInstallations(clusterID string) []*AddonInstallation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addonIDs := s.addonsByCluster[clusterID]
	addons := make([]*AddonInstallation, 0, len(addonIDs))

	for _, addonID := range addonIDs {
		if addon, exists := s.addonInstallations[addonID]; exists {
			addonCopy := *addon
			addons = append(addons, &addonCopy)
		}
	}

	return addons
}

func (s *Store) UpdateAddonInstallation(id string, updateFn func(*AddonInstallation) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	addon, exists := s.addonInstallations[id]
	if !exists {
		return fmt.Errorf("addon installation %s not found", id)
	}

	if err := updateFn(addon); err != nil {
		return err
	}

	addon.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteAddonInstallation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	addon, exists := s.addonInstallations[id]
	if !exists {
		return fmt.Errorf("addon installation %s not found", id)
	}

	// Remove from index
	clusterAddons := s.addonsByCluster[addon.ClusterID]
	for i, addonID := range clusterAddons {
		if addonID == id {
			s.addonsByCluster[addon.ClusterID] = append(clusterAddons[:i], clusterAddons[i+1:]...)
			break
		}
	}

	delete(s.addonInstallations, id)
	return nil
}

// UpgradePolicy operations

func (s *Store) CreateUpgradePolicy(policy *UpgradePolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.upgradePolicies[policy.ID]; exists {
		return fmt.Errorf("upgrade policy %s already exists", policy.ID)
	}

	// Verify cluster exists
	if _, exists := s.clusters[policy.ClusterID]; !exists {
		return fmt.Errorf("cluster %s not found", policy.ClusterID)
	}

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	s.upgradePolicies[policy.ID] = policy
	s.upgradesByCluster[policy.ClusterID] = append(s.upgradesByCluster[policy.ClusterID], policy.ID)

	return nil
}

func (s *Store) GetUpgradePolicy(id string) (*UpgradePolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.upgradePolicies[id]
	if !exists {
		return nil, fmt.Errorf("upgrade policy %s not found", id)
	}

	policyCopy := *policy
	return &policyCopy, nil
}

func (s *Store) ListUpgradePolicies(clusterID string) []*UpgradePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policyIDs := s.upgradesByCluster[clusterID]
	policies := make([]*UpgradePolicy, 0, len(policyIDs))

	for _, policyID := range policyIDs {
		if policy, exists := s.upgradePolicies[policyID]; exists {
			policyCopy := *policy
			policies = append(policies, &policyCopy)
		}
	}

	return policies
}

func (s *Store) UpdateUpgradePolicy(id string, updateFn func(*UpgradePolicy) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.upgradePolicies[id]
	if !exists {
		return fmt.Errorf("upgrade policy %s not found", id)
	}

	if err := updateFn(policy); err != nil {
		return err
	}

	policy.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeleteUpgradePolicy(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.upgradePolicies[id]
	if !exists {
		return fmt.Errorf("upgrade policy %s not found", id)
	}

	// Remove from index
	clusterPolicies := s.upgradesByCluster[policy.ClusterID]
	for i, policyID := range clusterPolicies {
		if policyID == id {
			s.upgradesByCluster[policy.ClusterID] = append(clusterPolicies[:i], clusterPolicies[i+1:]...)
			break
		}
	}

	delete(s.upgradePolicies, id)
	return nil
}
