package repository

import (
	"sync"

	"github.com/logitrack/core/internal/model"
)

type BranchRepository interface {
	List() []model.Branch
	Add(branch model.Branch)
	GetByID(id string) (model.Branch, bool)
	GetByCity(city string) (model.Branch, bool)
}

type inMemoryBranchRepository struct {
	mu       sync.RWMutex
	branches []model.Branch
}

func NewInMemoryBranchRepository() BranchRepository {
	return &inMemoryBranchRepository{}
}

func (r *inMemoryBranchRepository) List() []model.Branch {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Branch, len(r.branches))
	copy(result, r.branches)
	return result
}

func (r *inMemoryBranchRepository) Add(branch model.Branch) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.branches = append(r.branches, branch)
}

func (r *inMemoryBranchRepository) GetByID(id string) (model.Branch, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.branches {
		if b.ID == id {
			return b, true
		}
	}
	return model.Branch{}, false
}

func (r *inMemoryBranchRepository) GetByCity(city string) (model.Branch, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.branches {
		if b.City == city {
			return b, true
		}
	}
	return model.Branch{}, false
}
