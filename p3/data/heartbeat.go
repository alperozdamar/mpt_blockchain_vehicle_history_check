package data

import (
	"../../p1"
	"crypto/rsa"
	"encoding/json"
)

type HeartBeatData struct {
	IfNewBlock  		bool   			`json:"ifNewBlock"`
	Id          		int32  			`json:"id"`
	BlockJson   		string 			`json:"blockJson"`
	PeerMapJson 		string 			`json:"peerMapJson"`
	Addr        		string 			`json:"addr"`
	Hops        		int32			`json:"hops"`
	PeerPublicKey 		*rsa.PublicKey `json:"peerPublicKey"`
	IfValidTransaction 	bool			`json:"ifValidTransaction"`
	TransactionInfoJson string			`json:"transactionInfoJson"`
	Balance				int				`json:"balance"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string, peerPublicKey *rsa.PublicKey, ifValidTransaction bool, transactionInfoJson string, balance int) HeartBeatData {
	return HeartBeatData{
		IfNewBlock:  ifNewBlock,
		Id:          id,
		BlockJson:   blockJson,
		PeerMapJson: peerMapJson,
		Addr:        addr,
		Hops:        2,
		PeerPublicKey: peerPublicKey,
		IfValidTransaction: ifValidTransaction,
		TransactionInfoJson: transactionInfoJson,
		Balance: balance,
	}
}

//PrepareHeartBeatData() is used when you want to send a HeartBeat to other peers.
// PrepareHeartBeatData would first create a new instance of HeartBeatData, then decide whether or not you
// will create a new block and send the new block to other peers.
//In the arguments of "heartbeat.PrepareHeartBeatData()" function, there's a variable "peerMapBase64 string".
// It meant to be "peerMapJSON string" which is the JSON format of peerList.peerMap. It is the output of function
// "peerList.PeerMapToJSON()". Sorry for the confusion. There is no BASE64 in this project.
func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string, generateNewBlock bool,nonce string ,mpt p1.MerklePatriciaTrie,peerPublicKey *rsa.PublicKey,ifValidTransaction bool,transactionInfoJson string,balance int) HeartBeatData {
	newHeartBeatData := NewHeartBeatData(false, selfId, "", peerMapJson, addr,peerPublicKey,ifValidTransaction ,transactionInfoJson ,balance)
	if generateNewBlock && ifValidTransaction {
		newBlock := sbc.GenBlock(mpt, nonce)
		blockJson, _ := newBlock.EncodeToJson()
		newHeartBeatData = NewHeartBeatData(true, selfId, blockJson, peerMapJson, addr,peerPublicKey,ifValidTransaction ,transactionInfoJson ,balance)
	}else{
		newHeartBeatData = NewHeartBeatData(false, selfId, "", peerMapJson, addr,peerPublicKey,ifValidTransaction ,transactionInfoJson ,balance)
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

