package main

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"homo-space/src/common"
	"homo-space/src/dbTransition"
	"homo-space/src/homo"
	"homo-space/src/lottery"
	"homo-space/util"

	qqbotapi "github.com/catsworld/qq-bot-api"
	_ "github.com/go-sql-driver/mysql"
)

var bot *qqbotapi.BotAPI
var db *sql.DB
var keyWords = [11]string{
	"转蛋单抽", "转蛋十连", "转蛋奖池", "HOMOSPACE", "编辑HOMO", "我的转蛋券", "HOMO图鉴", "准备对战", "收集排行", "群登记", "Document",
}

var CDKMap map[string]int64
var GroupList = [1]int64{930378083}

var ChanList []chan qqbotapi.Update
var ChanMutex sync.RWMutex

const CQHttpConnKey = "cqhttp-ws-connect"
const MySqlConnKey = "mysql-connect"

func main() {
	var err error

	cqhttpConf := util.GetObjectByKey(CQHttpConnKey).(map[interface{}]interface{})
	bot, err = qqbotapi.NewBotAPI("",
		strings.Join([]string{
			"ws://",
			cqhttpConf["ipv4"].(string),
			":",
			cqhttpConf["port"].(string),
		}, ""), cqhttpConf["secret"].(string),
	)
	if err != nil {
		log.Println(err)
	}
	bot.Debug = false

	conf := qqbotapi.NewUpdate(0)
	conf.PreloadUserInfo = true
	updates, err := bot.GetUpdatesChan(conf)
	if err != nil {
		log.Println(err)
		panic("connect to websocket failed.")
	}

	mysqlConf := util.GetObjectByKey(MySqlConnKey).(map[interface{}]interface{})

	connectStr := strings.Join([]string{
		mysqlConf["username"].(string),
		":",
		mysqlConf["password"].(string),
		"@tcp(",
		mysqlConf["ipv4"].(string), ":", mysqlConf["port"].(string),
		")/",
		mysqlConf["db-name"].(string),
	}, "")
	db, err = sql.Open("mysql", connectStr)
	if err != nil {
		log.Println(err)
		panic("open database-MySql failed.")
	}
	defer db.Close()

	dbTransition.Init(db)
	lottery.Init(bot)
	homo.Init(bot)

	CDKMap = make(map[string]int64)
	go Time2SendCDK()

	for update := range updates {
		// 判断消息属性
		if update.Message == nil || update.MessageType != "group" {
			continue
		}
		if _, ok := util.GetObjectByKey("group-enable-homo-space").(map[int64]bool)[update.GroupID]; !ok {
			continue
		}
		// 向下一级分发消息
		for _, ch := range ChanList {
			ch <- update
		}
		// log.Println(update.Message)
		// detect is const operation
		var flag int = -1
		for index, str := range keyWords {
			if str == update.Message.Text {
				flag = index
				break
			}
		}
		switch flag {
		case 0:
			lottery.SingleDraw(update)
			break
		case 1:
			lottery.MultiDraw(update)
			break
		case 2:
			lottery.ShowDrawPool(update.GroupID)
			break
		case 3:
			homo.DisplayAsset(update)
			break
		case 4:
			addMissionChan := make(chan struct{}, 1)
			go func() {
				homo.EditHomo(
					updates, addMissionChan,
					update.Message.From.ID,
					update.GroupID,
				)
				lottery.GetHomoList()
			}()
			<-addMissionChan
			break
		case 5:
			lottery.ShowTicketCnt(
				update.Message.From.ID,
				update.GroupID,
			)
			break
		case 6:
			homo.DisplayAllHomo(update.GroupID)
			break
		case 7:
			go func() {
				updateChan := make(chan qqbotapi.Update, 1)

				ChanMutex.Lock()
				pos := len(ChanList)
				ChanList = append(ChanList, updateChan)
				ChanMutex.Unlock()

				homo.Prepare4Battle(
					updateChan, //addMissionChan,
					update.Message.From.ID,
					update.GroupID,
				)
				close(updateChan)
				ChanMutex.Lock()
				ChanList = append(ChanList[:pos], ChanList[pos+1:len(ChanList)]...)
				ChanMutex.Unlock()
			}()
			break
		case 8:
			lottery.PrintCollectionRank(update.GroupID)
			break
		case 9:
			err := dbTransition.UpdateFromGroup(update.Message.From.ID, update.GroupID)
			if err != nil {
				bot.NewMessage(update.GroupID, "group").
					At(string(update.Message.From.ID)).NewLine().
					Text(err.Error()).Send()
			} else {
				bot.NewMessage(update.GroupID, "group").
					At(strconv.FormatInt(update.Message.From.ID, 10)).NewLine().
					Text("更新成功力!已登记到本群").Send()
			}
		case 10:
			PrintHelpInfo(update.GroupID)
			break
		default:
			handleMsg(update)
			break
		}
	}
}

func handleMsg(update qqbotapi.Update) {
	if _, ok := util.GetObjectByKey("group-enable-homo-space").(map[int64]bool)[update.GroupID]; !ok {
		return
	}

	if group, ok := CDKMap[update.Message.Text]; ok && group == update.GroupID {
		go func() {
			delete(CDKMap, update.Message.Text)
			bot.NewMessage(update.GroupID, "group").At(strconv.FormatInt(update.Message.From.ID, 10)).
				NewLine().Text("兑换成功！获得24张转蛋券").Send()
			dbTransition.AddUser(update.Message.From.ID, update.GroupID)
			dbTransition.IncreaseUserTicket(update.Message.From.ID, 24)
		}()
	} else {
		go func() {
			dbTransition.AddUser(update.Message.From.ID, update.GroupID)
			if !dbTransition.DetectDailyLimit(update.Message.From.ID) {
				dbTransition.IncreaseUserTicket(update.Message.From.ID, 1)
			}
		}()
	}
	list := strings.Split(update.Message.Text, " ")
	if len(list) == 2 && list[0] == "查询" {
		bot.NewMessage(update.GroupID, "group").
			Text(dbTransition.DisplaySingleHomoInfo(list[1])).Send()
	}
}

func PrintHelpInfo(groupID int64) {
	msg := bot.NewMessage(groupID, "group").Text("")
	for index, api := range keyWords {
		msg = msg.Text(strconv.Itoa(index+1) + "." + api)
		msg = msg.NewLine()
	}
	msg.Text("12.查询 角色名").NewLine().
		Text("复制随机产生的CDK并发送即可获得24张转蛋券哦").Send()
}

func SendRandomCDK() {
	cdk := common.GetCDK()
	for group := range util.GetObjectByKey("group-enable-homo-space").(map[int64]bool) {
		CDKMap[cdk] = group
		bot.NewMessage(group, "group").Text("野生的CDK出现了！").NewLine().
			Text(cdk).Send()
	}
}

func Time2SendCDK() {
	for {
		rand.Seed(time.Now().UnixNano())
		duration := time.Hour*time.Duration(rand.Intn(2)) + time.Minute*time.Duration(rand.Intn(40)+20)
		//time.Minute * time.Duration(rand.Intn(10) + 10)
		timer := time.NewTimer(duration)
		<-timer.C
		SendRandomCDK()
	}
}
