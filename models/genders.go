package models

type Gender struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GenderObject struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
