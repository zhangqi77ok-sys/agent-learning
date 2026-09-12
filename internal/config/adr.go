package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ADRStore 把架构决策记在 ~/.tiancode/adr.json，按 AST 节点 ID 索引。
type ADRStore struct {
	mu       sync.RWMutex
	filePath string
	notes    map[string]string
}

func NewADRStore(filePath string) *ADRStore {
	if filePath == "" {
		filePath = filepath.Join(UserDataDir(), "adr.json")
	}
	s := &ADRStore{filePath: filePath, notes: map[string]string{}}
	_ = s.load()
	return s
}

func DefaultADRStore() *ADRStore {
	return NewADRStore("")
}

func (s *ADRStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		s.notes = map[string]string{}
		return err
	}
	var notes map[string]string
	if err := json.Unmarshal(data, &notes); err != nil {
		s.notes = map[string]string{}
		return err
	}
	if notes == nil {
		notes = map[string]string{}
	}
	s.notes = notes
	return nil
}

func (s *ADRStore) Get(nodeID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.notes[nodeID]
}

func (s *ADRStore) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.notes))
	for k, v := range s.notes {
		out[k] = v
	}
	return out
}

func (s *ADRStore) Save(nodeID, note string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.notes == nil {
		s.notes = map[string]string{}
	}
	if note == "" {
		delete(s.notes, nodeID)
	} else {
		s.notes[nodeID] = note
	}
	data, err := json.MarshalIndent(s.notes, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteConfig(s.filePath, data)
}
