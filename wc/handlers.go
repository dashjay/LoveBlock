package wc

import (
	"fmt"
	"github.com/astaxie/beego/context"
	ai "github.com/night-codes/mgo-ai"
	"gopkg.in/mgo.v2/bson"
	"main/MQ"
	"main/blocks"
	"main/database"
	"main/filter"
	"main/protos"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var lock = false
var global []blocks.BlockInMongo = nil

func UrlButton(msg, content string) string {
	return fmt.Sprintf("<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=%s'>%s</a>", msg, content)
}

// 获取最新的一个block
func GetOneBlock() string {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C(database.DBBlocks)

	var b blocks.BlockInMongo
	err := con.Find(nil).Sort("-timestamp").One(&b)
	if err != nil {

		return err.Error() + menu
	} else {

		return strings.Join([]string{
			"内容:\n----------\n",
			b.Data,
			"\n----------\n",
			UrlButton("getprev "+b.PrevBlockHash, "获取前一条表白"),
			menu}, "")
	}
}

func GetTargetBlock(hash string) string {
	temp := strings.TrimSpace(hash)

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	var b blocks.BlockInMongo

	err := con.Find(bson.M{"hash": temp}).One(&b)

	if err != nil {
		return "没有这条表白"
	}

	var str strings.Builder
	str.WriteString("内容:\n----------\n")
	str.WriteString(b.Data)
	str.WriteString("\n----------\n")
	str.WriteString(fmt.Sprintf("时间:%s", time.Unix(b.Timestamp, 0).Format("2006-01-02 15:04:05")))
	str.WriteString(fmt.Sprintf("\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=getprev %s'>获取前一条</a>", b.PrevBlockHash))
	str.WriteString(menu)
	return str.String()
}

func PostBlock(content, oid string) string {

	var f = filter.GetFilter()
	f.UpdateNoisePattern(`\*`)
	if b, _ := f.Validate(content); !b {
		MQ.InvalidChan <- MQ.MessageQueue{
			Content: content,
			OpenID:  oid,
		}
		return rReceiveLoveInNormal
	}

	f.UpdateNoisePattern(`x`)
	if b, _ := f.Validate(content); !b {
		MQ.InvalidChan <- MQ.MessageQueue{
			Content: content,
			OpenID:  oid,
		}
		return rReceiveLoveInNormal
	}

	MQ.ValidChan <- MQ.MessageQueue{
		Content: content,
		OpenID:  oid,
	}

	return rReceiveLoveNormal
}

func GetInvalidBlock() (d protos.BaseMessage) {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("invalid_blocks")

	var b []blocks.BlockInMongo = nil

	err := con.Find(nil).Limit(20).All(&b)

	if err != nil || len(b) == 0 || b == nil {
		d.Type = "text"
		d.Data = []byte("没有待审核的内容")
		return d
	}

	var buf strings.Builder

	for _, k := range b {

		buf.WriteString(k.Data)
		buf.WriteString(fmt.Sprintf("\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=pass %s'>点击通过</a>\n", k.Hash))

	}

	d.Type = "text"
	d.Data = []byte(buf.String())
	return d

}

func PassBlock(hash string) string {
	temp := strings.TrimSpace(hash)
	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("invalid_blocks")

	var f = filter.GetFilter()
	var b blocks.BlockInMongo
	if err := con.Find(bson.M{"hash": temp}).One(&b); err != nil {
		return err.Error()
	}
	if err := con.Remove(bson.M{"hash": temp}); err != nil {
		return err.Error()
	}
	MQ.ValidChan <- MQ.MessageQueue{
		Content: f.Replace(b.Data, '*'),
		OpenID:  b.OpenID,
	}
	return "通过内容" + b.Data
}

func SearchBlock(keyword string) string {
	temp := strings.TrimSpace(keyword)

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")
	var b []blocks.BlockInMongo
	err := con.Find(bson.M{"data": bson.RegEx{Pattern: temp, Options: "$i"}}).Limit(15).All(&b)
	if err != nil {
		return fmt.Sprintf("没有相关信息，msg_detail:%s", err.Error())
	}

	var buf strings.Builder
	buf.WriteString("---start---\n")
	for _, bb := range b {
		buf.WriteString(bb.Formatter())
	}
	buf.WriteString("---end---")

	return fmt.Sprintf("%s%s", buf.String(), menu)
}

func RandomBlock(num string) string {
	// 处理获得数字

	var e = "随机请回复：「随机查看 数字」,中间只包含一个空格，数字小于20"

	temp := strings.TrimSpace(num)

	size, err := strconv.Atoi(temp)
	if err != nil {
		return e
	}
	if size > 20 {
		return e
	}
	// 回去block
	b := GetRandomBlock(size)
	var buf strings.Builder

	buf.WriteString("---start---\n")
	for _, bb := range b {
		buf.WriteString(bb.Formatter())
	}
	buf.WriteString("---end---")

	buf.WriteString(menu)

	return buf.String()
}

func AddLike(id, oid string) string {
	// 将指令split成两个部分
	id = strings.TrimSpace(id)

	i, err := strconv.Atoi(id)

	if err != nil {
		return "添加喜欢失败: " + err.Error()
	}

	// 创建数据库连接
	ds := database.NewSessionStore()
	defer ds.Close()
	// 连接counter
	ai.Connect(ds.C(database.DBCounters))

	// 切换为Block
	con := ds.C(database.DBBlocks)
	var b blocks.BlockInMongo
	err = con.Find(bson.M{"id": i}).One(&b)
	if err != nil {
		return fmt.Sprintf("点赞第%d条表白失败，失败原因%s,%s", i, err.Error(), menu)
	}
	// 切换为object
	con = ds.C(database.DBObjects)
	var selecter = bson.M{"bid": b.ID, "open_id": oid, "type": "like"}
	if c, err := con.Find(selecter).Count(); err != nil {
		return fmt.Sprintf("点赞第%d条表白失败，失败原因%s，%s", i, err.Error(), menu)
	} else {
		if c != 0 {
			return fmt.Sprintf("您已经表白过该条表白~\n点击查看更多表白%s", menu)
		}
	}

	o := blocks.Object{
		BID:       b.ID,
		Type:      "like",
		ID:        ai.Next("like"),
		OpenID:    oid,
		TimeStamp: time.Now().Unix(),
	}

	con.Insert(&o)

	n, _ := con.Find(bson.M{"bid": b.ID, "type": "like"}).Count()

	return fmt.Sprintf("点赞成功，当前共有%d个用户点赞%s", n, menu)
}

func ConvertForFront(ori []blocks.BlockInMongo) []blocks.BlockFront {

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C(database.DBObjects)
	var temp []blocks.BlockFront
	// 原数组
	for _, k := range ori {
		var t = k.ConvertToBFront()
		// 喜欢数
		t.LikeNum, _ = con.Find(bson.M{"bid": k.ID, "type": "like"}).Count()
		// 回复数
		t.ReplyNum, _ = ds.C(database.DBBlocks).Find(bson.M{"data": bson.RegEx{Pattern: fmt.Sprintf("回复表白%d", k.ID), Options: "im",}}).Count()

		if strings.Contains(t.Data, "回复表白") {
			fmt.Println("回复表白")
			c := regexp.MustCompile(`^回复表白([\d])`)
			res := c.FindStringSubmatch(t.Data)
			var k blocks.BlockInMongo
			i, e := strconv.Atoi(res[1])
			if e != nil {
				continue
			}
			ds.C(database.DBBlocks).Find(bson.M{"id": i}).One(&k)
			t.ReplyTarget = fmt.Sprintf("#%d %s", k.ID, k.Data)
			fmt.Println(t.ReplyTarget)
		}

		t.ReplyTarget = ""
		temp = append(temp, t)
	}
	return temp
}

func Get(ctx *context.Context) {

	var k = blocks.GetLastBlock().ConvertToBlockInMongo()

	if global != nil {
		for _, s := range global {
			if s.Hash == k.Hash {
				// 没有更新的表白
				var res = map[string]interface{}{"status": 0, "content": ConvertForFront(global)}
				ctx.Output.JSON(res, false, false)
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

	ctx.Output.JSON(map[string]interface{}{"status": 1, "content": ConvertForFront(global)}, false, false)

	return
}

func LoadMore(ctx *context.Context) {

	hash := ctx.Input.Query("hash")

	if hash == "" {
		ctx.Output.JSON(map[string]interface{}{"status": 0, "content": nil}, false, false)
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
	ctx.Output.JSON(map[string]interface{}{"status": 0, "content": res}, false, false)

}

func Random(ctx *context.Context) {

	size := ctx.Input.Query("size")
	if size == "" {
		ctx.Output.JSON(map[string]interface{}{"status": 0}, false, false)
		return
	}
	sint, err := strconv.Atoi(size)
	if err != nil {
		ctx.Output.JSON(map[string]interface{}{"status": 0, "msg": err.Error()}, false, false)
		return
	}

	if sint <= 0 {
		ctx.Output.JSON(map[string]interface{}{"status": 0, "msg": "size should be  bigger than 0"}, false, false)
		return
	}

	r := GetRandomBlock(sint)

	ctx.Output.JSON(map[string]interface{}{"status": 1, "msg": "ok", "content": ConvertForFront(r)}, false, false)
}

func GetRandomBlock(sint int) []blocks.BlockInMongo {
	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C(database.DBBlocks)
	var b []blocks.BlockInMongo
	_ = con.Pipe([]bson.D{{{"$sample", bson.D{{"size", sint}}}}}).All(&b)
	return b
}
