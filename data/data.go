package data

import (
	"github.com/jinzhu/gorm"
)

// Kaitou : 回答のDB　回答を集める
type Kaitou struct {
	gorm.Model
	User      string `json:"created-by"`
	Answer    string `json:"answer"`
	GameID    uint   `json:"game-id"`
	Base      int    `json:"base"` //シャッフルのための変数
	Votes     []Vote
	IsCorrect bool `json:"is-correct"`
}

// Game : ゲームのDB index>履歴　にも使う
type Game struct {
	gorm.Model
	Odai      string `json:"odai"` //お題
	Kaitous   []Kaitou
	Phase     string `json:"phase"` // accepting | playing | archive
	CreatedBy string //作った人
	Secret    string //合言葉
}

// Vote : 投票。Kaitou has many Votes
type Vote struct {
	gorm.Model
	KaitouID  int    `json:"vote-to"`
	CreatedBy string `json:"created-by"`
}

type Donation struct {
	gorm.Model
	Who      string
	HowMuch  int
	HowToPay string
}
