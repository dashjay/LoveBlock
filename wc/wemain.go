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

	fmt.Println(s.Context)
	s.SetMessageHandler(func(v message.MixMessage) *message.Reply {

		u := GetUser(s.GetOpenID())

		u.UpdateTime()

		if len(u.OpenID) != 28 {
			return newTextMessage("openid error: " + s.GetOpenID())
		}

		switch v.MsgType {
		// if文本消息
		case message.MsgTypeText:

			// 获取内容长度
			var length = len(v.Content)

			// OpenidError

			switch {

			case length == 1:
				{
					switch v.Content {

					case "1":
						{
							u.SetMode(LoveMode)
							return newTextMessage(rHelpLove)
						}
					case "2":
						{
							u.SetMode(ResourcesMode)
							return newTextMessage("成功切换为资源模式（开发中......）")
						}
					case "3":
						{
							u.SetMode(NoneMode)
							return newTextMessage(rOnSubscribe)
						}
					}
				}

			case length == 2:
				{

				}
			case length == 3:
				{
					if v.Content == "~!@#" {
						return GetInvalidBlock()
					}
				}

			case length == 4:
				{
					if v.Content == "帮助" {

					}
				}
			case length >= 4:
				{

					fmt.Println(v.Content)
					if v.Content == "表白帮助" {

						return newTextMessage(rHelpLoveMode)
					}

					if v.Content == "资源帮助" {
						return newTextMessage("资源模式帮助")
					}

					switch u.Mode {

					case PreLoveMode:
						{

							if v.Content == "确定" {

								c, err := u.GetCTX(PreLoveMode)
								if err != nil {
									return newTextMessage("出现异常请重试: " + err.Error())
								}
								if c.(string) == "" {
									return newTextMessage("当前缓冲区为空，请回复内容尝试表白哟")
								}
								if len(c.(string)) > 3*520 {
									return newTextMessage("当前缓冲区文字超过520个字")
								}
								u.SetMode(LoveMode)
								return PostBlock(u.CTX[PreLoveMode].(string), u.OpenID)
							}

							if v.Content == "取消" {
								u.CTX[PreLoveMode] = ""
								u.SetMode(LoveMode)
								return newTextMessage("「已取消」爱就要大声说出来~！")
							}

							_ = u.SetCTX(PreLoveMode, v.Content)

							return newTextMessage(
								"确定表白的内容为：\n\n" + v.Content + "\n\n吗？\n " +
									UrlButton("取消", "取消表白") + "\t " +
									UrlButton("确定", "确定发送"))
						}
					case LoveMode:
						{

							if v.Content == "表白" {
								u.SetMode(PreLoveMode)
								return newTextMessage("请直接回复积极正向的表白(520字内), 寻人或求偶信息 或<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=取消'>点我取消表白</a> 或回复[取消]以取消")
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

							if strings.Contains(v.Content[0:4], "like") {
								return AddLike(v, u.OpenID)
							}

							if strings.Contains(v.Content[0:5], "reply") {

							}

							if strings.Contains(v.Content[0:6], "search") {
								return SearchBlock(v)
							}

							if strings.Contains(v.Content[0:6], "random") {
								return RandomBlock(v)
							}

						}
					case ResourcesMode:
						{

						}
					}

				}
				return newTextMessage(rOnSubscribe)

				//	//图片消息
				//case message.MsgTypeImage:
				//	//do something
				//
				//	//语音消息
				//case message.MsgTypeVoice:
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

				defer func() {
					_ = session.GetDb().Update(func(tx *bolt.Tx) error {
						b := tx.Bucket(session.DBUser)

						ub, e := u.MarshalJSON()
						if e != nil {
							return e
						}

						e = b.Put([]byte(u.OpenID), ub)
						if e != nil {
							return e
						}

						//fmt.Println(u.OpenID, "取消订阅：已经存入boltdb并且从map中删除")
						delete(Hub, u.OpenID)
						return nil
					})
				}()
			}

		}

		return newTextMessage(rOnSubscribe)
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
	return fmt.Sprintf("<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=%s'>%s</a>", msg, title)
}
