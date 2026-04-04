package repository

import "github.com/logitrack/core/internal/model"

// MLConfigRepository manages ML configuration versions and trained model blobs.
type MLConfigRepository interface {
	// GetActive returns the currently active config, or nil if none exists.
	GetActive() (*model.MLConfig, error)
	// List returns all configs ordered by created_at DESC.
	List() ([]model.MLConfig, error)
	// Create inserts a new config row (is_active = false).
	Create(config model.MLConfig) (model.MLConfig, error)
	// Activate makes the given config active (deactivates all others) in a transaction.
	Activate(id int) error
	// SaveModel stores a trained model blob linked to a config.
	SaveModel(configID int, data []byte) error
	// GetActiveModel returns the model blob for the active config.
	GetActiveModel() ([]byte, error)
}
