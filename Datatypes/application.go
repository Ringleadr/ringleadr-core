package Datatypes

type Application struct {
	Name       string      `json:"name" binding:"required"`
	Components []*Component `json:"components" binding:"required"`
}
