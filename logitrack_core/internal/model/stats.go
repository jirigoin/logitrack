package model

type Stats struct {
	Total    int            `json:"total"`
	ByStatus map[Status]int `json:"by_status"`
}
