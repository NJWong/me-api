package models

type Species struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SpeciesObject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
