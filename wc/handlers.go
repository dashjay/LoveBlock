package wc

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/kataras/iris/v12"
	"github.com/silenceper/wechat/message"
	"gopkg.in/mgo.v2/bson"
	"main/MQ"
	"main/blocks"
	"main/database"
	"main/filter"
	"strings"
	"time"
)

var lock = false
var global []blocks.BlockInMongo = nil

func GetOneBlock() *message.Reply {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	var b blocks.BlockInMongo
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

	var b blocks.BlockInMongo

	err := con.Find(bson.M{"prev_block_hash": temp[1]}).One(&b)

	if err != nil {
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("可能这是最后一条表白，或者有表白被人篡改了。")}
	}

	str := "内容" + b.Data + fmt.Sprintf("\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=getnext %s'>点击获取下一跳表白</a>", b.Hash)

	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(str)}
}

func PostBlock(v message.MixMessage) *message.Reply {

	var f = filter.GetFilter()
	f.UpdateNoisePattern(`\*`)
	if b, _ := f.Validate(v.Content); !b {
		MQ.InvalidChan <- MQ.MessageQueue{
			Content: v.Content,
			OpenID:  v.OpenID,
		}
		return newTextMessage("你的表述中包含敏感词,不能立刻显示,服务器正在处理，请稍后查看")
	}

	f.UpdateNoisePattern(`x`)
	if b, _ := f.Validate(v.Content); !b {
		MQ.InvalidChan <- MQ.MessageQueue{
			Content: v.Content,
			OpenID:  v.OpenID,
		}
		return newTextMessage(fmt.Sprintf("你的表述中包含敏感词,不能立刻显示,服务器正在处理，请稍后查看"))
	}

	MQ.ValidChan <- MQ.MessageQueue{
		Content: v.Content,
		OpenID:  v.OpenID,
	}

	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("接收到你的表白了~\n<a href='http://114.55.92.2:8002/'>点击查看</a>")}
}

func GetInvalidBlock() *message.Reply {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("invalid_blocks")

	var b []blocks.BlockInMongo

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

	var f = filter.GetFilter()
	var b blocks.BlockInMongo
	if err := con.Find(bson.M{"hash": temp[1]}).One(&b); err != nil {
		return newTextMessage(err.Error())
	}
	if err := con.Remove(bson.M{"hash": temp[1]}); err != nil {
		return newTextMessage(err.Error())
	}
	MQ.ValidChan <- MQ.MessageQueue{
		Content: f.Replace(b.Data, '*'),
		OpenID:  b.OpenID,
	}
	return newTextMessage("通过内容" + b.Data)
}

func Get(ctx *context.Context) {

	var k = blocks.GetLastBlock().NewBIMFromBlock()

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

	var temp blocks.BlockInMongo

	temp.Hash = hash
	var res []blocks.BlockInMongo
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

func SearchBlock(v message.MixMessage) *message.Reply {
	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	temp := strings.Split(v.Content, " ")
	if len(temp) != 2 {
		return newTextMessage("搜索请回复：「search 内容」,中间只包含一个空格")
	}
	var b []blocks.BlockInMongo
	err := con.Find(bson.M{"data": bson.M{"$regexp": temp[1]}}).Limit(15).All(&b)
	if err != nil {
		return newTextMessage("没有相关信息，msg_detail:" + err.Error())
	}
	var buf strings.Builder

	buf.WriteString("---start---\n")
	for _, bb := range b {
		buf.WriteString(bb.Formatter())
	}
	return newTextMessage(buf.String())
}
