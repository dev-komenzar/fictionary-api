package db

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgresqlを使うためのライブラリ
	"github.com/pkg/errors"
	"github.com/tuckKome/fictionary-api/data"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func argInit() string {
	host := getEnv("FICTIONARY_DATABASE_HOST", "127.0.0.1")
	port := getEnv("FICTIONARY_PORT", "5432")
	user := getEnv("FICTIONARY_USER", "tahoiya")
	dbname := getEnv("FICTIONARY_DB_NAME", "dbtahoiya")
	password := getEnv("FICTIONARY_DB_PASS", "password")
	sslmode := getEnv("FICTIONARY_SSLMODE", "disable")

	dbinfo := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
	)
	return dbinfo
}

//Init : DB初期化
func Init() {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&data.Kaitou{}, &data.Game{}, &data.Vote{}, &data.Donation{})
	defer db.Close()
}

//GetGame : DBから一つ取り出す：回答ページで使用
func GetGame(id int) (data.Game, error) {
	var m data.Game
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return m, errors.Wrap(err, "failed to connect DB (db.GetGame)")
	}
	defer db.Close()

	var g data.Game
	err = db.First(&g, id).Error
	if err != nil {
		return m, errors.Wrap(err, "failed to find game (db.GetGame)")
	}
	return g, nil
}

func GetGamesPhaseIs(st string) ([]data.Game, error) {
	var m []data.Game
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return m, errors.Wrap(err, "failed to connect DB (db.GetGamesPhaseIs)")

	}
	defer db.Close()

	var gs []data.Game
	err = db.Where("phase = ?", st).Find(&gs).Error
	if err != nil {
		return m, errors.Wrap(err, "failed to find games (db.GetGamesPhaseIs)")
	}
	return gs, nil
}

//GetKaitou get Kaitou with Kaitou.ID
func GetKaitou(id int) (data.Kaitou, error) {
	var k data.Kaitou //an empty struct to return when any error happens
	var kk data.Kaitou
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return k, errors.Wrap(err, "failed to connect DB (db.GetKaitous)")
	}
	defer db.Close()

	err = db.First(&kk, id).Error
	if err != nil {
		return k, errors.Wrap(err, "failed to find the answer (db.GetKaitou)")
	}
	return kk, nil

}

//GetKaitous : DBから指定されたGame属する[]Kaitouを取り出す
func GetKaitous(g data.Game) ([]data.Kaitou, error) {
	var k []data.Kaitou       //エラー時に返す空のスライス
	var kaitous []data.Kaitou //[]Kaitou を入れるスライス

	//g にIDがあるかチェック
	if g.ID == 0 {
		return k, fmt.Errorf("given game lacks ID (db.GetKaitous)")
	}

	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return k, errors.Wrap(err, "failed to connect DB (db.GetKaitous)")
	}
	defer db.Close()

	// db.Where("game_id = ?", id).Find(&kaitous)
	err = db.Model(&g).Association("Kaitous").Find(&kaitous).Error
	if err != nil {
		return k, errors.Wrap(err, "failed to find Kaitous associated to the Game (db.GetKaitous)")
	}

	return kaitous, nil
}

func GetVotes(k data.Kaitou) ([]data.Vote, error) {
	var v, votes []data.Vote // v is an empty slice
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return v, errors.Wrap(err, "failed to connect DB (db.GetVotes)")
	}
	defer db.Close()

	err = db.Model(&k).Related(&votes).Error
	if err != nil {
		return v, errors.Wrap(err, "failed to find votes (db.GetVotes)")
	}
	return votes, nil
}

func InsertGame(g data.Game) (data.Game, error) {
	var m data.Game //return するための空の構造体
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return m, errors.Wrap(err, "failed to connect DB (db.InsertGame)")
	}
	defer db.Close()

	err = db.Create(&g).Error
	if err != nil {
		return m, errors.Wrap(err, "failed to create new game (db.InsertGame)")
	}

	err = db.Last(&g).Error
	if err != nil {
		return m, errors.Wrap(err, "failed to get last game (db.InsertGame)")
	}
	return g, nil
}

//InsertKaitou : DBに新しいkaitouを追加
func InsertKaitou(g data.Game, k data.Kaitou) error {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		return errors.Wrap(err, "failed to connect DB (db.InsertKaitou)")
	}
	defer db.Close()

	if err := db.Model(&g).Association("Kaitous").Append(&k).Error; err != nil {
		return errors.Wrap(err, "failed to create new Kaitou (db.InsertKaitou)")
	}

	return nil
}

func VoteTo(k data.Kaitou, v data.Vote) error {
	connect := argInit()
	db, err := gorm.Open("postgres", connect)
	if err != nil {
		errors.Wrap(err, "failed to connect DB (db.VoteTo)")
	}
	defer db.Close()

	err = db.First(&k).Error //これ必要？？？？？
	if err != nil {
		errors.Wrap(err, "failed to find answer (db.VoteTo)")
	}

	err = db.Model(&k).Association("Votes").Append(&v).Error
	if err != nil {
		errors.Wrap(err, "failed to create vote and associate to answer (db.VoteTo)")
	}

	return nil
}
