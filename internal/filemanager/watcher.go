package filemanager

import (
	"os"
	"path/filepath"
	"strings"
)

// DirectoryScanner отвечает за сканирование директории с входящими файлами
type DirectoryScanner struct {
	incomingDir string
}

// NewDirectoryScanner создает новый экземпляр сканера директории
func NewDirectoryScanner(incomingDir string) *DirectoryScanner {
	return &DirectoryScanner{
		incomingDir: incomingDir,
	}
}

// ListMarkFiles возвращает список файлов OUT_MARK в директории
func (s *DirectoryScanner) ListMarkFiles() ([]string, error) {
	var result []string

	files, err := os.ReadDir(s.incomingDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Проверяем, что это файл OUT_MARK_*.xml
		if strings.HasPrefix(file.Name(), "OUT_MARK_") && strings.HasSuffix(file.Name(), ".xml") {
			result = append(result, filepath.Join(s.incomingDir, file.Name()))
		}
	}

	return result, nil
}
