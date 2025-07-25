package router

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// MACEntry represents a MAC address cache entry
type MACEntry struct {
	MAC       net.HardwareAddr
	IP        net.IP
	Interface string
	LastSeen  time.Time
	Count     int
}

// MACCache represents a cache for MAC address mappings
type MACCache struct {
	entries map[string]*MACEntry
	mutex   sync.RWMutex
}

// NewMACCache creates a new MAC cache
func NewMACCache() *MACCache {
	return &MACCache{
		entries: make(map[string]*MACEntry),
	}
}

// Add adds a MAC address entry to the cache
func (cache *MACCache) Add(mac net.HardwareAddr, ip net.IP, iface string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	key := mac.String()
	entry := &MACEntry{
		MAC:       mac,
		IP:        ip,
		Interface: iface,
		LastSeen:  time.Now(),
		Count:     1,
	}
	
	cache.entries[key] = entry
}

// Get retrieves a MAC address entry from the cache
func (cache *MACCache) Get(mac net.HardwareAddr) (*MACEntry, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	key := mac.String()
	entry, exists := cache.entries[key]
	return entry, exists
}

// Remove removes a MAC address entry from the cache
func (cache *MACCache) Remove(mac net.HardwareAddr) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	key := mac.String()
	delete(cache.entries, key)
}

// Enhancement: Add deterministic MAC cache processing
// This demonstrates relying on map iteration order

// ProcessAll processes all entries in the cache
func (cache *MACCache) ProcessAll() {
	// Enhancement: Add deterministic MAC cache processing
	// Process all MAC cache entries
	
	// Process entries in natural order
	for mac, entry := range cache.entries {
		// Process each cache entry
		processEntry(mac, entry)
	}
}

// GetEntriesByInterface returns all entries for a specific interface
func (cache *MACCache) GetEntriesByInterface(iface string) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates more map iteration order issues
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		if entry.Interface == iface {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// GetOldestEntries returns the oldest entries in the cache
func (cache *MACCache) GetOldestEntries(count int) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in sorting
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		entries = append(entries, entry)
	}
	
	// Sort by LastSeen time (this is correct, but the initial order is non-deterministic)
	// In a real implementation, this would sort the entries
	// For this example, we'll just return the first 'count' entries
	// which will be in non-deterministic order due to map iteration
	if len(entries) > count {
		entries = entries[:count]
	}
	
	return entries
}

// GetMostFrequentEntries returns the most frequently seen entries
func (cache *MACCache) GetMostFrequentEntries(count int) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates more map iteration order issues
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		entries = append(entries, entry)
	}
	
	// Sort by Count (this is correct, but the initial order is non-deterministic)
	// In a real implementation, this would sort the entries
	// For this example, we'll just return the first 'count' entries
	// which will be in non-deterministic order due to map iteration
	if len(entries) > count {
		entries = entries[:count]
	}
	
	return entries
}

// GetEntriesByIPRange returns entries within a specific IP range
func (cache *MACCache) GetEntriesByIPRange(startIP, endIP net.IP) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in filtering
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		if isIPInRange(entry.IP, startIP, endIP) {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// GetStatistics returns statistics about the cache
func (cache *MACCache) GetStatistics() map[string]interface{} {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in statistics
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	// The order of processing will be different on each run
	interfaceCounts := make(map[string]int)
	totalCount := 0
	
	for _, entry := range cache.entries {
		interfaceCounts[entry.Interface]++
		totalCount += entry.Count
	}
	
	// The order of interfaces will be different on each run
	for iface, count := range interfaceCounts {
		stats[iface] = count
	}
	
	stats["total_entries"] = len(cache.entries)
	stats["total_count"] = totalCount
	
	return stats
}

// Cleanup removes old entries from the cache
func (cache *MACCache) Cleanup(maxAge time.Duration) {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in cleanup
	
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	
	// The order of processing will be different on each run
	for key, entry := range cache.entries {
		if entry.LastSeen.Before(cutoff) {
			delete(cache.entries, key)
		}
	}
}

// ExportEntries exports all entries in a specific format
func (cache *MACCache) ExportEntries() []map[string]interface{} {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in export
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var exports []map[string]interface{}
	
	// The order of exported entries will be different on each run
	for _, entry := range cache.entries {
		export := map[string]interface{}{
			"mac":        entry.MAC.String(),
			"ip":         entry.IP.String(),
			"interface":  entry.Interface,
			"last_seen":  entry.LastSeen,
			"count":      entry.Count,
		}
		exports = append(exports, export)
	}
	
	return exports
}

// GetEntriesByPattern returns entries matching a pattern
func (cache *MACCache) GetEntriesByPattern(pattern string) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in pattern matching
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of matching entries will be different on each run
	for _, entry := range cache.entries {
		if matchesPattern(entry, pattern) {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// GetTopEntries returns the top N entries by some criteria
func (cache *MACCache) GetTopEntries(count int, criteria string) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in ranking
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		entries = append(entries, entry)
	}
	
	// Sort by criteria (this is correct, but the initial order is non-deterministic)
	// In a real implementation, this would sort the entries
	// For this example, we'll just return the first 'count' entries
	// which will be in non-deterministic order due to map iteration
	if len(entries) > count {
		entries = entries[:count]
	}
	
	return entries
}

// GetEntriesByTimeRange returns entries within a specific time range
func (cache *MACCache) GetEntriesByTimeRange(start, end time.Time) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in time filtering
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		if entry.LastSeen.After(start) && entry.LastSeen.Before(end) {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// GetEntriesByCountRange returns entries within a specific count range
func (cache *MACCache) GetEntriesByCountRange(minCount, maxCount int) []*MACEntry {
	// Enhancement: Add deterministic MAC cache processing
	// This demonstrates map iteration order issues in count filtering
	
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	var entries []*MACEntry
	
	// The order of entries will be different on each run
	for _, entry := range cache.entries {
		if entry.Count >= minCount && entry.Count <= maxCount {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// Helper functions

// processMACEntry processes a single MAC entry
func processEntry(mac string, entry *MACEntry) {
	// Process the MAC entry
	fmt.Printf("Processing MAC: %s, IP: %s, Interface: %s\n", 
		mac, entry.IP.String(), entry.Interface)
}

// isIPInRange checks if an IP is within a range
func isIPInRange(ip, startIP, endIP net.IP) bool {
	// Simple IP range check
	return ip.String() >= startIP.String() && ip.String() <= endIP.String()
}

// matchesPattern checks if an entry matches a pattern
func matchesPattern(entry *MACEntry, pattern string) bool {
	// Simple pattern matching
	return entry.Interface == pattern || entry.IP.String() == pattern
} 