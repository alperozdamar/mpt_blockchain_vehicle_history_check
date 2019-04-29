package main

import (
	"./p3"
	"log"
	"net/http"
	"os"
)

func main() {

//	testJson :="{\"hash\":\"\",\"timeStamp\":0,\"height\":0,\"parentHash\":\"\",\"size\":0,\"mpt\":{}}";

//	newBlock:=p2.Block{}

//	newBlock.DecodeFromJson(testJson)


//	fmt.Println("Alper Block :",newBlock)
//	fmt.Println("Alper P2.Block :",p2.Block{})

//	mpt1 := p1.MerklePatriciaTrie{}
//	mpt1.Initial()
//	block1 := p2.Block{}
//	block1.Initial(1,"ab",mpt1);

	//data.TestPeerListRebalance()
	router := p3.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":" + os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}
