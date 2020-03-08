package homo

import (
	"log"
	"strconv"
	"strings"
	
	"CQApp/src/dbTransition"
	"github.com/catsworld/qq-bot-api"
)

var aliasBot *qqbotapi.BotAPI

type BattleGround struct {
	HostID      int64
	GuestID     int64
	RoomID      int
}


func Init(bot *qqbotapi.BotAPI) {
	if bot == nil {
		panic("failed to bind a nil value")
	}
	aliasBot = bot
}

func EditHomo(
	updates qqbotapi.UpdatesChannel,
	syncChannel chan struct{},
	fromQQ      int64,
	fromGroup   int64,
) {
	aliasBot.NewMessage(fromGroup, "group").
		At(strconv.FormatInt(fromQQ, 10)).
		NewLine().
		Text("可以开始编辑HOMO信息了,使用 帮助 查看说明").Send()
	
	for update := range updates {
		if update.Message.From.ID != fromQQ || update.GroupID != fromGroup {
			continue
		}
		if update.Message.Text == "quit" {
			aliasBot.NewMessage(fromGroup, "group").
				At(strconv.FormatInt(fromQQ, 10)).NewLine().
				Text("编辑HOMO结束操作完成").Send()
			syncChannel <- struct{}{}
			return
		} else if update.Message.Text == "删除HOMO" {
			aliasBot.NewMessage(fromGroup, "group").Text("好啊来啊！输入：名称(quit返回上一层)").Send()
			for update := range updates {
				if update.Message.From.ID != fromQQ || update.GroupID != fromGroup {
					continue
				}
				if update.Message.Text == "quit" {
					aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
						NewLine().Text("返回上一层").Send()
					break
				} else {
					db := dbTransition.GetConn()
					_, err := db.Exec("DELETE FROM HOMO WHERE NAME=?", update.Message.Text)
					if err != nil {
						aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
							NewLine().Text("操作失败, "+err.Error()).Send()
					} else {
						aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
							NewLine().Text("操作成功").Send()
					}
				}
			}
		} else if update.Message.Text == "添加HOMO" {
			aliasBot.NewMessage(fromGroup, "group").Text("好啊来啊！输入：名称 稀有度(N,SR,UR,quit返回上一层)").Send()
			for update := range updates {
				if update.Message.From.ID != fromQQ || update.GroupID != fromGroup {
					continue
				}
				if update.Message.Text == "quit" {
					aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
						NewLine().Text("返回上一层").Send()
					break
				} else if len(update.Message.Text)>3 {
					list := strings.Split(update.Message.Text, " ")
					if len(list) == 2 {
						if list[1] != "N" && list[1] != "UR" && list[1] != "SR" {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作失败, 不支持的稀有度").Send()
						}
						db := dbTransition.GetConn()
						_, err := db.Exec("INSERT INTO HOMO(NAME,RARE) VALUES(?,?)", list[0], list[1])
						if err != nil {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作失败, "+err.Error()).Send()
						} else {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作成功").Send()
						}
					}
				}
			}
			
			
		} else if update.Message.Text == "修改属性" {
			aliasBot.NewMessage(fromGroup, "group").Text("好啊来啊！输入：名称 属性名 值(quit返回上一层) 使用 帮助 查看说明").Send()
			for update := range updates {
				if update.Message.From.ID != fromQQ || update.GroupID != fromGroup {
					continue
				}
				if update.Message.Text == "quit" {
					aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
						NewLine().Text("返回上一层").Send()
					break
				} else if update.Message.Text == "帮助" {
					aliasBot.NewMessage(fromGroup, "group").
						Text("常用属性: ").NewLine().
						Text("NAME(名称) DESCRIPTION(描述) RARE(稀有度) ACQUIRABLE(可否抽卡获得) PROB_UP(是否概率UP)").NewLine().
						Text("战斗属性: INITIAL_后跟HP/ATN/INT/DEF/RES/SPD/LUK 代表初始属性").NewLine().
						Text("GROWTH_后跟HP/ATN/INT/DEF/RES/SPD/LUK 代表升级时的成长属性(会获得level加成)").Send()
				} else if len(update.Message.Text)>3 {
					list := strings.Split(update.Message.Text, " ")
					if len(list) == 3 {
						if len(list[2]) > 5 && list[1] != "DESCRIPTION" {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作失败, 不安全的长度").Send()
							continue
						}
						db := dbTransition.GetConn()
						_, err := db.Exec("UPDATE HOMO SET "+list[1]+"=? WHERE NAME=?", list[2], list[0])
						if err != nil {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作失败, "+err.Error()).Send()
						} else {
							aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
								NewLine().Text("操作成功").Send()
						}
					} else {
						aliasBot.NewMessage(fromGroup, "group").At(strconv.FormatInt(fromQQ,10)).
							NewLine().Text("参数错误，操作废弃").Send()
					}
				}
			}
		} else if update.Message.Text == "帮助" {
			aliasBot.NewMessage(fromGroup, "group").
				At(strconv.FormatInt(fromQQ, 10)).
				NewLine().
				Text("1.删除HOMO，2.添加HOMO，3.修改属性").NewLine().
				Text("删除HOMO时输入该指令后输入HOMO的ID或名称即可").NewLine().
				Text("添加HOMO可暂时先输入Name(varchar) Rare(varchar)，用空格隔开，之后进行修改属性即可").NewLine().
				Text("修改属性需要发送 [属性名 值]（不带方括号）格式的消息，发送quit退出编辑").NewLine().
				Text("该功能目前一次只有一人能够使用").Send()
		}
	}
}

func DisplayAsset(update qqbotapi.Update) {
	data := dbTransition.GetOnesAsset(update.Message.From.ID)
	msg := aliasBot.NewMessage(update.GroupID, "group").
		At(strconv.FormatInt(update.Message.From.ID, 10))
	if len(data) == 0 {
		msg.NewLine().Text("恁麾下还妹有任何HOMO哦").Send()
	} else {
		for _, dat := range data {
			msg = msg.NewLine().Text(dat)
		}
		msg.Send()
	}
}

func DisplayAllHomo(groupID int64) {
	db := dbTransition.GetConn()
	var count int64
	err := db.QueryRow("SELECT count(*) FROM HOMO").Scan(&count)
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	if count < 1 {
		aliasBot.NewMessage(groupID, "group").Text("当前图鉴内还没有HOMO哦").Send()
	} else {
		rows, err := db.Query("SELECT ID,NAME,DESCRIPTION,RARE FROM HOMO")
		if err != nil {
			aliasBot.NewMessage(groupID, "group").
				NewLine().Text("操作失败, "+err.Error()).Send()
		} else {
			msg := aliasBot.NewMessage(groupID, "group").Text("")
			for rows.Next() {
				var ID          int64
				var Name        string
				var Description string
				var Rare        string
				err = rows.Scan(&ID, &Name, &Description, &Rare)
				if err != nil {
					aliasBot.NewMessage(groupID, "group").
						NewLine().Text("操作失败, "+err.Error()).Send()
				}
				msg = msg.Text("No"+strconv.FormatInt(ID,10)+":").NewLine().
					Text(Name+"["+Rare+"]: "+Description).NewLine()
			}
			msg.Send()
		}
	}
}

func Prepare4Battle(
	updates     qqbotapi.UpdatesChannel,
	fromQQ      int64,
	fromGroup   int64,
) {
	aliasBot.NewMessage(fromGroup, "group").Text("指令：").
		At(strconv.FormatInt(fromQQ, 10)).NewLine().
		Text("创建房间 房间号").NewLine().
		Text("加入房间 房间号").Send()
	//for
}
