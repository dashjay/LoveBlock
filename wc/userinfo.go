package wc

import (
	"errors"
	"github.com/boltdb/bolt"
	"main/session"
	"time"
)

const (
	NoneMode      = "none"
	LoveMode      = "love"
	PreLoveMode   = "pre_love"
	ResourcesMode = "resources"
)

var Hub map[string]*UserInfo = nil

var UserNotExistsError error
var UserMarshalFail error

func init() {
	Hub = make(map[string]*UserInfo, 256)
	UserNotExistsError = errors.New("user not exists")
	UserMarshalFail = errors.New("user marshal fail")
}

type UserInfo struct {
	Mode            string                 `json:"mode"`
	OpenID          string                 `json:"open_id"`
	CTX             map[string]interface{} `json:"ctx" bson:"ctx"`
	LastMessageTime int64                  `json:"last_message_time"`
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

func GetUser(oid string) *UserInfo {
	// 查看是否存在于Hub中
	if _, j := Hub[oid]; !j {
		//fmt.Println(oid, "不存在")
		// 当前hub中没有该用户数据
		// 初始化boltdb
		db := session.GetDb()
		var res UserInfo
		// 从bolt中获取
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("user"))
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

		// 用户不存在或者反序列化
		if err != nil {
			// 设定新的值
			Hub[oid] = &UserInfo{
				Mode:            LoveMode,
				OpenID:          oid,
				CTX:             make(map[string]interface{}, 5), // 每个人预留五个空间应该足够了
				LastMessageTime: time.Now().Unix(),
			}
			return Hub[oid]
		}
		Hub[oid] = &res
		return &res
	}

	return Hub[oid]
}
