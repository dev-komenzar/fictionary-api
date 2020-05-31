package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/tuckKome/fictionary-api/data"
	"github.com/tuckKome/fictionary-api/db"
)

const (
	accepting   = "accepting"
	playing     = "playing"
	archive     = "archive"
	linkToError = "/error"
	games       = "/games/"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func isNill(id string) error {
	if id == "" {
		return errors.New("ERROR : Cannnot get game ID")
	}
	return nil
}

func contains(a string, v []data.Vote) bool {
	for i := range v {
		if a == v[i].CreatedBy {
			return true
		}
	}
	return false
}

//Archives get all archives of games
func Archives(c *gin.Context) {
	a, err := db.GetGamesPhaseIs(archive)
	if err != nil {
		log.Println(fmt.Errorf("Error(handler.Archives): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "did NOT get archives",
		})
		return
	}

	// Sample json
	// {
	// 	"ID": 18,
	// 	"CreatedAt": "2020-05-18T11:41:20.487614+09:00",
	// 	"UpdatedAt": "2020-05-18T11:41:49.665435+09:00",
	// 	"DeletedAt": null,
	// 	"Odai": "てすとテスてすと",
	// 	"Kaitous": null,
	// 	"Phase": "archive",
	// 	"CreatedBy": "Kometan",
	// 	"Secret": "ひみつ"
	//   }

	sort.SliceStable(a, func(i, j int) bool {
		return a[i].UpdatedAt.After(a[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, a)
}

//Playing return json which contains playing games
func Playing(c *gin.Context) {
	a, err := db.GetGamesPhaseIs(playing)
	if err != nil {
		log.Println(fmt.Errorf("Error(handler.Playing): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "did NOT get playing games",
		})
		return
	}

	sort.SliceStable(a, func(i, j int) bool {
		return a[i].UpdatedAt.After(a[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, a)
}

//Accepting write json which contains accepting games to context
func Accepting(c *gin.Context) {
	a, err := db.GetGamesPhaseIs(accepting)
	if err != nil {
		log.Println(fmt.Errorf("Error(handler.Accepting): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "did NOT get accepting games",
		})
		return
	}

	sort.SliceStable(a, func(i, j int) bool {
		return a[i].UpdatedAt.After(a[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, a)
}

//CreateGame makes new game and set correct answer
func CreateGame(c *gin.Context) {
	g := data.Game{Phase: accepting}
	c.Bind(&g)
	g, err := db.InsertGame(g)
	if err != nil {
		log.Println(fmt.Errorf("Error(handle.GreateGame): %s", err))
	}

	//Creater the correct answer
	a := c.Query("answer")
	k := data.Kaitou{User: g.CreatedBy, Answer: a, GameID: g.ID, IsCorrect: true}
	if err := db.InsertKaitou(g, k); err != nil {
		log.Println(fmt.Errorf("Error(handler.CreateGame): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "did NOT create new game",
		})
		return
	}

	c.JSONP(http.StatusCreated, gin.H{
		"message":     "ok",
		"new-game-id": g.ID,
	})
}

func createKaitou(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (createKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	k := data.Kaitou{IsCorrect: false}
	c.Bind(&k)
	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (createKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not find this game",
		})
		return
	}
	err = db.InsertKaitou(g, k)
	if err != nil {
		log.Println(fmt.Errorf("Error (createKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not make new answer",
		})
		return
	}

	c.JSONP(http.StatusCreated, gin.H{
		"message": "ok",
	})
}

func createVote(c *gin.Context) {
	//Recieve {"vote-to":132, "created-by": "こめたん"}
	v := data.Vote{}
	c.Bind(&v)

	//check if duplicate
	k, err := db.GetKaitou(v.KaitouID)
	if err != nil {
		log.Println(fmt.Errorf("Error (createVote): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not get the answer",
		})
		return
	}

	n := c.Param("id")
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (createVote): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID error",
		})
		return
	}
	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (createVote): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not get the game",
		})
		return
	}

	kk, err := db.GetKaitous(g)
	if err != nil {
		log.Println(fmt.Errorf("Error (createVote): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not get the answers",
		})
		return
	}
	var vv []data.Vote
	for i := range kk {
		votes, _ := db.GetVotes(kk[i])
		vv = append(vv, votes...)
	}

	if contains(v.CreatedBy, vv) {
		c.JSONP(http.StatusConflict, gin.H{
			"message": "You have voted already.",
		})
		return
	}

	db.VoteTo(k, v) //Kaitou に Vote を紐つける
	c.JSONP(http.StatusOK, gin.H{
		"message": "Succeed",
	})
}

func getNewKaitou(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (getNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (getNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could NOT find this game.",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"game-id": g.ID,
		"odai":    g.Odai,
	})
}

//GetKaitous returns json which contains submitted answers related to the game
func GetKaitous(c *gin.Context) {
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (GetKaitousInAdvance): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is empty.",
		})
		return
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (GetKaitousInAdvance): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (GetKaitousInAdvance): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Game not found.",
		})
		return
	}

	k, err := db.GetKaitous(g)
	if err != nil {
		log.Println(fmt.Errorf("Error (GetKaitousInAdvance): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could not get answers.",
		})
		return
	}

	c.JSONP(http.StatusOK, k)
}

//CanGetKaitou is switch if the game's phase is accepting
func CanGetKaitou(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is empty.",
		})
		return
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))

		c.JSONP(http.StatusConflict, gin.H{
			"message": "Could NOT find this game.",
		})
		return
	}

	switch g.Phase {
	case accepting:
		getNewKaitou(c)
	default:
		c.JSONP(http.StatusForbidden, gin.H{
			"message": "New answers are NOT acceptable.",
		})
	}
}

//CanCreateKaitou is switch if the game's phase is accepting
func CanCreateKaitou(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is empty.",
		})
		return
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	g, err := db.GetGame(id)
	if err != nil {
		log.Println(fmt.Errorf("Error (CanGetNewKaitou): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Game not found.",
		})
	}
	switch g.Phase {
	case accepting:
		createKaitou(c)
	default:
		c.JSONP(http.StatusForbidden, gin.H{
			"message": "New answers are NOT acceptable.",
		})
	}
}

func CanCreateVote(c *gin.Context) {
	//idをint型に変換
	n := c.Param("id")
	err := isNill(n)
	if err != nil {
		log.Println(err)
		c.Redirect(302, linkToError)
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(err)
		c.Redirect(302, linkToError)
		return
	}
	g, err := db.GetGame(id)

	switch g.Phase {
	case playing:
		createVote(c)
	default:
		c.JSONP(http.StatusForbidden, gin.H{
			"message": "New votes are NOT acceptable.",
		})
	}
}

//IsQuestioner : id, secret を受け取って合言葉を検証
func IsQuestioner(c *gin.Context) {
	n := c.Query("game-id")
	err := isNill(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (IsQuestioner): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is empty.",
		})
		return
	}
	id, err := strconv.Atoi(n)
	if err != nil {
		log.Println(fmt.Errorf("Error (IsQuestioner): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "ID is not int.",
		})
		return
	}

	g, err := db.GetGame(id)
	if err == nil {
		log.Println(fmt.Errorf("Error (IsQuestioner): %s", err))
		c.JSONP(http.StatusConflict, gin.H{
			"message": "Game not found.",
		})
		return
	}

	s := c.Query("secret")
	if s == g.Secret {
		c.JSONP(http.StatusOK, gin.H{
			"verify": true,
		})
	} else {
		c.JSONP(http.StatusForbidden, gin.H{
			"verify": false,
		})
	}
}
