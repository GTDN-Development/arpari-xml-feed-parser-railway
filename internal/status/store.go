package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	Success = "success"
	Failed  = "failed"
)

type FeedStatus struct {
	Filename       string `json:"filename"`
	LastRunAt      string `json:"lastRunAt"`
	Status         string `json:"status"`
	ItemsProcessed int    `json:"itemsProcessed"`
	ItemsSkipped   int    `json:"itemsSkipped"`
	Error          string `json:"error"`
}

type File struct {
	Feeds map[string]FeedStatus `json:"feeds"`
}

type Store struct {
	Path string
}

func NewStore(dataDir string) Store {
	return Store{
		Path: filepath.Join(dataDir, "status.json"),
	}
}

func Empty() File {
	return File{Feeds: map[string]FeedStatus{}}
}

func (store Store) Read() (File, error) {
	data, err := os.ReadFile(store.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Empty(), nil
		}
		return File{}, fmt.Errorf("read status: %w", err)
	}

	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("parse status: %w", err)
	}
	if file.Feeds == nil {
		file.Feeds = map[string]FeedStatus{}
	}

	return file, nil
}

func (store Store) Update(feedName string, feedStatus FeedStatus) error {
	file, err := store.Read()
	if err != nil {
		return err
	}

	file.Feeds[feedName] = feedStatus
	return store.Write(file)
}

func (store Store) Write(file File) error {
	if file.Feeds == nil {
		file.Feeds = map[string]FeedStatus{}
	}

	if err := os.MkdirAll(filepath.Dir(store.Path), 0o755); err != nil {
		return fmt.Errorf("create status dir: %w", err)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal status: %w", err)
	}
	data = append(data, '\n')

	temp, err := os.CreateTemp(filepath.Dir(store.Path), "status.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp status: %w", err)
	}

	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempName)
		}
	}()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temp status: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temp status: %w", err)
	}

	if err := os.Rename(tempName, store.Path); err != nil {
		return fmt.Errorf("publish status: %w", err)
	}

	committed = true
	return nil
}

func NewFeedStatus(filename string, status string, processed int, skipped int, message string, now time.Time) FeedStatus {
	return FeedStatus{
		Filename:       filename,
		LastRunAt:      now.UTC().Format(time.RFC3339),
		Status:         status,
		ItemsProcessed: processed,
		ItemsSkipped:   skipped,
		Error:          message,
	}
}
