package main

import (
	"database/sql"
	"log"
	"strconv"
	
	"CQApp/src/dbTransition"
	"CQApp/src/homo"
	"CQApp/src/lottery"
	"github.com/catsworld/qq-bot-api"
	_ "github.com/go-sql-driver/mysql"
)

var bot  *qqbotapi.BotAPI
var db   *sql.DB
var keyWords = [9]string {
	"转蛋单抽", "转蛋十连", "转蛋奖池", "HOMOSPACE", "编辑HOMO", "我的转蛋券", "HOMO图鉴", "准备对战", "Document",
}

func main() {
	var err error
	bot, err = qqbotapi.NewBotAPI("", "ws://39.106.219.180:6700", "CQHTTP_SECRET")
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	
	conf := qqbotapi.NewUpdate(0)
	conf.PreloadUserInfo = true
	updates, err := bot.GetUpdatesChan(conf)
	
	db, err = sql.Open("mysql",
		"root:password@/homospace?charset=utf8")
	//连接数据库，格式 用户名：密码@/数据库名？charset=编码方式
	if err != nil {
		log.Println(err)
		panic("open database-MySql failed.")
	}
	defer db.Close()
	
	dbTransition.Init(db)
	lottery.Init(bot)
	homo.Init(bot)
	
	for update := range updates {
		if update.Message == nil || update.MessageType != "group" {
			continue
		}
		log.Println(update.Message)
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
			addMissionChan := make(chan struct{}, 1)
			go homo.Prepare4Battle(
				updates, addMissionChan,
				update.Message.From.ID,
				update.GroupID,
			)
			<-addMissionChan
			break
		case 8:
			PrintHelpInfo(update.GroupID)
			break
		default:
			handleMsg(update)
			break
		}
	}
}

func handleMsg(update qqbotapi.Update) {
	if update.MessageType == "group" || update.GroupID == 930378083 {
		go func() {
			dbTransition.AddUser(update.Message.From.ID)
			if !dbTransition.DetectDailyLimit(update.Message.From.ID) {
				dbTransition.IncreaseUserTicket(update.Message.From.ID, 1)
			}
		}()
	}
}

func PrintHelpInfo(groupID int64) {
	msg := bot.NewMessage(groupID, "group").Text("")
	for index, api := range keyWords {
		msg = msg.Text(strconv.Itoa(index+1)+"."+api)
		if index != len(keyWords)-1 {
			msg = msg.NewLine()
		}
	}
	msg.Send()
}