package main

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/kataras/iris/v12"
	"github.com/silenceper/wechat/message"
	"gopkg.in/mgo.v2/bson"
	"main/core"
	"main/database"
	"strings"
	"time"
)

var lock = false
var global []core.BlockInMongo = nil

func GetOneBlock() *message.Reply {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	var b core.BlockInMongo
	err := con.Find(nil).Sort("timestamp").One(&b)
	if err != nil {
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(err.Error())}
	} else {
		str := "内容" + b.Data + fmt.Sprintf("\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=getnext %s'>点击获取下一跳表白</a>", b.Hash)
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(str)}
	}
}

func GetTargetBlock(v message.MixMessage) *message.Reply {
	temp := strings.Split(v.Content, " ")

	if len(temp) != 2 {
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("表白不存在")}
	}

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	var b core.BlockInMongo

	err := con.Find(bson.M{"prev_block_hash": temp[1]}).One(&b)

	if err != nil {
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("表白不存在")}
	}

	str := "内容" + b.Data + fmt.Sprintf("\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=getnext %s'>点击获取下一跳表白</a>", b.Hash)

	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(str)}
}

func PostBlock(v message.MixMessage) *message.Reply {

	filter.UpdateNoisePattern(`\*`)
	if b, _ := filter.Validate(v.Content); !b {
		InvalidChan <- MessageQueue{
			Content: v.Content,
			OpenID:  v.OpenID,
		}
		return newTextMessage("你的表述中包含敏感词,不能立刻显示,服务器正在处理，请稍后查看")
	}

	filter.UpdateNoisePattern(`x`)
	if b, _ := filter.Validate(v.Content); !b {
		InvalidChan <- MessageQueue{
			Content: v.Content,
			OpenID:  v.OpenID,
		}
		return newTextMessage(fmt.Sprintf("你的表述中包含敏感词,不能立刻显示,服务器正在处理，请稍后查看"))
	}

	ValidChan <- MessageQueue{
		Content: v.Content,
		OpenID:  v.OpenID,
	}

	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("接收到你的表白了~")}
}

func GetInvalidBlock() *message.Reply {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("invalid_blocks")

	var b []core.BlockInMongo

	err := con.Find(nil).Limit(20).All(&b)

	if err != nil {
		return newTextMessage(fmt.Sprintf("出现错误%s，请检查", err.Error()))
	}

	var buf strings.Builder

	for _, k := range b {
		buf.WriteString(k.Data)
		buf.WriteString(fmt.Sprintf("\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=pass %s'>点击通过</a>\n", k.Hash))
	}

	return newTextMessage(fmt.Sprintf(buf.String()))
}

func PassBlock(v message.MixMessage) *message.Reply {
	temp := strings.Split(v.Content, " ")
	if len(temp) != 2 {
		return newTextMessage("出现错误split后不等于2")
	}

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("invalid_blocks")

	var b core.BlockInMongo
	if err := con.Find(bson.M{"hash": temp[1]}).One(&b); err != nil {
		return newTextMessage(err.Error())
	}
	if err := con.Remove(bson.M{"hash": temp[1]}); err != nil {
		return newTextMessage(err.Error())
	}
	ValidChan <- MessageQueue{
		Content: filter.Replace(b.Data, '*'),
		OpenID:  b.OpenID,
	}
	return newTextMessage("通过内容" + b.Data)
}

func Get(ctx *context.Context) {

	var k = LastBlock.NewBIMFromBlock()

	if global != nil {
		for _, s := range global {
			if s.Hash == k.Hash {
				// 没有更新的表白

				ctx.Output.JSON(iris.Map{"status": 1, "content": global}, false, false)
				return
			}
		}
		// 有更新的表白
	}

	global = nil
	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	for lock {
		time.Sleep(time.Millisecond)
	}
	lock = true
	con.Find(nil).Sort("-timestamp").Limit(20).All(&global)
	lock = false

	ctx.Output.JSON(iris.Map{"status": 1, "content": global}, false, false)
	return
}

func LoadMore(ctx *context.Context) {

	hash := ctx.Input.Query("hash")

	if hash == "" {
		ctx.Output.JSON(iris.Map{"status": 0, "content": nil}, false, false)
		return
	}

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	var temp core.BlockInMongo

	temp.Hash = hash
	var res []core.BlockInMongo
	var i = 20
	for i > 0 {

		err := con.Find(bson.M{"prev_block_hash": temp.Hash}).One(&temp)
		if err != nil {
			break
		}

		res = append(res, temp)
		i--
	}
	ctx.Output.JSON(iris.Map{"status": 0, "content": res}, false, false)

}
