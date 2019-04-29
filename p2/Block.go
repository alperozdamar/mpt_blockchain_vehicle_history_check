package p2

import (
	"../p1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"time"
)

type Block struct {
	Header BlockHeader        //`json:"header"`
	Value  p1.MerklePatriciaTrie //`json:"value"`
}

/*type BlockHeader struct {
	Hash       string
	Timestamp  int64
	Height     int32
	ParentHash string
	Size       int32
}*/

type BlockHeader struct {
	Size       int32
	ParentHash string
	Height     int32
	Timestamp  int64
	Hash       string
	Nonce 	   string
}

type JsonBlock struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Nonce 	   string			 `json:"nonce"`
	Mpt        map[string]string `json:"mpt"`
}

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

//Calculate Block's Size and Hash.
func (block *Block) calculateSizeAndHash() {
	block.Header.Size = int32(len([]byte(fmt.Sprintf("%v", block.Value))))
	hashSha3 := sha3.New256()
	hashStr := string(block.Header.Height) +
		string(block.Header.Timestamp) +
		block.Header.ParentHash +
		block.Value.Root +
		string(block.Header.Size)
	block.Header.Hash = hex.EncodeToString(hashSha3.Sum([]byte(hashStr)))
}

func (block *Block) Initial(height int32, parentHash string, value p1.MerklePatriciaTrie, nonce string ) {
	block.Header.Height = height
	block.Header.Timestamp = time.Now().Unix()
	block.Header.ParentHash = parentHash
	block.Header.Size = int32(len([]byte(fmt.Sprintf("%v", value))))
	block.Value = value
	block.Header.Nonce = nonce;
	block.Header.Hash = block.Hash()
}


func (block *Block) EncodeToJson() (string, error) {
	jsonBytes, error := json.Marshal(block)
	return string(jsonBytes), error
}

func (block *Block) DecodeFromJson(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), block)
}

func (block *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(JsonBlock{
		Height:     block.Header.Height,
		Timestamp:  block.Header.Timestamp,
		Hash:       block.Header.Hash,
		ParentHash: block.Header.ParentHash,
		Size:       block.Header.Size,
		Mpt:        block.Value.ToMptMap(),
		Nonce:		block.Header.Nonce,
	})
}

func (block *Block) UnmarshalJSON(data []byte) error {
	jsonBlock := JsonBlock{}
	err := json.Unmarshal(data, &jsonBlock)
	if err != nil {
		return err
	}
	block.Header.Height = jsonBlock.Height
	block.Header.Timestamp = jsonBlock.Timestamp
	block.Header.Hash = jsonBlock.Hash
	block.Header.ParentHash = jsonBlock.ParentHash
	block.Header.Size = jsonBlock.Size
	block.Header.Nonce=jsonBlock.Nonce

	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	for k, v := range jsonBlock.Mpt {
		mpt.Insert(k, v)
	}
	block.Value = mpt
	return nil
}


func (block *Block) GetHash() string {
	return block.Header.Hash
}

func (block *Block) Hash() string {
	var hashStr string
	hashStr = string(block.Header.Height) + string(block.Header.Timestamp) + string(block.Header.ParentHash) +
		string(block.Value.Root) + string(block.Header.Size) + string(block.Header.Nonce)
	sum := sha3.Sum256([]byte(hashStr))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}