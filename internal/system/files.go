package system

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Name     string
	FullPath string
	IsDir    bool
	Size     int64
}

type PresetPath struct {
	Title string
	Path  string
}

func GetPresetPaths() []PresetPath {
	homeDir, _ := os.UserHomeDir()

	return []PresetPath{
		{Title: "Systemd units (/etc/systemd/system)", Path: "/etc/systemd/system"},
		{Title: "Systemd units (/lib/systemd/system)", Path: "/lib/systemd/system"},
		{Title: "Nginx (/etc/nginx)", Path: "/etc/nginx"},
		{Title: "Docker (/etc/docker)", Path: "/etc/docker"},
		{Title: "Логи (/var/log)", Path: "/var/log"},
		{Title: "Домашняя папка", Path: homeDir},
	}
}

func sanitizePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	if strings.ContainsRune(path, 0) {
		return "", fmt.Errorf("path contains NUL byte")
	}

	// Require absolute paths. All navigation in the app is rooted at an absolute
	// preset path or the user's home, so this rejects ambiguous relative input
	// (and any residual ".." traversal is collapsed by Clean below).
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be absolute: %q", path)
	}

	return path, nil
}

func ListFiles(path string) ([]FileEntry, error) {
	path, err := sanitizePath(path)
	if err != nil {
		return nil, err
	}

	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]FileEntry, 0, len(items))

	for _, item := range items {
		info, err := item.Info()
		if err != nil {
			continue
		}

		result = append(result, FileEntry{
			Name:     item.Name(),
			FullPath: filepath.Join(path, item.Name()),
			IsDir:    item.IsDir(),
			Size:     info.Size(),
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

func ReadTextFile(path string, maxBytes int64) (string, error) {
	path, err := sanitizePath(path)
	if err != nil {
		return "", err
	}

	if maxBytes <= 0 {
		maxBytes = 1024 * 1024
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}

	if info.Size() > maxBytes {
		return "", fmt.Errorf("file is too large: %d bytes (limit %d)", info.Size(), maxBytes)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if !isTextContent(data) {
		return "", fmt.Errorf("file does not look like text")
	}

	return string(data), nil
}

func isTextContent(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	checkLen := len(data)
	if checkLen > 4096 {
		checkLen = 4096
	}

	for i := 0; i < checkLen; i++ {
		b := data[i]

		if b == 0 {
			return false
		}

		if b < 9 {
			return false
		}

		if b > 13 && b < 32 {
			return false
		}
	}

	return true
}

func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func WriteTextFile(path string, content string) error {
	path, err := sanitizePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	return os.WriteFile(path, []byte(content), info.Mode().Perm())
}

func CreateFile(path string) error {
	path, err := sanitizePath(path)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}

func RenameFile(oldPath, newPath string) error {
	oldPath, err := sanitizePath(oldPath)
	if err != nil {
		return err
	}

	newPath, err = sanitizePath(newPath)
	if err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

func DeleteFile(path string) error {
	path, err := sanitizePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Deliberately not recursive: deleting a whole tree from a monitoring
		// tool is too destructive. Only empty directories can be removed.
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("directory is not empty or cannot be removed: %w", err)
		}
		return nil
	}

	return os.Remove(path)
}

func CreateDirectory(path string) error {
	path, err := sanitizePath(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0o755)
}