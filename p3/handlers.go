package p3

import (
	"../p1"
	"../p2"
	"../p5"
	"./data"
	"bytes"
	"crypto/rsa"
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
	"sync"
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
var SPECIAL_BLOCK_PREFIX="0000"; //4 zeros...
//var SPECIAL_BLOCK_PREFIX="000000"; //6 zeros...
var newTransactionObject p5.Transaction;
var MinerKey *rsa.PrivateKey
var TxPool TransactionPool;
var userBalance int=1000;
var transactionFee int=10;

type TransactionPool struct {
	Pool      map[string]p5.Transaction `json:"pool"`
	Confirmed map[string]bool        `json:"confirmed"`
	mux       sync.Mutex
}

//Thread Safe
func (txp *TransactionPool) AddToTransactionPool(tx p5.Transaction) { //duplicates in transactinon pool
	txp.mux.Lock()
	defer txp.mux.Unlock()
	if _, ok := txp.Pool[tx.TransactionId]; !ok {
		log.Println("In AddToTransactionPool : Adding new TX:",tx.TransactionId)
		txp.Pool[tx.TransactionId] = tx
	}
}

//Thread Safe
func (txp *TransactionPool) DeleteFromTransactionPool(transactionId string) {
	txp.mux.Lock()
	defer txp.mux.Unlock()
	delete(txp.Pool, transactionId)
	log.Println("In DeleteFromTransactionPool : Deleting  TX:",transactionId)
}

func (txp *TransactionPool) GetTransactionPoolMap() map[string]p5.Transaction{
	//txp.mux.Lock()
	//defer txp.mux.Unlock()
	return txp.Pool
}

//Later use...
func (txp *TransactionPool) GetOneTxFromPool() *p5.Transaction{
	//txp.mux.Lock()
	//defer txp.mux.Unlock()
	if(len(TxPool.GetTransactionPoolMap())>0){
		for _, transactionObject := range TxPool.GetTransactionPoolMap() {
			if userBalance >= transactionObject.TransactionFee {
				transactionObject.Balance = transactionObject.Balance - transactionObject.TransactionFee
				//TODO check how to add
				//fmt.Println("transactionObject.Balance:",transactionObject.Balance)
				return &transactionObject
			}else{
				fmt.Println("User has not got enough balance:",userBalance)
				return nil;
			}
		}
	}
	return nil;
}

func (txp *TransactionPool) AddToConfirmedPool(tx p5.Transaction) { //duplicates in transactinon pool
	txp.mux.Lock()
	defer txp.mux.Unlock()

	//TODO:BUG. Transaction ID's coming "" (NULL) we should return false in that case.
	if(tx.TransactionId==""){
		fmt.Println("Tx ID is NULL. Do not add to CheckConfirmedPool,TX:",tx.TransactionId)
		return
	}
	if _, ok := txp.Confirmed[tx.TransactionId]; !ok {
		log.Println("In AddToConfirmedPool, TX:",tx.TransactionId)
		txp.Confirmed[tx.TransactionId] = true
	}
}


func (txp *TransactionPool) CheckConfirmedPool(tx p5.Transaction) bool { //duplicates in transactinon pool
	txp.mux.Lock()
	defer txp.mux.Unlock()

	//TODO:BUG. Transaction ID's coming "" (NULL) we should return false in that case.
	if(tx.TransactionId==""){
		fmt.Println("Tx ID is NULL. Returning false for CheckConfirmedPool,TX:",tx.TransactionId)
		return false
	}
	if _, ok := txp.Confirmed[tx.TransactionId]; ok {
		fmt.Println("Tx is in ConfirmedPool,TX:",tx.TransactionId)
		return true
	}else{
		fmt.Println("Tx is NOT in ConfirmedPool,TX:",tx.TransactionId)
		return false
	}
}


/*func NewTransactionPool() *TransactionPool {
	Pool :=  make(map[string]p5.Transaction)
	Confirmed:=make(map[string]bool)
	mutex:=sync.Mutex{}
	return &TransactionPool{Pool, Confirmed,mutex}
}*/

func NewTransactionPool() TransactionPool {
	Pool :=  make(map[string]p5.Transaction)
	Confirmed:=make(map[string]bool)
	mutex:=sync.Mutex{}
	return TransactionPool{Pool, Confirmed,mutex}
}

//Create SyncBlockChain and PeerList instances.
func init() {
		fmt.Println("Init method is triggered!")
		//TransactionMap = make(map[string]p5.Transaction);

		TxPool=NewTransactionPool();

		//ExistingTransactionMap = make(map[string]string);

		SBC = data.NewBlockChain()
		Peers = data.NewPeerList(Peers.GetSelfId(),32);
		mpt:=p1.MerklePatriciaTrie{}
		mpt.Initial()
		//mpt.Insert(p2.String(2),p2.String(5))
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
			//TODO: First Peer's Public Key hard coded????
			publicKey,_:=p5.ParseRsaPublicKeyFromPemStr("HARD_CODED_PEER1")
			Peers.AddPublicKey(publicKey,6686);
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

	MinerKey =p5.GenerateKeyPair(2014)

	fmt.Println("Public Key:", MinerKey.PublicKey)
	fmt.Println("Private Key::", MinerKey)


	//publicKeyAsPemStr,_:=p5.ExportRsaPublicKeyAsPemStr(&MinerKey.PublicKey);

	Peers.AddPublicKey(&MinerKey.PublicKey,Peers.GetSelfId());
	//Peers.Add(publicKeyAsPemStr,Peers.GetSelfId())

	go StartTryingNonce()
	//StartTryingNonce()

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
	fmt.Fprintf(w, "%s\n%s", Peers.Show2(), SBC.Show())
	//fmt.Fprintf(w, "%s\n%s", Peers.ShowPublicMap(), SBC.Show())
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
	var mutex = sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	fmt.Println("<<<<<<<<<<<<<< HeartBeat Received  From:[", r.Host, "] <<<<<<<<<<<<")
	fmt.Fprintf(w, "%s\n", r.Host)
	fmt.Fprintf(w, "%s", string(data))
	error := json.Unmarshal(data, &HeartBeatVariable)
	if (error != nil) {
		fmt.Println("Error occured in HeartBeatReceive: ", error)
	} else {

	}
	//fmt.Println("HeartBeatVariable.Addr", HeartBeatVariable.Addr)
	//fmt.Println("SELF_ADDR", SELF_ADDR)
	if (HeartBeatVariable.Addr == SELF_ADDR) {
		return
	}
	transaction := p5.Transaction{}
	transaction.DecodeFromJSON(HeartBeatVariable.TransactionInfoJson)
	if (transaction.TransactionId != "") {
		fmt.Println("TransactionId:", transaction.TransactionId,
			" came as a heartbeat from <<<<<<<<<<<<<< From:[", r.Host, "] <<<<<<<<<<<<")
	}
	//new added...
	Peers.AddPublicKey(HeartBeatVariable.PeerPublicKey, HeartBeatVariable.Id);
	Peers.Add(HeartBeatVariable.Addr, HeartBeatVariable.Id)
	Peers.InjectPeerMapJson(HeartBeatVariable.PeerMapJson, SELF_ADDR)
	//if (HeartBeatVariable.IfNewBlock && HeartBeatVariable.IfValidTransaction) {
	if (HeartBeatVariable.IfNewBlock) {
		fmt.Println("HeartBeat flag is true!!")

		heartBlock := p2.Block{}
		heartBlock.DecodeFromJson(HeartBeatVariable.BlockJson)
		fmt.Println("Received block! Root:",heartBlock.Value.Root)
		// Proof of Work...
		receivedPuzzle := heartBlock.Header.ParentHash + heartBlock.Header.Nonce + heartBlock.Value.Root
		sum := sha3.Sum256([]byte(receivedPuzzle))

		//I can not come to this point when TX is not NULL!!
		if (strings.HasPrefix(hex.EncodeToString(sum[:]), SPECIAL_BLOCK_PREFIX)) {
			fmt.Println("Block with SPECIAL PREFIX arrived from:[", r.Host, "]")
			latestBlocks := SBC.GetLatestBlocks()
			for i := 0; i < len(latestBlocks); i++ {
				if latestBlocks[i].Header.Hash == heartBlock.Header.ParentHash {
					StopGeneratingNewBlock = true
					break
				}
			}
			if (heartBlock.Header.Height == 1) {
				//TODO: BUG_FIX Check TX if it is inserted before!
				//TODO:BUG_FIX After we are adding block to the SBC, we need to delete TX from TxPool.
				if (TxPool.CheckConfirmedPool(transaction) == false) {
					SBC.Insert(heartBlock)
					TxPool.DeleteFromTransactionPool(transaction.TransactionId)

					fmt.Println("NEW.Tx.Pool.Size:",len(TxPool.GetTransactionPoolMap()))

					TxPool.AddToConfirmedPool(transaction)
				} else {
					fmt.Println("Do Not Insert TX. It already confirmed,TX:", transaction.TransactionId)
					TxPool.DeleteFromTransactionPool(transaction.TransactionId)
				}
			} else {
				_, flag := SBC.GetBlock(heartBlock.Header.Height-1, heartBlock.Header.ParentHash)
				if flag {
					fmt.Println("No.Gap.Inserting Heart Beat Block:", heartBlock)
					//TODO: BUG_FIX Check TX if it is inserted before!
					//TODO:BUG_FIX After we are adding block to the SBC, we need to delete TX from TxPool.
					//if (TxPool.CheckConfirmedPool(transaction) == false&&HeartBeatVariable.Balance!=0) {
					if (TxPool.CheckConfirmedPool(transaction) == false) {
						SBC.Insert(heartBlock)
						fmt.Println("No.Gap.New.SBC:", SBC)
						TxPool.DeleteFromTransactionPool(transaction.TransactionId)
						fmt.Println("NEW.Tx.Pool.Size:",len(TxPool.GetTransactionPoolMap()))
						TxPool.AddToConfirmedPool(transaction)
					} else {
						fmt.Println("Do Not Insert TX. It already confirmed,TX:", transaction.TransactionId)
						TxPool.DeleteFromTransactionPool(transaction.TransactionId)
					}
				} else {
					fmt.Println("Gap.Inserting Heart Beat Block:", heartBlock)
					AskForBlock(heartBlock.Header.Height-1, heartBlock.Header.ParentHash)
					//TODO: BUG_FIX Check TX if it is inserted before!
					//TODO:BUG_FIX After we are adding block to the SBC, we need to delete TX from TxPool.
					if (TxPool.CheckConfirmedPool(transaction) == false) {
						SBC.Insert(heartBlock)
						fmt.Println("Gap.New.SBC:", SBC)
						TxPool.DeleteFromTransactionPool(transaction.TransactionId)
						fmt.Println("NEW.Tx.Pool.Size:",len(TxPool.GetTransactionPoolMap()))
						TxPool.AddToConfirmedPool(transaction)
					} else {
						fmt.Println("Do Not Insert TX. It already confirmed,TX:", transaction.TransactionId)
						TxPool.DeleteFromTransactionPool(transaction.TransactionId)
					}
				}
			}


		} else {
			fmt.Println(" Ignoring incoming Heart Beat Block! Unmatched Puzzle! Calculated Puzzle:", hex.EncodeToString(sum[:]))
			fmt.Println("TEST.ALPER.RECEIVE********************************************************")
			fmt.Println("Incoming Heart Beat Block.Hash::", heartBlock.Header.Hash)
			fmt.Println("Incoming Heart Beat Nonce:", heartBlock.Header.Nonce)
			fmt.Println("Incoming Heart Beat mpt.Root:", heartBlock.Value.Root)
			fmt.Println("Calculated Incoming Hash Puzzle:", hex.EncodeToString(sum[:]))
		}
	} else {
		fmt.Println("HeartBeat flag is false! There is no block in heartBeat!")
	}
	HeartBeatVariable.Hops -= 1
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

			//if HeartBeatVariable.Balance==0 {
				//		fmt.Println("Transaction already inserted. Do not insert again, TX:",transaction.TransactionId)
			//}else{
				SBC.Insert(missingBlock)
			//}
		}
	}
}

// T.A. : Send the HeartBeatData to all peers in local PeerMap.
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	if heartBeatData.Hops != 0 {
		heartBData, _ := json.Marshal(heartBeatData)
		fmt.Println("Peers.Size:",len(Peers.GetPeerMap()))
		Peers.Rebalance()
		fmt.Println("ForwardHeartBeat.TransactionJSON:",heartBeatData.TransactionInfoJson)
		for addr,_ := range(Peers.Copy()) {

			transaction := p5.Transaction{}
			transaction.DecodeFromJSON(heartBeatData.TransactionInfoJson)

			resp, err := http.Post("http://"+addr+"/heartbeat/receive",
				"application/json; charset=UTF-8", strings.NewReader(string(heartBData)))



			if (transaction.TransactionId != "") {
				fmt.Println("TransactionId:", transaction.TransactionId,
					" is sent with heartbeat to >>>>>>>>>>>>>> ForwardHeartBeat Sent  To:[", addr, "] >>>>>>>>>>>>>>>>")
			}else{
				fmt.Println("ForwardHeartBeat.TransactionId is NULL!! TransactionJSON:",heartBeatData.TransactionInfoJson)
			}

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
			heartBearData:= data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapToJson, SELF_ADDR,false,"",mpt ,&MinerKey.PublicKey,false,"",0)

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

//TODO : Thread Safe
func StartTryingNonce() {
	var mutex = sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	isValidTransaction := false
	//forever...
	for {
	GetLatestBlock:
		newMpt := p1.MerklePatriciaTrie{}
		newMpt.Initial()
		blocks := SBC.GetLatestBlocks()
		StopGeneratingNewBlock = false
		var transactionJSON string
		//Get Thread Safe 1 TX object from Pool.
		transaction := TxPool.GetOneTxFromPool()
		if (transaction != nil) {

			HeartBeatVariable.Balance = userBalance - transaction.TransactionFee
			transactionJSON, _ = transaction.EncodeToJSON();

			//TODO: Check balance and the other things here later...
			newMpt.Insert(transaction.TransactionId, transactionJSON)
			//fmt.Println("POW..")
			validateNonce := p2.String(16)
			hashPuzzle := string(blocks[0].Header.Hash) + string(validateNonce) + string(newMpt.Root)
			sum := sha3.Sum256([]byte(hashPuzzle))

			//fmt.Printf("Found one TX.Plate:%s - Nonce:%s",transaction.Plate,hex.EncodeToString(sum[:5]))
			if strings.HasPrefix(hex.EncodeToString(sum[:]), SPECIAL_BLOCK_PREFIX) {
				fmt.Println("***********************************************************************************")
				fmt.Println("*** HashPuzzle solved:", time.Now().Unix(), ",hashPuzzel:", hex.EncodeToString(sum[:]))
				fmt.Println("***********************************************************************************")
				peerMapJson, _ := Peers.PeerMapToJson()
				transactionJSON, _ = transaction.EncodeToJSON()
				//newMpt.Insert(tempTransactionObject.TransactionId,"apple")
				fmt.Println("test.mpt:", newMpt);

				heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, true, validateNonce,
					newMpt, &MinerKey.PublicKey, isValidTransaction, transactionJSON, HeartBeatVariable.Balance)

				heartBlock := p2.Block{}
				heartBlock.DecodeFromJson(heartBeatData.BlockJson)

				fmt.Println("heartBlock.Root:",string(heartBlock.Value.Root))

				testPuzzle:=string(heartBlock.Header.ParentHash) + string(heartBlock.Header.Nonce) + string(heartBlock.Value.Root)
				sum = sha3.Sum256([]byte(testPuzzle))
				fmt.Println("testPuzzle:",	hex.EncodeToString(sum[:]))

				if(heartBeatData.IfNewBlock){
					fmt.Println("Yes.Send new block")
				}else{
					fmt.Println("Wierd thing happened!")
				}

				fmt.Println("startTryingNonce.TransactionJSON:",heartBeatData.TransactionInfoJson)
				ForwardHeartBeat(heartBeatData)
				isValidTransaction = false //TODO:NOT USING FOR NOW!! CRITICAL!!!!!
				fmt.Println("******** Miner solved the Puzzle and took the TX from Transaction Map!")
				fmt.Println("******** Before.TransactionMap.Size:", len(TxPool.GetTransactionPoolMap()))
				//delete(TransactionMap, tempTransactionObject.TransactionId);
				TxPool.DeleteFromTransactionPool(transaction.TransactionId)
				fmt.Println("******** After.TransactionMap.Size:", len(TxPool.GetTransactionPoolMap()))
				if StopGeneratingNewBlock {
					fmt.Println("Stop generating node!")
					goto GetLatestBlock
				}
			}
		}else{//if transaction is NOT NULL....
			//fmt.Println("No Transaction in TxPool.")
		}
	}//for forever
}


func CarFormAPI(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCarForm method is triggered!")

	//if r.URL.Path != "/getCarForm" {
	//	http.Error(w, "404 not found.", http.StatusNotFound)
	//	return
	//}

	switch r.Method {
	case "GET":
		log.Println("GET CarForm triggered!")

		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("PWD:",dir)

		http.ServeFile(w, r, "CarForm.html")
	case "POST":
		log.Println("POST CarForm triggered!")

		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		userBalance = userBalance-transactionFee;

		fmt.Fprintf(w, "HTTP Post sent to Registeration Server! PostForm = %v\n", r.PostForm)
		plate := r.FormValue("plate")
		mileage := r.FormValue("mileage")
		fmt.Fprintf(w, "plate = %s\n", plate)
		fmt.Fprintf(w, "mileage = %s\n", mileage)
		fmt.Fprintf(w, "User's New Balance = %d\n", userBalance)

		transactionId:=p2.String(8)
		transactionFee:=10;
		i64, _ := strconv.ParseInt(mileage, 10, 32)
		mileageInt := int32(i64)
		newTransactionObject =p5.NewTransaction(transactionId,mileageInt,plate,transactionFee,userBalance);
		fmt.Println("Transaction:", newTransactionObject);
		fmt.Println("StartTryingNonce.NewTransaction:",newTransactionObject);
		transactionJSON,_:=newTransactionObject.EncodeToJSON()
		fmt.Println("Transaction JSON:",transactionJSON)

		//mpt.Insert(transactionId,transactionJSON)
		//fmt.Println("mpt:",mpt)

		PeerPublicKeyMap := Peers.GetPeerPublicKeyMap()
		fmt.Println("Size of PeerPublicKeyMap:",len(PeerPublicKeyMap))
		//key is address
		//value is id
		// Send heart beat to every node !
		/*for publicKey, port := range  PeerPublicKeyMap {
			fmt.Printf("key[%s] value[%s]\n", publicKey, port)
			cipherTextToMiner, hash, label, _:=p5.Encrypt(transactionJSON,&MinerKey.PublicKey);
			fmt.Println("cipherTextToMiner is:", cipherTextToMiner )
			fmt.Println("hash is:", hash )
			fmt.Println("label is:", label )
			signature, opts, hashed, newhash, _:= p5.Sign(cipherTextToMiner, MinerKey) //Private Key
			fmt.Println("User Signature is:", signature)
			fmt.Println("opts is:", opts)
			fmt.Println("hashed is:", hashed)
			fmt.Println("newhash is:", newhash)
		}*/

		//plainTextfrom, _ := p5.Decrypt(cipherTextToMiner, hash , label ,minerKey.PrivateKey)
		//fmt.Println("plainTextfrom is:", plainTextfrom)

		//isVerified, _ := p5.Verification (Key.PublicKey, opts, hashed, newhash, signature)
		//fmt.Println("Is Verified is:", isVerified)

		//TransactionMap[transactionId] = newTransactionObject;
		go TxPool.AddToTransactionPool(newTransactionObject);
		//TxPool.Pool[newTransactionObject.TransactionId]=newTransactionObject
		//fmt.Println("TransactionMap Size:",len(TransactionMap))
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}


func QueryCarAPI(w http.ResponseWriter, r *http.Request) {
	log.Println("QueryCarAPI method is triggered!")

	switch r.Method {
	case "GET":
		log.Println("GET QueryCarAPI triggered!")

		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("PWD:",dir)

		http.ServeFile(w, r, "QueryCar.html")
	case "POST":
		log.Println("POST QueryCarAPI triggered!")

		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		//fmt.Fprintf(w, "HTTP Post sent to Server! PostForm = %v\n", r.PostForm)
		plate := r.FormValue("plate")
		fmt.Fprintf(w, "plate = %s\n", plate)
		fmt.Fprintf(w, "SEARCH RESULT = %s\n", SBC.GetCarInformation(plate))

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}




