package store

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

func (s *Store) AppendJSON(kind string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.root, kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "journal")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(data, '\n'))
	return err
}

func (s *Store) ReadJournal(kind string) ([][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(filepath.Join(s.root, kind, "journal"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rows [][]byte
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		rows = append(rows, append([]byte(nil), line...))
	}
	return rows, nil
}

func (s *Store) ReplaceJournal(kind string, rows [][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.root, kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var buffer bytes.Buffer
	for _, row := range rows {
		buffer.Write(row)
		buffer.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(dir, "journal"), buffer.Bytes(), 0o644)
}
