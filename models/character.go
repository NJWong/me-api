package models

type Character struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Species int    `json:"species"`
	Gender  int    `json:"gender"`
	Class   string `json:"class"`
}
