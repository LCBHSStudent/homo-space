package common

type HomoInfo struct {
	ID              int64
	Name            string
	Description     string
	Level           int
	HP              int
	ATN             int
	INT             int
	DEF             int
	RES             int
	SPD             int
	LUK             int
	Skills          [6]int
	Status          int
}

var SkillList = [1]string{"Unknown"}

func remove(slice []interface{}, elem interface{}) []interface{}{
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...)
			return remove(slice,elem)
			break
		}
	}
	return slice
}