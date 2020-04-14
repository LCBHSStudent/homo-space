package lottery

import (
	"fmt"
	qqbotapi "github.com/catsworld/qq-bot-api"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"CQApp/src/dbTransition"
)

var aliasBot *qqbotapi.BotAPI
var createUserTable = `
(
	HOMO_ID             BIGINT      NOT NULL    PRIMARY KEY,
	HOMO_NAME           CHAR(50),
	HOMO_LEVEL          INT,
	HOMO_POTENTIALITY   INT,
	HOMO_EXP            INT,
	HOMO_QUALITY        INT
);
`

type Homo = dbTransition.Homo
//
var prob  = [4]int{15, 100, 1000, 1000}
//
var RareN   []Homo
var RareSR  []Homo
var RareUR  []Homo

var UpItem = Homo{
	ID:   35,
	Rare: "[UR★★★★★]",
	Name: "柏油贵公子(联通法师)",
}

func Init(bot *qqbotapi.BotAPI) {
	if bot == nil {
		panic("failed to bind a nil value")
	}
	aliasBot = bot
	GetHomoList()
}

func GetHomoList() {
	
	RareN  = make([]Homo, 0)
	RareSR = make([]Homo, 0)
	RareUR = make([]Homo, 0)
	runtime.GC()
	
	dbTransition.GetHomoList(&RareN,  "N")
	dbTransition.GetHomoList(&RareSR, "SR")
	dbTransition.GetHomoList(&RareUR, "UR")
	
	fmt.Println(RareN)
	fmt.Println(RareSR)
	fmt.Println(RareUR)
}

func draw(id int64, msg *qqbotapi.FlatSender) *qqbotapi.FlatSender {
	rand.Seed(time.Now().UnixNano())
	
	msg = msg.NewLine()
	
	value   := rand.Intn(1000) + 1
	rank    := 0
	for i := 0; i < len(prob); i++ {
		if value <= prob[i] {
			break
		}
		rank++
	}
	var item int
	switch rank {
	case 0:
		item = rand.Intn(len(RareUR))
		msg  = msg.Text("[UR★★★★★]" + RareUR[item].Name)
		dbTransition.NewHomoGet(id, RareUR[item].ID, RareUR[item].Name)
		break //UR
	case 1:
		item = rand.Intn(len(RareSR))
		msg  = msg.Text("[SR★★★]" + RareSR[item].Name)
		dbTransition.NewHomoGet(id, RareSR[item].ID, RareSR[item].Name)
		break //SR
	case 2:
		item = rand.Intn(len(RareN))
		msg  = msg.Text("[N★]" + RareN[item].Name)
		dbTransition.NewHomoGet(id, RareN[item].ID, RareN[item].Name)
		break //N
	case 3:
		msg  = msg.Text(UpItem.Rare + UpItem.Name)
		dbTransition.NewHomoGet(id, UpItem.ID, UpItem.Name)
		break
	default:
		break
	}
	return msg
}

func SingleDraw(update qqbotapi.Update) {
	id := update.Message.From.ID
	dbTransition.AddUser(update.Message.From.ID, update.GroupID)
	
	msg := aliasBot.NewMessage(update.GroupID, "group").At(strconv.FormatInt(id, 10))
	if len(RareUR) == 0 || len(RareN) == 0 || len(RareSR) == 0 {
		msg.Text("请先补充蛋池").Send()
		return
	}
	
	if dbTransition.GetUserTicket(id) < 5 {
		msg.NewLine().Text("恁的转蛋券尚不足5!").Send()
		return
	}
	dbTransition.IncreaseUserTicket(id, -5)
	
	
	dbTransition.CreateTableIfNotExist(
		"`"+strconv.FormatInt(update.Message.From.ID, 10)+"`", createUserTable)
	
	draw(id, msg).Send()
}

func MultiDraw(update qqbotapi.Update) {
	id := update.Message.From.ID
	dbTransition.AddUser(update.Message.From.ID, update.GroupID)
	
	msg := aliasBot.NewMessage(update.GroupID, "group").At(strconv.FormatInt(id, 10))
	
	if len(RareUR) == 0 || len(RareN) == 0 || len(RareSR) == 0 {
		msg.Text("请先补充蛋池").Send()
		return
	}
	
	if dbTransition.GetUserTicket(id) < 45 {
		msg.NewLine().Text("恁的转蛋券尚不足45!").Send()
		return
	}
	dbTransition.IncreaseUserTicket(id, -45)
	
	
	dbTransition.CreateTableIfNotExist(
		"`"+strconv.FormatInt(update.Message.From.ID, 10)+"`", createUserTable)
	
	for i := 0; i < 10; i++ {
		msg = draw(id, msg)
	}
	msg.Send()
}

func ShowTicketCnt(id int64, group int64) {
	dbTransition.AddUser(id, group)
	cnt := dbTransition.GetUserTicket(id)
	if cnt != -810 {
		aliasBot.NewMessage(group, "group").
			At(strconv.FormatInt(id, 10)).
			NewLine().
			Text("[剩余转蛋券] ").
			Text(strconv.FormatInt(cnt, 10)).
			Send()
	} else {
		aliasBot.NewMessage(group, "group").
			At(strconv.FormatInt(id, 10)).
			Text("您还未被加入HomoSpace的数据库哦").
			Send()
	}
}

func ShowDrawPool(groupID int64) {
	msg := aliasBot.NewMessage(groupID, "group").Text("")
	msg = msg.Text("[N★]: ")
	for _, homo := range RareN {
		msg = msg.Text("["+homo.Name+"], ")
	}
	msg = msg.NewLine()
	msg = msg.NewLine()
	
	msg = msg.Text("[SR★★★]: ")
	for _, homo := range RareSR {
		msg = msg.Text("["+homo.Name+"], ")
	}
	msg = msg.NewLine()
	msg = msg.NewLine()
	
	msg = msg.Text("[UR★★★★★]: ")
	for _, homo := range RareUR {
		msg = msg.Text("["+homo.Name+"], ")
	}
	msg.Send()
}

type Member = dbTransition.Member
type Members []Member

func (m Members) Len() int {
	return len(m)
}
func (m Members) Swap(i, j int) {
	m[i].Id, m[j].Id = m[j].Id, m[i].Id
	m[i].ColRate, m[j].ColRate = m[j].ColRate, m[i].ColRate
}
// rise ordered
func (m Members) Less(i, j int) bool {
	return m[i].ColRate > m[j].ColRate
}

func PrintCollectionRank(group int64) {

	users := make(Members, 0, 45)

	con := dbTransition.GetConn()
	rows, err := con.Query("SELECT ID FROM USER WHERE FROM_GROUP=?", group)
	if err != nil {
		aliasBot.NewMessage(group, "group").
			Text(err.Error()).Send()
	} else {
		var id int64
		for rows.Next() {
			_ = rows.Scan(&id)
			users = append(users, Member{Id: id})
		}
		if len(users) == 0 {
			aliasBot.NewMessage(group, "group").Text("本群内还无人登记哦").Send()
		}
		dbTransition.UpdateMemberInfo((*[]dbTransition.Member)(&users))
		sort.Sort(users)

		msg := aliasBot.NewMessage(group, "group").Text("")
		homoCnt := float64(dbTransition.GetHomoCount())

		for index, user := range users {
			info, err := aliasBot.GetGroupMemberInfo(group, user.Id, false)
			if err != nil {
				aliasBot.NewMessage(group, "group").
					Text(err.Error()).Send()
			} else {
				colRate := strconv.FormatFloat(float64(user.ColRate)/homoCnt*100, 'f', 2, 64) + "%"
				msg = msg.Text(strings.Join([]string{"No.", strconv.Itoa(index+1), "【", info.Name(), "】: ", colRate}, ""))
				if index != len(users) - 1 {
					msg = msg.NewLine()
				}
			}
		}
		msg.Send()
	}
}