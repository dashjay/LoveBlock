package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"main/MQ"
	"main/blocks"
	"main/database"
	"main/wc"
)

func init() {

	// 初始化数据库
	err := database.InitMongoDB()
	if err != nil {
		panic(err)
	}

	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	// 开始状态

	var b blocks.BlockInMongo

	err = con.Find(nil).Sort("-timestamp").One(&b)

	if err != nil {
		b := blocks.NewGenesisBlock()

		con.Insert(b.ConvertToBlockInMongo())

		blocks.SetLastBlock(b)
	}
	blocks.SetLastBlock(b.ConvertToBlock())
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
	go MQ.Lover(&MQ.ValidChan, s1)
	go MQ.InValidLover(&MQ.InvalidChan, s2)

	beego.Any("/", wc.Index)
	beego.Any("/get", wc.Get)
	beego.Any("/loadmore", wc.LoadMore)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	beego.Run(":8001")
}
