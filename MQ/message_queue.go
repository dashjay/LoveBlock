package MQ

import (
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
}

func Lover(c *chan MessageQueue, s *database.SessionStore) {
	for {

		select {

		case data := <-*c:
			{

				b := blocks.NewBlock(data.Content, blocks.GetLastBlock().Hash, data.OpenID)

				err := s.C("blocks").Insert(b.NewBIMFromBlock())

				if err == nil {
					blocks.SetLastBlock(b)
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
				b := blocks.NewBlock(data.Content, blocks.GetLastBlock().Hash, data.OpenID)
				_ = s.C("invalid_blocks").Insert(b.NewBIMFromBlock())
			}
		}
	}
}
