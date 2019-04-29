package p3

import (
	"../p1"
	"../p2"
	"./data"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var ASK_PEER_REQUEST = "/block"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "localhost:6686";				//peer's address! It will updated after initialization.

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

var FIRST_PEER_ADDRESS="localhost:6686"			//first peer's hard-coded address!
var SELF_ID=0
var PROBLEM_IN_TA_SERVER=1
var HeartBeatVariable data.HeartBeatData
var StopGeneratingNewBlock =false

var SPECIAL_BLOCK_PREFIX="00000"; //5 zeros...

//Create SyncBlockChain and PeerList instances.
func init() {
		fmt.Println("Init method is triggered!")
		SBC = data.NewBlockChain()
		Peers = data.NewPeerList(Peers.GetSelfId(),32);
		mpt:=p1.MerklePatriciaTrie{}
		mpt.Initial()
		mpt.Insert(p2.String(2),p2.String(5))
		block:=p2.Block{}
		block.Initial(1,"gensis",mpt,SPECIAL_BLOCK_PREFIX)
		block.Header.Nonce = SPECIAL_BLOCK_PREFIX
		SBC.Insert(block)

		if len(os.Args) > 1 {
			responseString := string(os.Args[1])
			fmt.Println(responseString)
			result , err := strconv.ParseInt(responseString,10,32)
			if err != nil {
				panic(err)
			}
			id  := int32(result)
			fmt.Printf("Parsed int is %d\n", result)
			Peers.Register(id);
			SELF_ADDR="localhost"+os.Args[1];
			//Add First Node's IP to here!!
			Peers.Add(FIRST_PEER_ADDRESS,6686)
		} else {
			Peers.Register(6686)
			SELF_ADDR="localhost:6686";
		}
	}


// Register ID, download BlockChain, start HeartBeat
//Start() function would start the logic such as register on server, and start heartBeats.
//TA: Get an ID from TA's server, download the BlockChain from your own first node,
// use "go StartHeartBeat()" to start HeartBeat loop.
func Start(w http.ResponseWriter, r *http.Request) {
	//Register
	if(PROBLEM_IN_TA_SERVER==0) {
		Register()
	}else{
		fmt.Println("Problem in TA Server. SelfId set manually!")
		fmt.Println("Manuel Self-Id:",Peers.GetSelfId())
	}
	fmt.Println("My Host:",r.Host)
	SELF_ADDR=r.Host
	//Download BlockChain
	//r.Host: localhost:6686
	if r.Host != FIRST_PEER_ADDRESS{
		fmt.Println("I am not the first node! I need to download!")
		Download()
	}else{
		fmt.Println("I am the first node! No need to download!")
	}

	//go StartTryingNonce()
	StartTryingNonce()

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <- ticker.C:
			StartHeartBeat()
		case <- quit:
			ticker.Stop()
			return
		}
	}
}

// Display peerList and sbc
// T.A. :  Shows the PeerMap and the BlockChain.
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", SBC.Canonical())
}

// Register to TA's server, get an ID and register to PeerList
func Register() {
	response, err := http.Get(REGISTER_SERVER)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	responseString := string(responseData)
	fmt.Println(responseString)
	result , err := strconv.ParseInt(responseString,10,32)
	if err != nil {
		panic(err)
	}
	id  := int32(result)
	fmt.Printf("Parsed int is %d\n", result)
	Peers = data.NewPeerList(id,32);
}


// Download blockchain from TA server
//Download URL: the address would be where you launch your own node(For example, http://localhost:6688);
// the API would be "/upload" since it is uploaded by the uploader. Here are the steps:
// (1) You launch the first node at http://localhost:6688.
// (2) You launch the second node at another address.
// (3) The second node would download the current blockChain from http://localhost:6688/upload.
//How should we find the PeerList and download the BlockChain? Can you please add more explanations to this?
//You can send a HeartBeatData to the first node, and add a flag says you want the PeerList and the BlockChain.

//T.A. : Download the current BlockChain from your own first node(can be hardcoded).
// It's ok to use this function only after launching a new node. You may not need it after node starts heartBeats.
func Download() {
	uploadAddress:="http://"+FIRST_PEER_ADDRESS+"/upload";
	fmt.Println("Upload Post Request will be sent to :" + uploadAddress)
	peerMapStringValue,_ :=Peers.EncodePeerMapToJSON()
	registerData :=data.NewRegisterData(Peers.GetSelfId(),peerMapStringValue);
	jsonBytes, err := json.Marshal(registerData)
	req, err := http.NewRequest("POST", uploadAddress, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)
	fmt.Println(">>>>>>>>>>>>>> Upload Request Sent  To:[", FIRST_PEER_ADDRESS ,"] >>>>>>>>>>>>>>>>")
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	//body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println("response Body:", string(body))
	fmt.Println("Got the response from peer:",response.Body)
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	responseString := string(responseData)
	SBC.UpdateEntireBlockChain(responseString)
	fmt.Println(SBC)
}

//T.A. : Return the BlockChain's JSON. And add the remote peer into the PeerMap.
func Upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("<<<<<<<<<<<<<< Upload Received  From:[", r.Host ,"] <<<<<<<<<<<<")
	fmt.Println("Upload Method is triggered!")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	registerData := new(data.RegisterData)
	registerData.DecodeFromJSON(string(body))
	fmt.Println("registerData.AssignedId:",registerData.AssignedId)
	fmt.Println("registerData.PeerMapJson:",registerData.PeerMapJson)
	blockChainJSON, err := SBC.BlockChainToJson()
	fmt.Println("SBC:",SBC);
	//fmt.Println("SBC.Show:",SBC.Show());
	fmt.Println("Data.NewBlockchain:",data.NewBlockChain())
	if err != nil {
		//data.PrintError(err, "Upload")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	fmt.Println("blockChain:"+blockChainJSON)
	//UploadBlock(w,r);
	fmt.Fprint(w, blockChainJSON)
}


// Upload a block to whoever called this method, return jsonStr
//T.A. : Return the Block's JSON.
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	fmt.Println("<<<<<<<<<<<<<< UploadBlock Received  From:[", r.Host ,"] <<<<<<<<<<<<")
	param := strings.Split(r.URL.Path,"/")
	h, err := strconv.ParseInt(param[2], 10, 32)
	fmt.Println("param0:",param[0])
	fmt.Println("param1:",param[1])
	fmt.Println("param2:",param[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: InternalServerError. " + err.Error()))
	} else {
		encode := param[3]
		block, flag := SBC.GetBlock(int32(h), encode)
		if flag == false {
			w.WriteHeader(http.StatusNoContent)
		} else {
			blockStr, err := block.EncodeToJson()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("HTTP 500: InternalServerError. " + err.Error()))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(blockStr))
			}
		}
	}
}

// Received a heartbeat
// T.A. : Add the remote address, and the PeerMapJSON into local PeerMap. Then check if the HeartBeatData contains a new block.
// 			If so, do these:
// 			(1) check if the parent block exists. If not, call AskForBlock() to download the parent block.
// 			(2) insert the new block from HeartBeatData.
// 			(3) HeartBeatData.hops minus one, and if it's still bigger than 0, call ForwardHeartBeat()
// 				to forward this heartBeat to all peers.
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	fmt.Println("<<<<<<<<<<<<<< HeartBeat Received  From:[", r.Host ,"] <<<<<<<<<<<<")
	fmt.Fprintf(w, "%s\n", r.Host)
	fmt.Fprintf(w, "%s", string(data))
	error := json.Unmarshal(data, &HeartBeatVariable)
	if (error != nil) {
		fmt.Println("Error occured in HeartBeatReceive: ", error)
	} else {

	}
	if (HeartBeatVariable.Addr ==  SELF_ADDR) {
		return
	}
	Peers.Add(HeartBeatVariable.Addr, HeartBeatVariable.Id)
	Peers.InjectPeerMapJson(HeartBeatVariable.PeerMapJson,  SELF_ADDR)
	if (HeartBeatVariable.IfNewBlock) {
		heartBlock := p2.Block{}
		heartBlock.DecodeFromJson(HeartBeatVariable.BlockJson)

		// Proof of Work...
		receivedPuzzle := heartBlock.Header.ParentHash+ heartBlock.Header.Nonce + heartBlock.Value.Root
		sum := sha3.Sum256([]byte(receivedPuzzle))

		if (strings.HasPrefix(hex.EncodeToString(sum[:]), SPECIAL_BLOCK_PREFIX)){
			fmt.Println("Block with SPECIAL PREFIX arrived from:[", r.Host ,"]")
			latestBlocks := SBC.GetLatestBlocks()
			for i:= 0 ; i < len(latestBlocks); i++ {
				if latestBlocks[i].Header.Hash == heartBlock.Header.ParentHash {
					StopGeneratingNewBlock =true
					break
				}
			}
			if (heartBlock.Header.Height == 1) {
				SBC.Insert(heartBlock)
			} else {
				_, flag := SBC.GetBlock(heartBlock.Header.Height-1, heartBlock.Header.ParentHash)
				if flag {
					SBC.Insert(heartBlock)
				} else {
					AskForBlock(heartBlock.Header.Height-1, heartBlock.Header.ParentHash)
					SBC.Insert(heartBlock)
				}
			}
		}else{
			fmt.Println(" Ignoring incoming Heart Beat Block! Unmatched Puzzle! Calculated Puzzle:", hex.EncodeToString(sum[:]))
			fmt.Println("TEST.ALPER.RECEIVE********************************************************")
			fmt.Println("Incoming Heart Beat Block.Hash::",heartBlock.Header.Hash)
			fmt.Println("Incoming Heart Beat Nonce:",heartBlock.Header.Nonce)
			fmt.Println("Incoming Heart Beat mpt.Root:",heartBlock.Value.Root)
			fmt.Println("Calculated Incoming Hash Puzzle:",hex.EncodeToString(sum[:]))
		}
	} else {
		fmt.Println("HeartBeat flag is false! There is no block in heartBeat!")
	}
	HeartBeatVariable.Hops -=  1
	if HeartBeatVariable.Hops > 0 {
		HeartBeatVariable.Addr = SELF_ADDR
		HeartBeatVariable.Id = Peers.GetSelfId()
		ForwardHeartBeat(HeartBeatVariable)
	}
}


// Ask another server to return a block of certain height and hash .
// AskForBlock will be called in HeartBeatReceive, in AskForBlock you will call http
// get to /localhost:port/block/{height}/{hash} (UploadBlock) to get the Block
//
// Ask another server to return a block of certain height and hash
// T.A. : Loop through all peers in local PeerMap to download a block. As soon as one peer returns the block, stop the loop.
func AskForBlock(height int32, hash string) {
	fmt.Println("Ask for Block is called!")
	PeerMap := Peers.GetPeerMap()
	fmt.Println("AskForBlock.Size of PeerMap:",len(PeerMap))
	//key is address
	//value is id
	// Send heart beat to every node !
	for key, value := range PeerMap {
		fmt.Printf("key[%s] value[%s]\n", key, value)
		//fmt.Println("height:",height)
		heightString:=ConvertIntToString(height)
		//fmt.Println("heightString:",heightString)
		prepareRequest:="http://"+key+ASK_PEER_REQUEST+"/"+heightString+"/"+hash; //http://localhost:8863:/block/1/323EEFEFEE
		fmt.Println("PrepareRequest:",prepareRequest)
		response, err := http.Get(prepareRequest)
		fmt.Println(">>>>>>>>>>>>>> AskForBlock Sent  To:[", key ,"] >>>>>>>>>>>>>>>>")
		if(response.StatusCode==204){
			fmt.Println("There is no block available block on this Peer:", key)
		}else if(response.StatusCode==500){
			fmt.Println("Internal Server Error happened on this Peer:", key)
		}else{
			if err != nil {
				log.Fatal(err)
			}
			defer response.Body.Close()
			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			responseString := string(responseData)
			fmt.Println("Missing Block's JSON Response:",responseString)
			missingBlock := p2.Block{}
			missingBlock.DecodeFromJson(responseString)
			//SBC.Insert(missingBlock)
			//fmt.Println("Missing Block:",missingBlock)
			//fmt.Println("We found the new block on Peer:",key)
			if !SBC.CheckParentHash(missingBlock) {
				fmt.Println("Recursive Call,missingBlock.Header.Height-1:" ,missingBlock.Header.Height-1,", missingBlock.Header.ParentHash",missingBlock.Header.ParentHash);
				AskForBlock(missingBlock.Header.Height-1,missingBlock.Header.ParentHash)
				fmt.Println("Get Block after recursively AskForBlock:", missingBlock)
				fmt.Println("From peer:", key)

			}
			SBC.Insert(missingBlock)
		}
	}
}

// T.A. : Send the HeartBeatData to all peers in local PeerMap.
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	if heartBeatData.Hops != 0 {
		heartBData, _ := json.Marshal(heartBeatData)
		Peers.Rebalance()
		for addr,_ := range(Peers.Copy()) {
			resp, err := http.Post("http://"+ addr + "/heartbeat/receive",
				"application/json; charset=UTF-8", strings.NewReader(string(heartBData)))

			fmt.Println(">>>>>>>>>>>>>> ForwardHeartBeat Sent  To:[", addr ,"] >>>>>>>>>>>>>>>>")

			if(err != nil || resp.StatusCode != 200) {
				Peers.Delete(addr)
			}
		}
	}
}

//Assume we have launched the first node node1. According to the workflow, after launch the second node node2, it calls "start()".
// Node2 will first go to server and ask for an ID. Then, node2 will download the BlockChain from node1. After that, node2 calls
// "StartHeartBeat()" to start the heartBeat loop. When node2 receives a HeartBeat which contains a new block, node2 will check
// if the parent block exists. If not, node2 will call "AskForBlock" to download that parent block from one of node2's peers.

//T.A. : Start a while loop. Inside the loop, sleep for randomly 5~10 seconds, then use PrepareHeartBeatData() to create a
// HeartBeatData, and send it to all peers in the local PeerMap.
func StartHeartBeat() {
	fmt.Println("Start Heart Beat!")
	//data.NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string)
	Peers.Rebalance()
	//data.PrepareHeartBeatData()
	//Iterate over peer List and send Post request
	PeerMap := Peers.GetPeerMap()
	fmt.Println("Size of PeerMap:",len(PeerMap))
		//key is address
		//value is id
		// Send heart beat to every node !
		for key, value := range PeerMap {
			fmt.Printf("key[%s] value[%d]\n", key, value)
			uploadAddress := "http://" + key + "/heartbeat/receive";
			fmt.Println("/heartbeat/receive Request will be sent to :" + uploadAddress)
			//destination := "http://localhost:6688" +/heartbeat/receive"
			peerMapToJson, err := Peers.PeerMapToJson()
			if err != nil {
				log.Fatal(err)
			}

			mpt:=p1.MerklePatriciaTrie{};

			//heartBearData:= data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapToJson, SELF_ADDR,false,"",mpt )
			heartBearData:= data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapToJson, SELF_ADDR,false,"",mpt )

			jsonBytes, err := json.Marshal(heartBearData)
			req, err := http.NewRequest("POST", uploadAddress, bytes.NewBuffer(jsonBytes))
			req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			fmt.Println(">>>>>>>>>>>> HeartBeatSent To:[",key,"]>>>>>>>>>>>>>")
			if err != nil {
				//panic(err)
				fmt.Println("Problem in peer[",key,"] Deleting peer from Peer List");
				Peers.Delete(key)
				return
			}
			defer resp.Body.Close()
			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("response Body:", string(body))
	}
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

func StartTryingNonce(){
	mpt:=p1.MerklePatriciaTrie{}
	mpt.Initial()
	mpt.Insert(p2.String(2),p2.String(5))
	for {
	GetLatestBlock:
		blocks := SBC.GetLatestBlocks()
		StopGeneratingNewBlock = false
		validateNonce := p2.String(16)
		hashPuzzle := string(blocks[0].Header.Hash) + string(validateNonce) + string(mpt.Root)
		sum := sha3.Sum256([]byte(hashPuzzle))

		if strings.HasPrefix(hex.EncodeToString(sum[:]), SPECIAL_BLOCK_PREFIX){
			fmt.Println("HashPuzzle solved:",time.Now().Unix(), ",hashPuzzel:",	hex.EncodeToString(sum[:]))
			peerMapJson,_ :=Peers.PeerMapToJson()
			heartBeatData :=data.PrepareHeartBeatData(&SBC,Peers.GetSelfId(),peerMapJson,SELF_ADDR,true , validateNonce, mpt)
			ForwardHeartBeat(heartBeatData)
			if StopGeneratingNewBlock {
				fmt.Println("Someone solved hash puzzle before you did! Stop solving...!")
				goto GetLatestBlock
			}
		}
	}
}


