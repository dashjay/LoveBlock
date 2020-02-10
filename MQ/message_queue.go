package MQ

import (
	ai "github.com/night-codes/mgo-ai"
	"main/blocks"
	"main/database"
)

var ValidChan chan MessageQueue
var InvalidChan chan MessageQueue

type MessageQueue struct {
	Content string
	OpenID  string
}

func init() {
	ValidChan = make(chan MessageQueue, 64)
	InvalidChan = make(chan MessageQueue, 64)

	// 两个消息队列
	go Lover(&ValidChan)
	go InValidLover(&InvalidChan)
}

func Lover(c *chan MessageQueue) {

	s := database.NewSessionStore()

	for {

		select {

		case data := <-*c:
			{

				b := blocks.NewBlock(data.Content, blocks.GetLastBlock().Hash, data.OpenID)

				ai.Connect(s.C(database.DBCounters))

				mb := b.ConvertToBlockInMongo()

				mb.ID = ai.Next(database.AIBlockID)

				err := s.C("blocks").Insert(&mb)

				if err == nil {
					blocks.SetLastBlock(b)
				} else {
					panic(err)
				}
			}
		}
	}
}

func InValidLover(c *chan MessageQueue) {
	s := database.NewSessionStore()
	for {

		select {

		case data := <-*c:
			{
				b := blocks.NewBlock(data.Content, blocks.GetLastBlock().Hash, data.OpenID)
				_ = s.C("invalid_blocks").Insert(b.ConvertToBlockInMongo())
			}
		}
	}
}
