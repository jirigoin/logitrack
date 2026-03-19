package model

import "time"

type Route struct {
	ID          string    `json:"id"`
	Date        string    `json:"date"`
	DriverID    string    `json:"driver_id"`
	ShipmentIDs []string  `json:"shipment_ids"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r Route) HasShipment(trackingID string) bool {
	for _, id := range r.ShipmentIDs {
		if id == trackingID {
			return true
		}
	}
	return false
}

type CreateRouteRequest struct {
	Date        string   `json:"date"         binding:"required"`
	DriverID    string   `json:"driver_id"    binding:"required"`
	ShipmentIDs []string `json:"shipment_ids" binding:"required"`
}
