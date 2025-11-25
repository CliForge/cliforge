package state

import (
	"sync"
	"time"
)

// Recent manages recently used values for autocomplete and smart defaults.
type Recent struct {
	Lists      map[string]*RecentList `yaml:"lists,omitempty" json:"lists,omitempty"`
	MaxPerList int                    `yaml:"max_per_list,omitempty" json:"max_per_list,omitempty"`
	mu         sync.RWMutex           `yaml:"-" json:"-"`
}

// RecentList represents a list of recent values for a specific category.
type RecentList struct {
	Name    string        `yaml:"name" json:"name"`
	Entries []*RecentItem `yaml:"entries" json:"entries"`
	Max     int           `yaml:"max,omitempty" json:"max,omitempty"`
}

// RecentItem represents a single recent value with metadata.
type RecentItem struct {
	Value    string    `yaml:"value" json:"value"`
	LastUsed time.Time `yaml:"last_used" json:"last_used"`
	UseCount int       `yaml:"use_count" json:"use_count"`
}

const (
	// DefaultMaxRecentEntries is the default maximum recent entries per list.
	DefaultMaxRecentEntries = 10
)

// NewRecent creates a new recent values manager.
func NewRecent() *Recent {
	return &Recent{
		Lists:      make(map[string]*RecentList),
		MaxPerList: DefaultMaxRecentEntries,
	}
}

// NewRecentWithMax creates a new recent values manager with custom max entries.
func NewRecentWithMax(maxPerList int) *Recent {
	if maxPerList <= 0 {
		maxPerList = DefaultMaxRecentEntries
	}
	return &Recent{
		Lists:      make(map[string]*RecentList),
		MaxPerList: maxPerList,
	}
}

// Add adds a value to a recent list.
// If the value already exists, it updates the last used time and count.
// Values are kept in most-recently-used order.
func (r *Recent) Add(listName, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create list
	list, exists := r.Lists[listName]
	if !exists {
		list = &RecentList{
			Name:    listName,
			Entries: make([]*RecentItem, 0),
			Max:     r.MaxPerList,
		}
		r.Lists[listName] = list
	}

	// Check if value already exists
	for i, item := range list.Entries {
		if item.Value == value {
			// Update existing item
			item.LastUsed = time.Now()
			item.UseCount++
			// Move to front (most recent)
			if i > 0 {
				list.Entries = append([]*RecentItem{item}, append(list.Entries[:i], list.Entries[i+1:]...)...)
			}
			return
		}
	}

	// Add new item at the front
	newItem := &RecentItem{
		Value:    value,
		LastUsed: time.Now(),
		UseCount: 1,
	}
	list.Entries = append([]*RecentItem{newItem}, list.Entries...)

	// Trim if exceeds max
	if len(list.Entries) > list.Max {
		list.Entries = list.Entries[:list.Max]
	}
}

// Get returns recent values for a list.
func (r *Recent) Get(listName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return []string{}
	}

	values := make([]string, len(list.Entries))
	for i, item := range list.Entries {
		values[i] = item.Value
	}
	return values
}

// GetWithMetadata returns recent items with full metadata.
func (r *Recent) GetWithMetadata(listName string) []*RecentItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return []*RecentItem{}
	}

	// Return a copy to prevent external modifications
	items := make([]*RecentItem, len(list.Entries))
	for i, item := range list.Entries {
		items[i] = &RecentItem{
			Value:    item.Value,
			LastUsed: item.LastUsed,
			UseCount: item.UseCount,
		}
	}
	return items
}

// GetTop returns the top N recent values for a list.
func (r *Recent) GetTop(listName string, n int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return []string{}
	}

	if n <= 0 || n > len(list.Entries) {
		n = len(list.Entries)
	}

	values := make([]string, n)
	for i := 0; i < n; i++ {
		values[i] = list.Entries[i].Value
	}
	return values
}

// Remove removes a value from a recent list.
func (r *Recent) Remove(listName, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list, exists := r.Lists[listName]
	if !exists {
		return
	}

	for i, item := range list.Entries {
		if item.Value == value {
			list.Entries = append(list.Entries[:i], list.Entries[i+1:]...)
			return
		}
	}
}

// Clear clears all values from a recent list.
func (r *Recent) Clear(listName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if list, exists := r.Lists[listName]; exists {
		list.Entries = make([]*RecentItem, 0)
	}
}

// ClearAll clears all recent lists.
func (r *Recent) ClearAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, list := range r.Lists {
		list.Entries = make([]*RecentItem, 0)
	}
}

// ListNames returns all list names.
func (r *Recent) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.Lists))
	for name := range r.Lists {
		names = append(names, name)
	}
	return names
}

// GetList returns a complete list with metadata.
func (r *Recent) GetList(listName string) (*RecentList, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return nil, false
	}

	// Return a copy
	listCopy := &RecentList{
		Name:    list.Name,
		Entries: make([]*RecentItem, len(list.Entries)),
		Max:     list.Max,
	}
	for i, item := range list.Entries {
		listCopy.Entries[i] = &RecentItem{
			Value:    item.Value,
			LastUsed: item.LastUsed,
			UseCount: item.UseCount,
		}
	}

	return listCopy, true
}

// SetMax sets the maximum entries for a specific list.
func (r *Recent) SetMax(listName string, max int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if max <= 0 {
		max = DefaultMaxRecentEntries
	}

	list, exists := r.Lists[listName]
	if !exists {
		list = &RecentList{
			Name:    listName,
			Entries: make([]*RecentItem, 0),
			Max:     max,
		}
		r.Lists[listName] = list
		return
	}

	list.Max = max

	// Trim if needed
	if len(list.Entries) > max {
		list.Entries = list.Entries[:max]
	}
}

// GetMostUsed returns values sorted by use count (descending).
func (r *Recent) GetMostUsed(listName string, limit int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return []string{}
	}

	// Create a copy and sort by use count
	items := make([]*RecentItem, len(list.Entries))
	copy(items, list.Entries)

	// Simple bubble sort by use count (descending)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].UseCount > items[i].UseCount {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}

	values := make([]string, len(items))
	for i, item := range items {
		values[i] = item.Value
	}
	return values
}

// GetByPattern returns values matching a pattern (simple substring match).
func (r *Recent) GetByPattern(listName, pattern string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return []string{}
	}

	matches := make([]string, 0)
	for _, item := range list.Entries {
		if matchesPattern(item.Value, pattern) {
			matches = append(matches, item.Value)
		}
	}
	return matches
}

// matchesPattern checks if a value matches a pattern (simple substring).
func matchesPattern(value, pattern string) bool {
	// Simple case-insensitive substring match
	// Could be enhanced with regex or glob patterns
	if pattern == "" {
		return true
	}

	// Convert to lowercase for case-insensitive matching
	valueLower := toLower(value)
	patternLower := toLower(pattern)

	return contains(valueLower, patternLower)
}

// toLower is a simple lowercase converter.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Prune removes entries older than a specified duration.
func (r *Recent) Prune(listName string, maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list, exists := r.Lists[listName]
	if !exists {
		return
	}

	cutoff := time.Now().Add(-maxAge)
	newEntries := make([]*RecentItem, 0)

	for _, item := range list.Entries {
		if item.LastUsed.After(cutoff) {
			newEntries = append(newEntries, item)
		}
	}

	list.Entries = newEntries
}

// PruneAll removes old entries from all lists.
func (r *Recent) PruneAll(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for _, list := range r.Lists {
		newEntries := make([]*RecentItem, 0)
		for _, item := range list.Entries {
			if item.LastUsed.After(cutoff) {
				newEntries = append(newEntries, item)
			}
		}
		list.Entries = newEntries
	}
}

// GetStats returns statistics about a recent list.
func (r *Recent) GetStats(listName string) *RecentStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, exists := r.Lists[listName]
	if !exists {
		return &RecentStats{
			ListName: listName,
		}
	}

	stats := &RecentStats{
		ListName:    listName,
		TotalItems:  len(list.Entries),
		MaxCapacity: list.Max,
	}

	if len(list.Entries) == 0 {
		return stats
	}

	var totalUseCount int
	for _, item := range list.Entries {
		totalUseCount += item.UseCount

		if stats.MostRecentTime.IsZero() || item.LastUsed.After(stats.MostRecentTime) {
			stats.MostRecentTime = item.LastUsed
			stats.MostRecentValue = item.Value
		}

		if item.UseCount > stats.MostUsedCount {
			stats.MostUsedCount = item.UseCount
			stats.MostUsedValue = item.Value
		}
	}

	if stats.TotalItems > 0 {
		stats.AverageUseCount = float64(totalUseCount) / float64(stats.TotalItems)
	}

	return stats
}

// RecentStats represents statistics about a recent list.
type RecentStats struct {
	ListName        string    `json:"list_name"`
	TotalItems      int       `json:"total_items"`
	MaxCapacity     int       `json:"max_capacity"`
	AverageUseCount float64   `json:"average_use_count"`
	MostUsedValue   string    `json:"most_used_value,omitempty"`
	MostUsedCount   int       `json:"most_used_count,omitempty"`
	MostRecentValue string    `json:"most_recent_value,omitempty"`
	MostRecentTime  time.Time `json:"most_recent_time,omitempty"`
}
