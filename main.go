package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	ai "github.com/night-codes/mgo-ai"
	"main/blocks"
	"main/database"
	"main/wc"
)

func init() {
	// 初始化数据库
	ds := database.NewSessionStore()
	defer ds.Close()
	con := ds.C("blocks")

	// 开始状态

	var b blocks.BlockInMongo

	err := con.Find(nil).Sort("-timestamp").One(&b)

	ai.Connect(ds.C(database.DBCounters))

	if err != nil {
		b := blocks.NewGenesisBlock()
		mb := b.ConvertToBlockInMongo()
		mb.ID = ai.Next(database.AIBlockID)
		con.Insert(mb)
		blocks.SetLastBlock(b)
	} else {
		blocks.SetLastBlock(b.ConvertToBlock())
	}
}

func main() {

	beego.Any("/", wc.Main)
	beego.Any("/valid", wc.Valid)
	beego.Any("/get", wc.Get)
	beego.Any("/random", wc.Random)
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
