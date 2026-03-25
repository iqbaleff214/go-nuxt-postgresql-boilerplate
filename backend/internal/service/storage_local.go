package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorageService stores files on the local filesystem.
type LocalStorageService struct {
	basePath    string
	baseURL     string
}

func NewLocalStorageService(basePath, baseURL string) *LocalStorageService {
	_ = os.MkdirAll(basePath, 0755)
	return &LocalStorageService{basePath: basePath, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *LocalStorageService) Upload(_ context.Context, file io.Reader, path string) (string, error) {
	dest := filepath.Join(s.basePath, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, file); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return s.baseURL + "/files/" + path, nil
}

func (s *LocalStorageService) Delete(_ context.Context, path string) error {
	return os.Remove(filepath.Join(s.basePath, filepath.FromSlash(path)))
}

func (s *LocalStorageService) GetSignedURL(_ context.Context, path string, _ time.Duration) (string, error) {
	return s.baseURL + "/files/" + path, nil
}
