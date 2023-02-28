package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

//inter macro define
const (
	TplFile = "example.html"
)

type View struct {
	Title string `json:"title"`
	Nick string `json:"nick"`
}

//face
type SubApp struct {

}

//construct
func NewSubApp() *SubApp {
	this := &SubApp{}
	return this
}

//main entry
func (f *SubApp) Entry(c *gin.Context) {
	var (
		view View
	)

	//get request form

	//set view
	view.Title = "Home"
	view.Nick = "You"

	//output page
	c.HTML(http.StatusOK, TplFile, view)
}
