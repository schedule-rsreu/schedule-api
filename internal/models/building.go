package models

type Building struct {
	Title  string `json:"title"  example:"Центральный корпус"`
	Letter string `json:"letter" example:"C"`
	Id     int    `json:"id"     example:"1"`
}
