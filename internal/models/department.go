package models

type Department struct {
	Title      string  `json:"title"       bson:"title"   example:"Кафедра вычислительной и прикладной математики"`
	TitleShort string  `json:"title_short" bson:"short"   example:"ВПМ"`
	Faculty    Faculty `json:"faculty"     bson:"faculty"`
	Id         int     `json:"id"          bson:"id"      example:"1"`
}
