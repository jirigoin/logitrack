package model

// PriorityPrediction is the response from the ML prediction service.
type PriorityPrediction struct {
	Priority   string                  `json:"priority"`   // alta / media / baja
	Confidence float64                 `json:"confidence"` // 0.0-1.0
	Score      float64                 `json:"score"`      // 0.0-1.0 weighted score
	Factors    map[string]FactorDetail `json:"factors"`
}

// FactorDetail shows how each factor contributed to the priority score.
type FactorDetail struct {
	Value        interface{} `json:"value"`
	Normalized   float64     `json:"normalized"`
	Weight       float64     `json:"weight"`
	Contribution float64     `json:"contribution"`
}

// MLServiceRequest is the payload sent to the Python ML service.
type MLServiceRequest struct {
	OriginProvince      string  `json:"origin_province"`
	DestinationProvince string  `json:"destination_province"`
	ShipmentType        string  `json:"shipment_type"`
	TimeWindow          string  `json:"time_window"`
	PackageType         string  `json:"package_type"`
	WeightKg            float64 `json:"weight_kg"`
	IsFragile           bool    `json:"is_fragile"`
	ColdChain           bool    `json:"cold_chain"`
}
