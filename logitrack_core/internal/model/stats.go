package model

type Stats struct {
	Total    int            `json:"total"`
	ByStatus map[Status]int `json:"by_status"`
	ByBranch map[string]int `json:"by_branch"` // branch ID → shipment count (excludes delivered/returned)
}
