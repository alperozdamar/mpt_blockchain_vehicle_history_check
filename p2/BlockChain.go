package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"sort"
	"strings"
)

//Chain:This is a map which maps a block height to a list of blocks. The value is a list so that it can handle the forks
//Length equals to the highest block height.
type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (blockChain *BlockChain) Initial() {
	blockChain.Length = 0
	blockChain.Chain = make(map[int32][]Block)
}

func NewBlockChain() BlockChain{
	return BlockChain{
		Chain: make(map[int32][]Block),
		Length: 0,
	}
}

// This function takes a height as the argument, returns the list of blocks stored in that
// height or None if the height doesn't exist.
func (blockChain *BlockChain) Get(height int32) []Block {
	return blockChain.Chain[height]
}

// This function takes a block as the argument, use its height to find the corresponding list
// in blockchain's Chain map. If the list has already contained that block's hash, ignore it because we don't
// store duplicate blocks; if not, insert the block into the list.
func (blockChain *BlockChain) Insert(block Block) {
	if block.Header.Height > blockChain.Length {
		blockChain.Length = block.Header.Height
	}
	heightOfBlocks := blockChain.Chain[block.Header.Height]
	if heightOfBlocks == nil {
		heightOfBlocks = []Block{}
	}
	for _, heightBlock := range heightOfBlocks {
		if heightBlock.Header.Hash == block.Header.Hash {
			return
		}
	}
	blockChain.Chain[block.Header.Height] = append(heightOfBlocks, block)
}

// This function iterates over all the blocks, generate blocks' JsonString by the function you
// implemented previously, and return the list of those JsonStrings.
func (blockChain *BlockChain) EncodeToJSON() (string, error) {
	jsonBytes, error := json.Marshal(blockChain)
	return string(jsonBytes), error
}

//This function is called upon a blockchain instance. It takes a blockchain JSON string as input,
// decodes the JSON string back to a list of block JSON strings, decodes each block JSON string
// back to a block instance, and inserts every block into the blockchain.
func (blockChain *BlockChain) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), blockChain)
}

func (blockChain *BlockChain) MarshalJSON() ([]byte, error) {
	blockArray := make([]Block, 0)
	for _, v := range blockChain.Chain {
		blockArray = append(blockArray, v...)
	}
	return json.Marshal(blockArray)
}

func (blockChain *BlockChain) UnmarshalJSON(data []byte) error {
	blockArray := make([]Block, 0)
	err := json.Unmarshal(data, &blockArray)
	if err != nil {
		return err
	}
	blockChain.Initial()
	for _, block := range blockArray {
		blockChain.Insert(block)
	}
	return nil
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash + "<=" + block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

func (blockChain *BlockChain) GetParentBlock(block Block) Block {
	parentBlock:= Block{}
	blocks := blockChain.Chain[block.Header.Height-1]
	lenBlock := len(blocks)
	if lenBlock != 0{
		for i:=0; i < lenBlock; i++{
			if blocks[i].Header.Hash == block.Header.ParentHash {
				parentBlock = blocks[i]
				return parentBlock
			}
		}
	}
	return parentBlock
}

func (blockChain *BlockChain) GetLatestBlocks() []Block {
	height := blockChain.Length
	blocks := blockChain.Chain[height]
	return blocks
}

func (blockChain *BlockChain) Canonical() string {
	rs := ""
	forksBlocks:= blockChain.GetLatestBlocks()
	for i, currentBlock := range forksBlocks {
		height := blockChain.Length
		rs += "\n"
		rs += fmt.Sprintf("Chain # %d:\n ", i+1)
		for  height > 0{
			rs += fmt.Sprintf("height=%d, timestamp=%d, hash=%s, parentHash=%s, size=%d , value=%s\n",
				currentBlock.Header.Height, currentBlock.Header.Timestamp, currentBlock.Header.Hash,
				currentBlock.Header.ParentHash, currentBlock.Header.Size, currentBlock.Value)

			//rs += fmt.Sprintf("height=%d, timestamp=%d, hash=%s, parentHash=%s, size=%d , value=%s\n",
			//	currentBlock.Header.Height, currentBlock.Header.Timestamp, currentBlock.Header.Hash,
			//	currentBlock.Header.ParentHash, currentBlock.Header.Size, currentBlock.Value.DB)

			currentBlock, _= blockChain.GetBlock(currentBlock.Header.Height-1, currentBlock.Header.ParentHash)
			height = height - 1
		}
	}
	rs += "\n"
	fmt.Println(rs)
	return rs
}

func (blockChain *BlockChain) GetCarInformation(plate string) string {
	rs := ""
	forksBlocks:= blockChain.GetLatestBlocks()
	for i, currentBlock := range forksBlocks {
		height := blockChain.Length
		rs += "\n"
		rs += fmt.Sprintf("Chain # %d:\n ", i+1)
		for  height > 0{
			for _, valueObject := range currentBlock.Value.DB {
				if strings.Contains(valueObject.String(), plate) {
					fmt.Println("TransactionObject:", valueObject.String())
					rs += fmt.Sprintf("Value=%s\n", valueObject.String());
				}else{
					fmt.Println("plate:",plate," does not exist in our BlockChain!")
				}
			}
			//rs += fmt.Sprintf("height=%d, timestamp=%d, hash=%s, parentHash=%s, size=%d , value=%s\n",
			//	currentBlock.Header.Height, currentBlock.Header.Timestamp, currentBlock.Header.Hash,
			//		currentBlock.Header.ParentHash, currentBlock.Header.Size, currentBlock.Value)
			//	}
			currentBlock, _= blockChain.GetBlock(currentBlock.Header.Height-1, currentBlock.Header.ParentHash)
			height = height - 1
		}
	}
	rs += "\n"
	fmt.Println(rs)
	return rs
}


func (blockChain *BlockChain) GetBlock(height int32, hash string) (Block, bool) {
	isAvali := false
	block:= Block{}
	blocks := blockChain.Chain[height]
	lenBlock := len(blocks)
	if lenBlock != 0{
		for i:=0; i < lenBlock; i++{
			if blocks[i].Header.Hash == hash {
				block = blocks[i]
				isAvali = true
				return block, isAvali
			}
		}
	}
	return block, isAvali
}

func ConvertIntToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}