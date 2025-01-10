package store

import "sync"

const MemTableMaxSize = 1000

type MemTable struct {
	entries map[string]string
	size int
	mu sync.RWMutex
}

func NewMemTable() *MemTable {
	return &MemTable{
		entries: make(map[string]string),
	}
}

func (m *MemTable) Set(key, value string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.entries[key] = value
    m.size++
    return nil
}

func (m *MemTable) Get(key string) (string, bool){
    m.mu.RLock()
    defer m.mu.RUnlock()

    value, exists := m.entries[key]
    return value, exists
}

func (m *MemTable) Delete(key string){
    m.mu.Lock()
    defer m.mu.Unlock()

    delete(m.entries, key)
}

func (m *MemTable) Size() int {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.size
}

func (m *MemTable) Entries() map[string]string {
    m.mu.RLock()
    defer m.mu.RUnlock()

    entries := make(map[string]string, len(m.entries))
    for k, v := range m.entries {
        entries[k] = v
    }

    return entries
}

func (m *MemTable) Clear() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.entries = make(map[string]string)
    m.size = 0
}
