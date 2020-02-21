package wc

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/silenceper/wechat/user"
	"main/session"
	"time"
)

const (
	NoneMode      = "none"
	LoveMode      = "love"
	PreLoveMode   = "pre_love"
	ResourcesMode = "resources"
)

var UserNotExistsError error
var UserMarshalFail error

func init() {
	UserNotExistsError = errors.New("user not exists")
	UserMarshalFail = errors.New("user marshal fail")
}

type Info struct {
	NickName       string  `json:"nick_name" bson:"nick_name"`
	Sex            int32   `json:"sex" bson:"sex"`
	City           string  `json:"city" bson:"city"`         // 普洱
	Country        string  `json:"country" bson:"country"`   // 中国
	Province       string  `json:"province" bson:"province"` // 云南
	Language       string  `json:"language" bson:"language"`
	Headimgurl     string  `json:"headimgurl" bson:"headimgurl"`
	SubscribeTime  int32   `json:"subscribe_time" bson:"subscribe_time"`
	UnionID        string  `json:"union_id" bson:"union_id"`
	GroupID        int32   `json:"group_id" bson:"group_id"`
	TagidList      []int32 `json:"tagid_list" bson:"tagid_list"`
	SubscribeScene string  `json:"subscribe_scene" bson:"subscribe_scene"`
}

type UserInfo struct {
	Mode            string                 `json:"mode"`
	OpenID          string                 `json:"open_id"`
	CTX             map[string]interface{} `json:"ctx" bson:"ctx"`
	Info            Info                   `json:"info" bson:"info"`
	LastMessageTime int64                  `json:"last_message_time"`
}

func (u *UserInfo) SetCTX(key string, value interface{}) error {
	if u == nil {
		return errors.New("empty pointer")
	}
	if u.CTX == nil {
		u.CTX = make(map[string]interface{}, 4)
	}
	u.CTX[key] = value
	return nil
}
func (u *UserInfo) GetCTX(key string) (interface{}, error) {
	if u == nil {
		return nil, errors.New("empty pointer")
	}

	if u.CTX == nil {
		u.CTX = make(map[string]interface{}, 8)
		return nil, errors.New("ctx empty")
	}
	return u.CTX[key], nil
}
func (u *UserInfo) SetOpenID(oid string) {
	u.OpenID = oid
}
func (u *UserInfo) SetMode(mode string) {
	u.Mode = mode
}
func (u *UserInfo) UpdateTime() {
	u.LastMessageTime = time.Now().Unix()
}
func GetUser(inf *user.Info, oid string) UserInfo {
	// 查看是否存在于Hub中

	var res UserInfo
	// 从bolt中获取
	err := session.GetDb().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(session.DBUser)
		if b != nil {
			data := b.Get([]byte(oid))

			// 如果用户数据不存在
			if data == nil {
				return UserNotExistsError
			}

			err := res.UnmarshalJSON(data)
			// 反序列化错误
			if err != nil {
				return UserMarshalFail
			}

			if res.OpenID == "" {
				return UserNotExistsError
			}

		} else {
			panic("bucket not exist")
		}
		return nil
	})

	if err != nil {
		res := UserInfo{
			Mode:            NoneMode,
			OpenID:          oid,
			CTX:             make(map[string]interface{}, 4), // 每个人预留五个空间应该足够了
			LastMessageTime: time.Now().Unix(),
		}

		rb, err := res.MarshalJSON()
		if err != nil {
			// 直接返回
			return res
		}
		_ = session.GetDb().Update(func(tx *bolt.Tx) error {

			b := tx.Bucket(session.DBUser)

			if b != nil {
				b.Put([]byte(oid), rb)
			}
			return nil
		})

		return res
	}

	return &res
}
