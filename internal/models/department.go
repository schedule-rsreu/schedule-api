package models

type Department struct {
	Id         int     `json:"id" bson:"id" example:"1"`
	Title      string  `json:"title" bson:"title" example:"Кафедра вычислительной и прикладной математики"`
	TitleShort string  `json:"title_short" bson:"short" example:"ВПМ"`
	Faculty    Faculty `json:"faculty" bson:"faculty" `
}
