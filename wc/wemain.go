package wc

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/message"
	"github.com/silenceper/wechat/server"
	"main/env"
	"strings"
)

var wc *wechat.Wechat

func init() {

	//配置微信参数
	config := &wechat.Config{
		AppID:     env.AppID,
		AppSecret: env.AppSecret,
		Token:     "wechat",
	}
	wc = wechat.NewWechat(config)
}
func GetWc(ctx *context.Context) *server.Server {
	return wc.GetServer(ctx.Request, ctx.ResponseWriter)
}

func newTextMessage(s string) *message.Reply {
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(s)}
}


func Index(ctx *context.Context) {

	server := GetWc(ctx)

	server.SetMessageHandler(func(v message.MixMessage) *message.Reply {

		switch v.MsgType {
		//文本消息
		case message.MsgTypeText:
			//do something

			if len(v.Content) > 3 {

				// 表白
				if strings.Contains(v.Content[0:3], "to") || strings.Contains(v.Content[0:3], "To") {
					return PostBlock(v)
				}
				// 查看表白
				if strings.Contains(v.Content, "查看表白") {
					return GetOneBlock()
				}
				if strings.Contains(v.Content[0:10], "search") {
					return SearchBlock(v)
				}

				// 获取下一条
				if strings.Contains(v.Content, "getnext") {
					return GetTargetBlock(v)
				}
				// 通过表白
				if strings.Contains(v.Content, "pass") {
					return PassBlock(v)
				}
			} else {
				if v.Content == "~!@" {
					return GetInvalidBlock()
				}
				if v.Content == "oid" {

					return newTextMessage("openid是:" + server.GetOpenID())
				}
			}

			return newTextMessage("没有对应的指令，请点击下方指示回复\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=查看表白'>查看最新表白~</a>\n 想要表白回复To + 表白内容。例如：To 最后一次表白ZFQ小姐姐了，😭😢")

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
