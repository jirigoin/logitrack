package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/logitrack/core/internal/model"
)

// MLClient calls the Python ML prediction service.
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Predict calls the ML service and returns the priority prediction.
// Returns nil (no error, no prediction) if the service is unavailable.
func (c *MLClient) Predict(shipment model.Shipment) *model.PriorityPrediction {
	if c.baseURL == "" {
		return nil
	}

	req := model.MLServiceRequest{
		OriginProvince:      shipment.Sender.Address.Province,
		DestinationProvince: shipment.Recipient.Address.Province,
		ShipmentType:        string(shipment.ShipmentType),
		TimeWindow:          string(shipment.TimeWindow),
		PackageType:         string(shipment.PackageType),
		WeightKg:            shipment.WeightKg,
		IsFragile:           shipment.IsFragile,
		ColdChain:           shipment.ColdChain,
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("[MLClient] failed to marshal request: %v", err)
		return nil
	}

	resp, err := c.httpClient.Post(c.baseURL+"/predict", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[MLClient] prediction service unavailable: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[MLClient] prediction service returned %d: %s", resp.StatusCode, string(respBody))
		return nil
	}

	var prediction model.PriorityPrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		log.Printf("[MLClient] failed to decode response: %v", err)
		return nil
	}

	log.Printf("[MLClient] predicted priority=%s (confidence=%.2f, score=%.2f) for %s → %s",
		prediction.Priority, prediction.Confidence, prediction.Score,
		shipment.Sender.Address.Province, shipment.Recipient.Address.Province)

	return &prediction
}

// PredictFromCreateRequest is used when creating a shipment (before the Shipment struct exists).
func (c *MLClient) PredictFromCreateRequest(req model.CreateShipmentRequest) *model.PriorityPrediction {
	if c.baseURL == "" {
		return nil
	}

	// Default values for optional fields
	shipmentType := string(req.ShipmentType)
	if shipmentType == "" {
		shipmentType = "normal"
	}
	timeWindow := string(req.TimeWindow)
	if timeWindow == "" {
		timeWindow = "flexible"
	}

	mlReq := model.MLServiceRequest{
		OriginProvince:      req.Sender.Address.Province,
		DestinationProvince: req.Recipient.Address.Province,
		ShipmentType:        shipmentType,
		TimeWindow:          timeWindow,
		PackageType:         string(req.PackageType),
		WeightKg:            req.WeightKg,
		IsFragile:           req.IsFragile,
		ColdChain:           req.ColdChain,
	}

	body, err := json.Marshal(mlReq)
	if err != nil {
		log.Printf("[MLClient] failed to marshal request: %v", err)
		return nil
	}

	resp, err := c.httpClient.Post(c.baseURL+"/predict", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[MLClient] prediction service unavailable: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[MLClient] prediction service returned %d: %s", resp.StatusCode, string(respBody))
		return nil
	}

	var prediction model.PriorityPrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		log.Printf("[MLClient] failed to decode response: %v", err)
		return nil
	}

	log.Printf("[MLClient] predicted priority=%s (confidence=%.2f, score=%.2f) for %s → %s",
		prediction.Priority, prediction.Confidence, prediction.Score,
		req.Sender.Address.Province, req.Recipient.Address.Province)

	return &prediction
}

// setPriority sets the priority on a shipment from a prediction result.
func setPriority(shipment *model.Shipment, prediction *model.PriorityPrediction) {
	if prediction != nil {
		shipment.Priority = prediction.Priority
		fmt.Printf("[ML] Shipment %s priority set to: %s\n", shipment.TrackingID, prediction.Priority)
	}
}
