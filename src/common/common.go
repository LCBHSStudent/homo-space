package common

type HomoInfo struct {
	ID              int64
	Name            string
	Description     string
	Rare            string
	Level           int64
	HP              int64
	ATN             int64
	INT             int64
	DEF             int64
	RES             int64
	SPD             int64
	LUK             int64
	Skills          [6]int64
}