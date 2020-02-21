package blocks

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"

	"sync"
	"time"
)

type lastBlock struct {
	b  *Block
	mu sync.RWMutex
}

var lb *lastBlock

func GetLastBlock() *Block {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.b
}
func SetLastBlock(b *Block) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.b = b
}

func init() {
	lb = &lastBlock{
		b:  &Block{},
		mu: sync.RWMutex{},
	}
}

// Block keeps block headers
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	OpenID        string
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte, OpenID string) *Block {
	block := &Block{Timestamp: time.Now().Unix(), Data: []byte(data), PrevBlockHash: prevBlockHash, Hash: []byte{}, OpenID: OpenID}
	block.SetHash()
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock() *Block {
	return NewBlock("我ZWJ永远爱ZFQ", []byte("I Love You"), "")
}

type BlockInMongo struct {
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
	Data          string `json:"data" bson:"data"`
	PrevBlockHash string `json:"prev_block_hash" bson:"prev_block_hash"`
	Hash          string `json:"hash" bson:"hash"`
	OpenID        string `json:"open_id" bson:"open_id"`
	ID            uint64 `json:"id" bson:"id"`
}

type BlockFront struct {
	BlockInMongo
	LikeNum     int
	ReplyNum    int
	ReplyTarget string
}

func (b BlockInMongo) ConvertToBFront() BlockFront {
	return BlockFront{
		BlockInMongo: b,
		LikeNum:      0,
		ReplyNum:     0,
	}
}

func (b *BlockInMongo) Formatter() string {

	return fmt.Sprintf("#%d 内容：", b.ID) + b.Data + fmt.Sprintf("\n<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=reply %d'>回复该表白❤</a>\t<a href='weixin://bizmsgmenu?msgmenuid=1&msgmenucontent=like %d'>点个赞️🌟</a>\n\n", b.ID, b.ID)
}

func (b *BlockInMongo) ConvertToBlock() *Block {
	return &Block{
		Timestamp:     b.Timestamp,
		Data:          []byte(b.Data),
		PrevBlockHash: Base58Decode([]byte(b.PrevBlockHash)),
		Hash:          Base58Decode([]byte(b.Hash)),
		OpenID:        b.OpenID,
	}
}

// ConvertToBlockInMongo 从Block转化成能存入数据库的的MongoBlock
func (b *Block) ConvertToBlockInMongo() BlockInMongo {

	return BlockInMongo{
		Timestamp:     b.Timestamp,
		Data:          string(b.Data),
		PrevBlockHash: string(Base58Encode(b.PrevBlockHash)),
		Hash:          string(Base58Encode(b.Hash)),
		OpenID:        b.OpenID,
	}
}
