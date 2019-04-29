package data

import (
	"encoding/json"
	"../../p1"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	return HeartBeatData{
		IfNewBlock:  ifNewBlock,
		Id:          id,
		BlockJson:   blockJson,
		PeerMapJson: peerMapJson,
		Addr:        addr,
		Hops:        2,
	}
}

//PrepareHeartBeatData() is used when you want to send a HeartBeat to other peers.
// PrepareHeartBeatData would first create a new instance of HeartBeatData, then decide whether or not you
// will create a new block and send the new block to other peers.
//In the arguments of "heartbeat.PrepareHeartBeatData()" function, there's a variable "peerMapBase64 string".
// It meant to be "peerMapJSON string" which is the JSON format of peerList.peerMap. It is the output of function
// "peerList.PeerMapToJSON()". Sorry for the confusion. There is no BASE64 in this project.
func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string, generateNewBlock bool,nonce string , mpt p1.MerklePatriciaTrie) HeartBeatData {
	newHeartBeatData := NewHeartBeatData(false, selfId, "", peerMapJson, addr)

	if generateNewBlock {
		newBlock := sbc.GenBlock(mpt, nonce)
		blockJson, _ := newBlock.EncodeToJson()
		newHeartBeatData = NewHeartBeatData(true, selfId, blockJson, peerMapJson, addr)
	}else{
		newHeartBeatData = NewHeartBeatData(false, selfId, "", peerMapJson, addr)
	}
	return newHeartBeatData
}

func (data *HeartBeatData) EncodeToJson() (string, error) {
	jsonBytes, error := json.Marshal(data)
	return string(jsonBytes), error
}

func (data *HeartBeatData) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), data)
}

