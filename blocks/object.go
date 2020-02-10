package blocks

type Object struct {
	BID       uint64 `json:"bid" bson:"bid"` // block id
	ID        uint64 `json:"id" bson:"id"`   //id
	Type      string `json:"type" bson:"type"`
	OpenID    string `json:"open_id" bson:"open_id"`       // 微信openid
	TimeStamp int64  `json:"time_stamp" bson:"time_stamp"` // 时间
}
