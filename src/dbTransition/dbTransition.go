package dbTransition

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	
	"CQApp/src/common"
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
	ID      BIGINT  NOT NULL    PRIMARY KEY,
	TICKET  INT     NOT NULL,
	DAILY_LIMIT INT NOT NULL    DEFAULT 0
);
	`)
	CreateTableIfNotExist("HOMO",
		`
(
	ID          BIGINT           NOT NULL    PRIMARY KEY   AUTO_INCREMENT,
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
		_, err := aliasDB.Exec("INSERT INTO USER(ID,TICKET,DAILY_LIMIT) VALUES (?,?,?)", id, 90, 0)
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
	if daily < 25 {
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
	rand.Seed(time.Now().UnixNano())
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

func DisplaySingleHomoInfo(name string) string {
	var ID, MaxLevel int
	var InitialHP, InitialATN, InitialINT, InitialDEF, InitialRES, InitialSPD, InitialLUK int
	var GrowthHP, GrowthATN, GrowthINT, GrowthDEF, GrowthRES, GrowthSPD, GrowthLUK int
	var Name, DESCRIPTION, RARE string
	var ActiveSkills [4]int
	var PassiveSkills [2]int
	var Acquirable, UP bool
	err := aliasDB.QueryRow(
		"SELECT * FROM HOMO WHERE NAME=?",
		name).Scan(&ID, &Name, &DESCRIPTION, &RARE, &MaxLevel,
			&InitialHP, &InitialATN, &InitialINT, &InitialDEF, &InitialRES, &InitialSPD, &InitialLUK,
			&GrowthHP, &GrowthATN, &GrowthINT, &GrowthDEF, &GrowthRES, &GrowthSPD, &GrowthLUK,
			&ActiveSkills[0], &ActiveSkills[1], &ActiveSkills[2], &ActiveSkills[3],
			&PassiveSkills[0], &PassiveSkills[1], &Acquirable, &UP)
	if err != nil {
		return err.Error()
	}
	msg := strings.Join([]string{
		"[No.", strconv.Itoa(ID), "] " , Name, "\n",
		"稀有度: ", RARE, "\n【", DESCRIPTION, "】 最大等级: ", strconv.Itoa(MaxLevel),
		"\n初始属性: \n", "HP: ", strconv.Itoa(InitialHP), "  物理攻击: ", strconv.Itoa(InitialATN),
		"\n魔法攻击: ", strconv.Itoa(InitialINT), "  物理防御: ", strconv.Itoa(InitialDEF),
		"\n魔法防御: ", strconv.Itoa(InitialRES), "  速度: ", strconv.Itoa(InitialSPD), "  幸运: ", strconv.Itoa(InitialLUK),
		"\n成长属性: \n", "物理攻击: ", strconv.Itoa(GrowthATN),
		"\n魔法攻击: ", strconv.Itoa(GrowthINT), "  物理防御: ", strconv.Itoa(GrowthDEF),
		"\n魔法防御: ", strconv.Itoa(GrowthRES), "  速度: ", strconv.Itoa(GrowthSPD), "  幸运: ", strconv.Itoa(GrowthLUK),
		"\n主动技能：", "[", common.SkillList[0], "], " , "[", common.SkillList[0], "], ",
		"[", common.SkillList[0], "], ", "[", common.SkillList[0], "]",
		"\n被动技能：", "[", common.SkillList[0], "], ", "[", common.SkillList[0], "]",
		"\n可否转蛋获得: ", strconv.FormatBool(Acquirable),
		"\n是否概率UP中: ", strconv.FormatBool(UP),
	}, "")
	return msg
}

func GetConn() *sql.DB {
	return aliasDB
}