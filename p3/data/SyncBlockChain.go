package data

import (
	"../../p2"
	"fmt"
	"../../p1"
	"sync"
)

type SyncBlockChain struct {
	bc p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	isAvail := false
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	if sbc.bc.Get(height) != nil {
		isAvail = true
	}
	return sbc.bc.Get(height),isAvail
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetBlock(height,hash);
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

// This function would check if the block with the given "parentHash" exists in the blockChain. If we have the parent
// block, we can insert the next block; if we don't have the parent block, we have to download the parent block before
// inserting the next block.
//CheckParentHash() is used to check if the parent hash(or parent block) exist in the current blockchain when you want
// to insert a new block sent by others. For example, if you received a block of height 7, you should check if its
// parent block(of height 6) exist in your blockchain. If not, you should ask others to download that parent block of
// height 6 before inserting the block of height 7.
func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	heightBlocks := sbc.bc.Get((insertBlock.Header.Height) - 1)
	for _, heightBlock := range heightBlocks { // find simmilar hash in blockchain
		if heightBlock.Header.Hash == insertBlock.Header.ParentHash {
			return true
		}
	}
	return false
}

//json to blockchain, use when download blockchain for second node
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blockChain := sbc.bc.DecodeFromJSON(blockChainJson)
	fmt.Println(blockChain)
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJSON()
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie,nonce string) p2.Block {
	var parentHash string
	var blockList []p2.Block
	var found bool
	currHeight := sbc.bc.Length


	fmt.Println("alper.test.currHeight:",currHeight)

	for currHeight >= 1 {
		blockList, found = sbc.Get(currHeight)
		if found == true {
			parentHash = blockList[0].Header.Hash
			//fmt.Println("ParentHash:",parentHash)
			break
		}
		currHeight--
	}
	if currHeight == 0 {
		parentHash = "firstParentHash"
	}

	var newBlock p2.Block
	newBlock.Initial(currHeight+1, parentHash, mpt,nonce)

	//fmt.Println("TEST.newBlock.nonce:",nonce)
	//fmt.Println("TEST.newBlock.newBlock.parentHash:",newBlock.Header.ParentHash)
	return newBlock
}

func (sbc *SyncBlockChain) Show() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Show()
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) p2.Block  {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}


func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

func (sbc *SyncBlockChain) Canonical() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Canonical()
}

func (sbc *SyncBlockChain) GetCarInformation(plate string) string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetCarInformation(plate)
}