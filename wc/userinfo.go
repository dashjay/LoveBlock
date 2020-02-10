package wc

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"main/session"
	"os"
	"strings"
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
	go SwapUserInfo()
	UserNotExistsError = errors.New("user not exists")
	UserMarshalFail = errors.New("user marshal fail")
}

type UserInfo struct {
	Mode            string                 `json:"mode"`
	OpenID          string                 `json:"open_id"`
	CTX             map[string]interface{} `json:"ctx" bson:"ctx"`
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
				Mode:            NoneMode,
				OpenID:          oid,
				CTX:             make(map[string]interface{}, 4), // 每个人预留五个空间应该足够了
				LastMessageTime: time.Now().Unix(),
			}

			return Hub[oid]
		}
		Hub[oid] = &res
		return &res
	}

	return Hub[oid]
}

// 实时将内存中的用户信息放入boltdb中
func SwapUserInfo() {

	var duration = time.Minute * 30
	// ticker 设定为1分钟
	t := time.Tick(duration)
	for {
		select {
		case <-t: // 每分钟执行一次
			{
				var all = len(Hub) // 所有数量
				var i = 0
				var e = 0
				_ = session.GetDb().Update(func(tx *bolt.Tx) error {
					b := tx.Bucket(session.DBUser)
					for k, v := range Hub {
						pre, err := v.MarshalJSON()
						if err != nil {
							// 这里要记录日志
							e++
							continue
						}
						// 24小时未发言
						if time.Unix(v.LastMessageTime, 0).Add(time.Hour * 24).Before(time.Now()) {

						}
						_ = b.Put([]byte(k), pre)
						i++
					}
					return nil
				})

				var buf strings.Builder
				buf.WriteString(fmt.Sprintf("-------%s--------\n", time.Now().String()))
				buf.WriteString(fmt.Sprintf("一共更新了%d条,正常%d条，不正常%d条\n\n", all, i, e))
				f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					fmt.Println(buf.String())
				}
				f.WriteString(buf.String())
				f.Close()

				t = time.Tick(duration)
			}
		}
	}
}
