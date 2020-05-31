package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tuckKome/fictionary-api/db"
	"github.com/tuckKome/fictionary-api/handler"
)

func main() {
	router := gin.Default()
	db.Init()

	//Get games in Index
	router.GET("/archives", handler.Archives)
	router.GET("/accepting", handler.Accepting)
	router.GET("/playing", handler.Playing)

	//Post new game
	router.POST("/games", handler.CreateGame)

	//New Kaitou
	router.GET("/games/:id/new", handler.CanGetKaitou)
	router.POST("/games/:id/new", handler.CanCreateKaitou)

	//Verify if user is the questioner
	router.POST("/games/:id/verify", handler.IsQuestioner)

	//Get answers related to the Game
	router.GET("/games/:id/answers", handler.GetKaitous)

	//Post vote
	router.POST("/games/:id/new-vote", handler.CanCreateVote)

	//Launch
	router.Run()
}
