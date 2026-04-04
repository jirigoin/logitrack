package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/logitrack/core/internal/model"
)

type postgresMLConfigRepository struct {
	db *sql.DB
}

func NewPostgresMLConfigRepository(db *sql.DB) MLConfigRepository {
	return &postgresMLConfigRepository{db: db}
}

func (r *postgresMLConfigRepository) GetActive() (*model.MLConfig, error) {
	row := r.db.QueryRow(`
		SELECT id, factors, alta_threshold, media_threshold, is_active, created_by, created_at, notes
		FROM ml_configs
		WHERE is_active = TRUE
		LIMIT 1
	`)
	cfg, err := scanMLConfig(row.Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active ml config: %w", err)
	}
	return &cfg, nil
}

func (r *postgresMLConfigRepository) List() ([]model.MLConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, factors, alta_threshold, media_threshold, is_active, created_by, created_at, notes
		FROM ml_configs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list ml configs: %w", err)
	}
	defer rows.Close()

	var configs []model.MLConfig
	for rows.Next() {
		cfg, err := scanMLConfig(rows.Scan)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (r *postgresMLConfigRepository) Create(config model.MLConfig) (model.MLConfig, error) {
	factorsJSON, err := json.Marshal(config.Factors)
	if err != nil {
		return model.MLConfig{}, fmt.Errorf("marshal factors: %w", err)
	}

	var id int
	var createdAt time.Time
	err = r.db.QueryRow(`
		INSERT INTO ml_configs (factors, alta_threshold, media_threshold, is_active, created_by, notes)
		VALUES ($1, $2, $3, FALSE, $4, $5)
		RETURNING id, created_at
	`, factorsJSON, config.AltaThreshold, config.MediaThreshold, config.CreatedBy, config.Notes).
		Scan(&id, &createdAt)
	if err != nil {
		return model.MLConfig{}, fmt.Errorf("create ml config: %w", err)
	}

	config.ID = id
	config.CreatedAt = createdAt
	config.IsActive = false
	return config, nil
}

func (r *postgresMLConfigRepository) Activate(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE ml_configs SET is_active = FALSE WHERE is_active = TRUE`); err != nil {
		return fmt.Errorf("deactivate configs: %w", err)
	}
	if _, err := tx.Exec(`UPDATE ml_configs SET is_active = TRUE WHERE id = $1`, id); err != nil {
		return fmt.Errorf("activate config %d: %w", id, err)
	}
	return tx.Commit()
}

func (r *postgresMLConfigRepository) SaveModel(configID int, data []byte) error {
	_, err := r.db.Exec(`
		INSERT INTO ml_models (config_id, model_data, size_bytes)
		VALUES ($1, $2, $3)
	`, configID, data, len(data))
	if err != nil {
		return fmt.Errorf("save ml model: %w", err)
	}
	return nil
}

func (r *postgresMLConfigRepository) GetActiveModel() ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(`
		SELECT m.model_data
		FROM ml_models m
		JOIN ml_configs c ON c.id = m.config_id
		WHERE c.is_active = TRUE
		ORDER BY m.created_at DESC
		LIMIT 1
	`).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active model: %w", err)
	}
	return data, nil
}

func scanMLConfig(scan func(...any) error) (model.MLConfig, error) {
	var cfg model.MLConfig
	var factorsJSON []byte
	err := scan(
		&cfg.ID, &factorsJSON, &cfg.AltaThreshold, &cfg.MediaThreshold,
		&cfg.IsActive, &cfg.CreatedBy, &cfg.CreatedAt, &cfg.Notes,
	)
	if err != nil {
		return model.MLConfig{}, err
	}
	if err := json.Unmarshal(factorsJSON, &cfg.Factors); err != nil {
		return model.MLConfig{}, fmt.Errorf("unmarshal factors: %w", err)
	}
	return cfg, nil
}
