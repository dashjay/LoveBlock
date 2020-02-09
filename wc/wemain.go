package wc

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/boltdb/bolt"
	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/message"
	"github.com/silenceper/wechat/server"
	"main/env"
	"main/session"
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

	s := GetWc(ctx)

	s.SetMessageHandler(func(v message.MixMessage) *message.Reply {

		switch v.MsgType {

		// if文本消息
		case message.MsgTypeText:

			// 获取内容长度
			var length = len(v.Content)

			// 获取openid
			var oid = s.GetOpenID()

			// OpenidError
			if len(oid) != 28 {
				return newTextMessage("openid error: " + s.GetOpenID())
			}

			switch {

			case length == 1:
				{
					switch v.Content {
					case "1":
						{
							user := GetUser(oid)
							user.SetMode(LoveMode)
							return newTextMessage("成功切换为表白墙模式")
						}
					case "2":
						{
							user := GetUser(oid)
							user.SetMode(ResourcesMode)
							return newTextMessage("成功切换为资源模式")
						}
					}
					return newTextMessage("其他功能开发中")
				}
			case length >= 2:
				{
					var u = GetUser(oid)

					switch u.Mode {

					case PreLoveMode:
						{

							if v.Content == "Confirm" {
								if u.CTX[PreLoveMode].(string) == "" {
									return newTextMessage("当前缓冲区为空，请回复内容尝试表白哟")
								}
								if len(u.CTX[PreLoveMode].(string)) > 3*520 {
									return newTextMessage("当前缓冲区文字超过520个字")
								}
								u.SetMode(LoveMode)
								return PostBlock(u.CTX[PreLoveMode].(string), oid)
							}

							if v.Content == "Cancel" {
								u.CTX[PreLoveMode] = ""
								u.SetMode(LoveMode)
								return newTextMessage("「已取消」爱就要大声说出来~！")
							}

							u.CTX[PreLoveMode] = v.Content
							return newTextMessage(
								"确定表白的内容为：\n\n" + v.Content + "\n\n  吗？\n " +
									UrlButton("Cancel", "取消表白") + "\t " +
									UrlButton("Confirm", "确定发送"))
						}
					case LoveMode:
						{

							if v.Content == "~!@#" {
								return GetInvalidBlock()
							}

							if v.Content == "表白" {
								u.SetMode(PreLoveMode)
								return newTextMessage("请直接回复积极正向的表白(520字内), 寻人或求偶信息 或 \n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=取消表白'>点我取消表白</a>")
							}

							// 查看表白
							if strings.Contains(v.Content, "查看表白") {
								return GetOneBlock()
							}
							// 通过表白
							if strings.Contains(v.Content, "pass") {
								return PassBlock(v)
							}
							// 获取下一条
							if strings.Contains(v.Content, "getnext") {
								return GetTargetBlock(v)
							}
							if strings.Contains(v.Content[0:6], "search") {
								return SearchBlock(v)
							}
						}
					case ResourcesMode:
						{

						}
					case NoneMode:
						{
							return newTextMessage("帮助信息")
						}

					}

				}
				return newTextMessage("没有对应的指令，请点击下方指示回复\n\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=查看表白'>查看最新表白~</a>\n 想要表白回复To + 表白内容。例如：To 最后一次表白ZFQ小姐姐了，😭😢")

				//	//图片消息
				//case message.MsgTypeImage:
				//	//do something
				//
				//	//语音消息
				//case message.MsgTypeVoice:
				//	//do something
				//
				//	//视频消息
				//case message.MsgTypeVideo:
				//	//do something
				//
				//	//小视频消息
				//case message.MsgTypeShortVideo:
				//	//do something
				//
				//	//地理位置消息
				//case message.MsgTypeLocation:
				//	//do something
				//
				//	//链接消息
				//case message.MsgTypeLink:
				//	//do something
			}
		//事件推送消息
		case message.MsgTypeEvent:
			switch v.Event {
			//
			//EventSubscribe 订阅
			case message.EventSubscribe:

				return newTextMessage(rOnSubscribe)

			//取消订阅
			case message.EventUnsubscribe:

				oid := s.GetOpenID()
				u := GetUser(oid)
				defer func() {
					_ = session.GetDb().Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(session.DBUser))

						ub, e := u.MarshalJSON()
						if e != nil {
							return e
						}

						e = b.Put([]byte(u.OpenID), ub)
						if e != nil {
							return e
						}

						fmt.Println(u.OpenID, "取消订阅：已经存入boltdb并且从map中删除")
						delete(Hub, u.OpenID)
						return nil
					})
				}()
				//	//用户已经关注公众号，则微信会将带场景值扫描事件推送给开发者
				//case message.EventScan:
				//	//do something
				//
				//	// 上报地理位置事件
				//case message.EventLocation:
				//	//do something
				//
				//	// 点击菜单拉取消息时的事件推送
				//case message.EventClick:
				//	//do something
				//
				//	// 点击菜单跳转链接时的事件推送
				//case message.EventView:
				//	//do something
				//
				//	// 扫码推事件的事件推送
				//case message.EventScancodePush:
				//	//do something
				//
				//	// 扫码推事件且弹出“消息接收中”提示框的事件推送
				//case message.EventScancodeWaitmsg:
				//	//do something
				//
				//	// 弹出系统拍照发图的事件推送
				//case message.EventPicSysphoto:
				//	//do something
				//
				//	// 弹出拍照或者相册发图的事件推送
				//case message.EventPicPhotoOrAlbum:
				//	//do something
				//
				//	// 弹出微信相册发图器的事件推送
				//case message.EventPicWeixin:
				//	//do something
				//
				//	// 弹出地理位置选择器的事件推送
				//case message.EventLocationSelect:
				//	//do something
				//
			}
		}
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: "unknown message"}
	})

	//处理消息接收以及回复
	err := s.Serve()

	if err != nil {

		fmt.Println(err)

		return
	}

	//createButton(wc)

	//发送回复的消息
	s.Send()

}

func UrlButton(msg, title string) string {
	return fmt.Sprintf("<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=%s'>%s</a", msg, title)
}
