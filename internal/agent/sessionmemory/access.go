package sessionmemory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type accessRecord struct {
	Count        int       `json:"count"`
	LastAccessed time.Time `json:"last_accessed"`
}

type accessStore struct {
	Repositories map[string]map[string]accessRecord `json:"repositories"`
}

func loadAccessStore(path string) (*accessStore, error) {
	store := &accessStore{Repositories: map[string]map[string]accessRecord{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}
	if store.Repositories == nil {
		store.Repositories = map[string]map[string]accessRecord{}
	}
	return store, nil
}

func saveAccessStore(path string, store *accessStore) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *Service) recordAccess(repositoryID, factID string) error {
	store, err := loadAccessStore(s.accessPath)
	if err != nil {
		return err
	}
	if store.Repositories[repositoryID] == nil {
		store.Repositories[repositoryID] = map[string]accessRecord{}
	}
	rec := store.Repositories[repositoryID][factID]
	rec.Count++
	rec.LastAccessed = time.Now().UTC()
	store.Repositories[repositoryID][factID] = rec
	return saveAccessStore(s.accessPath, store)
}

func (s *Service) accessRanking(repositoryID string, factIDs []string) []string {
	store, err := loadAccessStore(s.accessPath)
	if err != nil {
		return factIDs
	}
	repo := store.Repositories[repositoryID]
	type ranked struct {
		id   string
		time time.Time
		cnt  int
	}
	items := make([]ranked, 0, len(factIDs))
	for _, id := range factIDs {
		rec := repo[id]
		items = append(items, ranked{id: id, time: rec.LastAccessed, cnt: rec.Count})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].time.Equal(items[j].time) {
			return items[i].time.After(items[j].time)
		}
		if items[i].cnt != items[j].cnt {
			return items[i].cnt > items[j].cnt
		}
		return items[i].id < items[j].id
	})
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.id
	}
	return out
}

func mergeAccessRanking(path, repositoryID string, ranking []string) error {
	store, err := loadAccessStore(path)
	if err != nil {
		return err
	}
	if store.Repositories[repositoryID] == nil {
		store.Repositories[repositoryID] = map[string]accessRecord{}
	}
	now := time.Now().UTC()
	for _, id := range ranking {
		rec := store.Repositories[repositoryID][id]
		rec.Count++
		rec.LastAccessed = now
		store.Repositories[repositoryID][id] = rec
	}
	return saveAccessStore(path, store)
}
