package model

import "time"

// MLConfig holds the factor weights and classification thresholds used for priority prediction.
type MLConfig struct {
	ID             int                `json:"id"`
	Factors        map[string]float64 `json:"factors"`
	AltaThreshold  float64            `json:"alta_threshold"`
	MediaThreshold float64            `json:"media_threshold"`
	IsActive       bool               `json:"is_active"`
	CreatedBy      string             `json:"created_by"`
	CreatedAt      time.Time          `json:"created_at"`
	Notes          string             `json:"notes"`
}
