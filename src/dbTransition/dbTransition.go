package dbTransition

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var (
	aliasDB     *sql.DB
)

func Init(db *sql.DB) {
	if db == nil {
		panic("failed to bind a nil value")
	}
	aliasDB  = db
	CreateTableIfNotExist("USER",
		`
(
	ID      INT     NOT NULL    PRIMARY KEY,
	TICKET  INT     NOT NULL,
	DAILY_LIMIT INT NOT NULL    DEFAULT 0
);
	`)
	CreateTableIfNotExist("HOMO",
		`
(
	ID          INT     NOT NULL    PRIMARY KEY   AUTO_INCREMENT,
	NAME        VARCHAR(128)     NOT NULL,
	DESCRIPTION VARCHAR(255)     DEFAULT `+"'妹有描述desu'," +`
	RARE        VARCHAR(15)      NOT NULL,
	MAX_LEVEL   INT              DEFAULT 114,

	INITIAL_HP  INT     DEFAULT 1,
	INITIAL_ATN INT     DEFAULT 0,
	INITIAL_INT INT     DEFAULT 0,
	INITIAL_DEF INT     DEFAULT 0,
	INITIAL_RES INT     DEFAULT 0,
	INITIAL_SPD INT     DEFAULT 0,
	INITIAL_LUK INT     DEFAULT 0,
	
	GROWTH_HP   INT     DEFAULT 0,
	GROWTH_ATN  INT     DEFAULT 0,
	GROWTH_INT  INT     DEFAULT 0,
	GROWTH_DEF  INT     DEFAULT 0,
	GROWTH_RES  INT     DEFAULT 0,
	GROWTH_SPD  INT     DEFAULT 0,
	GROWTH_LUK  INT     DEFAULT 0,

	SKILL_ID_1  INT     DEFAULT -1,
	SKILL_ID_2  INT     DEFAULT -1,
	SKILL_ID_3  INT     DEFAULT -1,
	SKILL_ID_4  INT     DEFAULT -1,
	SKILL_ID_5  INT     DEFAULT -1,
	SKILL_ID_6  INT     DEFAULT -1,
	
	ACQUIRABLE  BOOL    DEFAULT TRUE,
	PROB_UP     BOOL    DEFAULT FALSE
);
	`)
	go TimeToUpdateDailyLimit()
}

func checkTableExist(table string) bool {
	query  := "SELECT nsp FROM " + table
	_, err := aliasDB.Query(query)
	if  err != nil &&
		err.Error() == "Error 1054: Unknown column 'nsp' in 'field list'" {
		return true
	} else {
		return false
	}
}

func CreateTableIfNotExist(table string, tableFormat string) {
	
	if checkTableExist(table) {
		return
	}
	
	create := `
CREATE TABLE IF NOT EXISTS ` + table + tableFormat
	_, err := aliasDB.Exec(create)
	if err != nil {
		panic(err)
	}
}

func AddUser(id int64) {
	var count int64
	err := aliasDB.QueryRow("SELECT count(*) FROM USER WHERE ID=?", id).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		_, err := aliasDB.Exec("INSERT INTO USER(ID,TICKET,DAILY_LIMIT) VALUES (?,?,?)", id, 45, 0)
		if err != nil {
			panic(err)
		}
	}
}

type Homo struct {
	ID      int
	Rare    string
	Name    string
}

func GetHomoList(list *[]Homo, rare string) () {
	rows, err := aliasDB.Query("SELECT ID,RARE,NAME FROM HOMO WHERE RARE = ? AND ACQUIRABLE = TRUE", rare)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		pkm := Homo{}
		if err := rows.Scan(&pkm.ID, &pkm.Rare, &pkm.Name); err != nil {
			log.Fatal(err)
		}
		*list = append(*list, pkm)
		//fmt.Println("Load HOMO message:", pkm.ID, pkm.Rare, pkm.Name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func GetUserTicket(id int64) int64 {
	var ticket int64
	err := aliasDB.QueryRow("SELECT TICKET FROM USER WHERE ID=?", id).Scan(&ticket)
	if err != nil {
		return -810
	}
	return ticket
}

func DetectDailyLimit(id int64) bool {
	var daily int64
	err := aliasDB.QueryRow("SELECT DAILY_LIMIT FROM USER WHERE ID=?", id).Scan(&daily)
	if err != nil {
		panic(err)
	}
	if daily < 20 {
		_, _ = aliasDB.Exec("UPDATE USER SET DAILY_LIMIT=? WHERE ID=?", daily+1, id)
		return false
	} else {
		return true
	}
}

func IncreaseUserTicket(id int64, count int64) {
	var ticket int64
	err := aliasDB.QueryRow("SELECT TICKET FROM USER WHERE ID=?", id).Scan(&ticket)
	if err != nil {
		panic(err)
	}
	_, err = aliasDB.Exec("UPDATE USER SET TICKET=? WHERE ID=?", ticket+count, id)
	if err != nil {
		panic(err)
	}
}

func NewHomoGet(id int64, HomoID int, HomoName string) {
	var count int64
	tableId := strconv.FormatInt(id,10)
	cntQuery :=  "SELECT count(*) FROM `" + tableId + "` WHERE HOMO_ID=?"
	err := aliasDB.QueryRow(cntQuery, HomoID).Scan(&count)
	if count > 0 {
		var potentiality int64
		var quality      int64
		query :=  "SELECT HOMO_POTENTIALITY,HOMO_QUALITY FROM `" + tableId + "` WHERE HOMO_ID=?"
		err = aliasDB.QueryRow(query, HomoID).Scan(&potentiality, &quality)
		if err != nil {
			log.Println(err)
		}
		if potentiality < 99 {
			_, err = aliasDB.Exec("UPDATE `"+tableId+
				"` SET HOMO_POTENTIALITY=? WHERE HOMO_ID=?", potentiality+1, HomoID)
			if err != nil {
				log.Println(err)
			}
		}
		newQuality := rand.Intn(100)+1
		if newQuality > int(quality) {
			_, err = aliasDB.Exec("UPDATE `"+tableId+
				"` SET HOMO_QUALITY=? WHERE HOMO_ID=?", newQuality, HomoID)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		_, err = aliasDB.Exec("INSERT INTO `" + tableId +
			"`(HOMO_ID,HOMO_NAME,HOMO_LEVEL,HOMO_POTENTIALITY,HOMO_EXP,HOMO_QUALITY)" +
			"VALUES(?,?,?,?,?,?)", HomoID, HomoName,
			1, 0, 0, rand.Intn(100)+1,
		)
		if err != nil {
			log.Println(err)
		}
	}
}

func UpdateDailyLimit() {
	_, err := aliasDB.Exec("UPDATE USER SET DAILY_LIMIT=0")
	if err != nil {
		log.Println(err)
	}
}

func TimeToUpdateDailyLimit() {
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(
			next.Year(), next.Month(), next.Day(),
			0, 0, 0, 0,
			next.Location(),
		)
		t := time.NewTimer(next.Sub(now))
		<-t.C
		
		UpdateDailyLimit()
	}
}

func GetOnesAsset(id int64) (homo []string) {
	tableId := strconv.FormatInt(id,10)
	homo = make([]string, 0)
	rows, err := aliasDB.Query(
		"SELECT * FROM `" + tableId + "`",
	)
	if err != nil {
		log.Println(err)
	} else {
		var ID           int
		var Name         string
		var LEVEL        int
		var Potentiality int
		var Exp          int
		var Quality      int
		for rows.Next() {
			err := rows.Scan(&ID, &Name, &LEVEL, &Potentiality, &Exp, &Quality)
			if err != nil {
				log.Println(err)
			} else {
				data := "No." +strconv.Itoa(ID) + "\n名称: [" + Name + "] 等级: [" + strconv.Itoa(LEVEL) +
					"] 突破等级: [" + strconv.Itoa(Potentiality) + "] 经验值: [" +
					strconv.Itoa(Exp) + "] 潜能: [" + strconv.Itoa(Quality) + "]"
				homo = append(homo, data)
			}
		}
	}
	return homo
}

func GetConn() *sql.DB {
	return aliasDB
}