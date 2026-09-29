package tokenmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var _ TokenStorage = (*LocalStorage)(nil)

// LocalStorage persiste tokens em um arquivo JSON.
type LocalStorage struct {
	mu   sync.Mutex
	path string
}

func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

// Save grava os tokens no arquivo de forma atômica.
func (s *LocalStorage) Save(ctx context.Context, tokens StorageTokens) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tokens: %w", err)
	}

	// Garante que o diretório existe
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// ✅ Escrita atômica: grava em .tmp e renomeia
	// Se o processo morrer no meio, o arquivo original fica intacto.
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Load lê os tokens do arquivo.
func (s *LocalStorage) Load(ctx context.Context) (*StorageTokens, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoStoredTokens
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	var tokens *StorageTokens
	if err := json.Unmarshal(data, tokens); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return tokens, nil
}
