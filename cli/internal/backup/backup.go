package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type BackupManager struct {
	backupDir  string
	maxBackups int
}

func NewBackupManager(backupDir string, maxBackups int) *BackupManager {
	return &BackupManager{
		backupDir:  backupDir,
		maxBackups: maxBackups,
	}
}

func (bm *BackupManager) Backup(sourcePath string) (string, error) {
	if err := os.MkdirAll(bm.backupDir, 0700); err != nil {
		return "", err
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s.bak", timestamp)
	backupPath := filepath.Join(bm.backupDir, backupName)

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", err
	}

	if err := bm.cleanup(); err != nil {
		return backupPath, err
	}

	return backupPath, nil
}

func (bm *BackupManager) cleanup() error {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return err
	}

	if len(entries) <= bm.maxBackups {
		return nil
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}

	var files []fileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			name:    entry.Name(),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	toDelete := len(files) - bm.maxBackups
	for i := 0; i < toDelete; i++ {
		path := filepath.Join(bm.backupDir, files[i].name)
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

func (bm *BackupManager) List() ([]string, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(backups)))
	return backups, nil
}

func (bm *BackupManager) Restore(backupName, targetPath string) error {
	backupPath := filepath.Join(bm.backupDir, backupName)
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, data, 0600)
}
