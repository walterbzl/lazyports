package labels

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Store holds persistent port labels backed by ~/.config/lazyports/labels.json.
type Store struct {
	path string
	data map[string]string
}

// Load reads labels from the config directory. Returns an empty store if the
// file does not exist; returns the store with a logged error if the file is corrupt.
func Load() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	path := filepath.Join(dir, "lazyports", "labels.json")
	s := &Store{path: path, data: make(map[string]string)}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(raw, &s.data); err != nil {
		return s, nil // treat corrupt file as empty rather than fatal
	}
	return s, nil
}

// Get returns the label for port, or "" if none is set.
func (s *Store) Get(port string) string {
	return s.data[port]
}

// Set assigns label to port and persists. An empty label deletes the entry.
func (s *Store) Set(port, label string) error {
	if label == "" {
		delete(s.data, port)
	} else {
		s.data[port] = label
	}
	return s.save()
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}
