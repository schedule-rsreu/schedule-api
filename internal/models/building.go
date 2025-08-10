package models

type Building struct {
	Id     int    `json:"id"     example:"1"`
	Title  string `json:"title"  example:"Центральный корпус"`
	Letter string `json:"letter" example:"C"`
}
