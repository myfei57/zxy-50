package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var ErrNotFound = errors.New("store record not found")

type Store struct {
	root string
	mu   sync.RWMutex
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) Path(kind string, id string) string {
	return filepath.Join(s.root, kind, id+".json")
}

func (s *Store) Write(kind string, id string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.root, kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dir, id+".json")
	temporary := target + ".tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, target)
}

func (s *Store) Read(kind string, id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(s.Path(kind, id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func (s *Store) Delete(kind string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.Path(kind, id))
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	return err
}

func (s *Store) List(kind string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dir := filepath.Join(s.root, kind)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(ids)
	return ids, nil
}

func (s *Store) WriteJSON(kind string, id string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.Write(kind, id, data)
}

func (s *Store) ReadJSON(kind string, id string, value any) error {
	data, err := s.Read(kind, id)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
