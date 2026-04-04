package service

import (
	"fmt"
	"strings"

	"github.com/logitrack/core/internal/model"
	"github.com/logitrack/core/internal/repository"
)

type BranchService struct {
	repo repository.BranchRepository
}

func NewBranchService(repo repository.BranchRepository) *BranchService {
	return &BranchService{repo: repo}
}

func (s *BranchService) List() []model.Branch {
	return s.repo.List()
}

func (s *BranchService) ListActive() []model.Branch {
	return s.repo.ListActive()
}

func (s *BranchService) Search(query string) []model.Branch {
	if strings.TrimSpace(query) == "" {
		return s.repo.List()
	}
	return s.repo.GetByNameOrID(query)
}

func (s *BranchService) Create(req model.CreateBranchRequest) (model.Branch, error) {
	if strings.TrimSpace(req.Name) == "" {
		return model.Branch{}, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Street) == "" {
		return model.Branch{}, fmt.Errorf("street is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return model.Branch{}, fmt.Errorf("city is required")
	}
	if strings.TrimSpace(req.Province) == "" {
		return model.Branch{}, fmt.Errorf("province is required")
	}
	if strings.TrimSpace(req.PostalCode) == "" {
		return model.Branch{}, fmt.Errorf("postal code is required")
	}
	if req.CapacityKg <= 0 {
		return model.Branch{}, fmt.Errorf("capacity must be greater than 0")
	}

	branch := model.Branch{
		Name: req.Name,
		Address: model.Address{
			Street:     req.Street,
			City:       req.City,
			Province:   req.Province,
			PostalCode: req.PostalCode,
		},
		Province:   req.Province,
		CapacityKg: req.CapacityKg,
		Status:     model.BranchStatusActive,
	}

	if err := s.repo.Create(branch); err != nil {
		if err == repository.ErrDuplicateBranchName {
			return model.Branch{}, fmt.Errorf("a branch with name '%s' already exists", req.Name)
		}
		return model.Branch{}, fmt.Errorf("failed to create branch: %w", err)
	}

	created := s.repo.GetByNameOrID(req.Name)
	if len(created) > 0 {
		return created[0], nil
	}
	return branch, nil
}

func (s *BranchService) Update(id string, req model.UpdateBranchRequest) (model.Branch, error) {
	branch, found := s.repo.GetByID(id)
	if !found {
		return model.Branch{}, fmt.Errorf("branch not found")
	}

	if branch.Status != model.BranchStatusActive {
		return model.Branch{}, fmt.Errorf("cannot edit a branch that is not active (current status: %s)", branch.Status)
	}

	if strings.TrimSpace(req.Name) == "" {
		return model.Branch{}, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Street) == "" {
		return model.Branch{}, fmt.Errorf("street is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return model.Branch{}, fmt.Errorf("city is required")
	}
	if strings.TrimSpace(req.Province) == "" {
		return model.Branch{}, fmt.Errorf("province is required")
	}
	if strings.TrimSpace(req.PostalCode) == "" {
		return model.Branch{}, fmt.Errorf("postal code is required")
	}
	if req.CapacityKg <= 0 {
		return model.Branch{}, fmt.Errorf("capacity must be greater than 0")
	}

	update := model.Branch{
		Name: req.Name,
		Address: model.Address{
			Street:     req.Street,
			City:       req.City,
			Province:   req.Province,
			PostalCode: req.PostalCode,
		},
		Province:   req.Province,
		CapacityKg: req.CapacityKg,
	}

	if err := s.repo.Update(id, update); err != nil {
		if err == repository.ErrDuplicateBranchName {
			return model.Branch{}, fmt.Errorf("a branch with name '%s' already exists", req.Name)
		}
		if repository.IsNotUpdatable(err) {
			return model.Branch{}, fmt.Errorf("cannot edit a branch that is not active")
		}
		return model.Branch{}, fmt.Errorf("failed to update branch: %w", err)
	}

	updated, _ := s.repo.GetByID(id)
	return updated, nil
}

func (s *BranchService) UpdateStatus(id string, req model.UpdateBranchStatusRequest, username string) (model.Branch, error) {
	_, found := s.repo.GetByID(id)
	if !found {
		return model.Branch{}, fmt.Errorf("branch not found")
	}

	validStatuses := map[model.BranchStatus]bool{
		model.BranchStatusActive:       true,
		model.BranchStatusInactive:     true,
		model.BranchStatusOutOfService: true,
	}
	if !validStatuses[req.Status] {
		return model.Branch{}, fmt.Errorf("invalid status: %s", req.Status)
	}

	if err := s.repo.UpdateStatus(id, req.Status, username); err != nil {
		return model.Branch{}, fmt.Errorf("failed to update status: %w", err)
	}

	updated, _ := s.repo.GetByID(id)
	return updated, nil
}

func (s *BranchService) IsBranchActive(branchID string) bool {
	b, found := s.repo.GetByID(branchID)
	return found && b.Status == model.BranchStatusActive
}
