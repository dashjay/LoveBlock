package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/importcjj/sensitive"
	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/cache"
	"github.com/silenceper/wechat/message"
	"main/core"
	"main/database"
	"main/env"
	"strings"
)

var filter *sensitive.Filter
var wc *wechat.Wechat
var config *wechat.Config
var ValidChan chan MessageQueue
var InvalidChan chan MessageQueue

var LastBlock core.Block

type MessageQueue struct {
	Content string
	OpenID  string
}

func init() {

	filter = sensitive.New()

	err := filter.LoadWordDict("./dic.txt")
	if err != nil {
		panic(err)
	}

	env.Init()

	err = database.InitMongoDB()
	if err != nil {

		panic(err)

	}

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	// 开始状态
	err = con.Find(nil).Sort("-timestamp").One(&LastBlock)
	if err != nil {
		b := core.NewGenesisBlock()
		con.Insert(b.NewBIMFromBlock())
		LastBlock = *b
	}

	ValidChan = make(chan MessageQueue, 64)

	InvalidChan = make(chan MessageQueue, 64)

	redisCache := cache.NewRedis(&cache.RedisOpts{
		Host:        "0.0.0.0",
		Database:    1,
		MaxIdle:     100,
		MaxActive:   100,
		IdleTimeout: 0,
	})

	//配置微信参数
	config = &wechat.Config{
		AppID:     env.AppID,
		AppSecret: env.AppSecret,
		Token:     "wechat",
		Cache:     redisCache,
	}

	wc = wechat.NewWechat(config)

}

func newTextMessage(s string) *message.Reply {
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(s)}
}

func index(ctx *context.Context) {

	server := wc.GetServer(ctx.Request, ctx.ResponseWriter)

	server.SetMessageHandler(func(v message.MixMessage) *message.Reply {

		switch v.MsgType {
		//文本消息
		case message.MsgTypeText:
			//do something

			if len(v.Content) > 3 {

				if strings.Contains(v.Content[0:3], "to") || strings.Contains(v.Content[0:3], "To") {
					return PostBlock(v)
				}
				if strings.Contains(v.Content, "查看表白") {

					return GetOneBlock()
				}
				if strings.Contains(v.Content, "getnext") {

					return GetTargetBlock(v)
				}
				if strings.Contains(v.Content, "pass") {
					return PassBlock(v)
				}
			} else {
				if v.Content == "~!@" {
					return GetInvalidBlock()
				}
			}

			//图片消息
		case message.MsgTypeImage:
			//do something

			//语音消息
		case message.MsgTypeVoice:
			//do something

			//视频消息
		case message.MsgTypeVideo:
			//do something

			//小视频消息
		case message.MsgTypeShortVideo:
			//do something

			//地理位置消息
		case message.MsgTypeLocation:
			//do something

			//链接消息
		case message.MsgTypeLink:
			//do something

			//事件推送消息
		case message.MsgTypeEvent:
			switch v.Event {

			//EventSubscribe 订阅
			case message.EventSubscribe:
				//do something

				//取消订阅
			case message.EventUnsubscribe:
				//do something

				//用户已经关注公众号，则微信会将带场景值扫描事件推送给开发者
			case message.EventScan:
				//do something

				// 上报地理位置事件
			case message.EventLocation:
				//do something

				// 点击菜单拉取消息时的事件推送
			case message.EventClick:
				//do something

				// 点击菜单跳转链接时的事件推送
			case message.EventView:
				//do something

				// 扫码推事件的事件推送
			case message.EventScancodePush:
				//do something

				// 扫码推事件且弹出“消息接收中”提示框的事件推送
			case message.EventScancodeWaitmsg:
				//do something

				// 弹出系统拍照发图的事件推送
			case message.EventPicSysphoto:
				//do something

				// 弹出拍照或者相册发图的事件推送
			case message.EventPicPhotoOrAlbum:
				//do something

				// 弹出微信相册发图器的事件推送
			case message.EventPicWeixin:
				//do something

				// 弹出地理位置选择器的事件推送
			case message.EventLocationSelect:
				//do something

			}
		}

		return &message.Reply{MsgType: message.MsgTypeText, MsgData: "unknown message"}
	})

	//处理消息接收以及回复
	err := server.Serve()

	if err != nil {

		fmt.Println(err)

		return
	}

	//createButton(wc)

	//发送回复的消息
	server.Send()

}

func main() {

	// 创建两个Session
	s1 := database.NewSessionStore()
	s2 := database.NewSessionStore()
	defer func() {
		s1.Close()
		s2.Close()
	}()

	// 两个消息队列
	go Lover(&ValidChan, s1)
	go InValidLover(&InvalidChan, s2)

	beego.Any("/", index)
	beego.Any("/get", Get)
	beego.Any("/loadmore", LoadMore)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	beego.Run(":8001")

}

func Lover(c *chan MessageQueue, s *database.SessionStore) {
	for {

		select {

		case data := <-*c:
			{

				b := core.NewBlock(data.Content, LastBlock.Hash, data.OpenID)

				err := s.C("blocks").Insert(b.NewBIMFromBlock())

				if err == nil {
					LastBlock = *b
				}
			}
		}
	}
}

func InValidLover(c *chan MessageQueue, s *database.SessionStore) {
	for {

		select {

		case data := <-*c:
			{
				b := core.NewBlock(data.Content, LastBlock.Hash, data.OpenID)
				_ = s.C("invalid_blocks").Insert(b.NewBIMFromBlock())
			}
		}
	}
}
