package wc

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/golang/protobuf/proto"
	"main/protos"
)

func Valid(ctx *context.Context) {
	prebm := GetInvalidBlock()
	mb, _ := proto.Marshal(&prebm)
	ctx.ResponseWriter.Write(mb)
	return
}
func Main(ctx *context.Context) {
	res := Index(ctx)
	prebm := protos.BaseMessage{
		Type: "text",
		Data: []byte(res),
	}
	bm, err := proto.Marshal(&prebm)
	if err != nil {
		ctx.ResponseWriter.Write([]byte(err.Error()))
	} else {
		ctx.ResponseWriter.Write(bm)
	}
}

func Index(ctx *context.Context) string {

	function := ctx.Input.Query("func")
	switch function {

	case "create":
		{
			content := ctx.Input.Query("content")
			if content == "" {
				return "表白内容为空"
				//mb, _ := proto.Marshal(&prebm)
				//ctx.ResponseWriter.Write(mb)
				//return
			}
			if len(content) > 4*520 {
				return "当前缓冲区文字超过520个字"
				//mb, _ := proto.Marshal(&prebm)
				//ctx.ResponseWriter.Write(mb)
				//return
			}
			// 获取openid
			oid := ctx.Input.Query("oid")
			res := PostBlock(content, oid)

			//prebm.Data = []byte(res)
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			return res
		}
	case "view":
		{
			res := GetOneBlock()
			return res
		}
	case "pass":
		{

			hash := ctx.Input.Query("hash")
			res := PassBlock(hash)
			return res
			//prebm.Data = []byte(res)
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			//return
		}
	case "get_prev":
		{
			hash := ctx.Input.Query("hash")
			res := GetTargetBlock(hash)
			return res
			//prebm.Data = []byte(res)
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			//return
		}
	case "like":
		{
			//获取必要参数id 是表白的编号，oid是用户的openid
			id := ctx.Input.Query("id")
			oid := ctx.Input.Query("oid")
			// 执行点赞
			res := AddLike(id, oid)
			return res
			//prebm.Data = []byte(res)
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			//return
		}
	case "reply":
		{
			return fmt.Sprintf("回复功能当前正在开发中%s", menu)
			//prebm.Data = []byte()
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			//return
		}
	case "search":
		{

			keyword := ctx.Input.Query("keyword")
			res := SearchBlock(keyword)
			return res
			//prebm.Data = []byte(res)
			//mb, _ := proto.Marshal(&prebm)
			//ctx.ResponseWriter.Write(mb)
			//return
		}
	case "random":
		{
			num := ctx.Input.Query("num")
			res := RandomBlock(num)
			return res
		}

	}
	return ""
}
