package data

import (
	"container/ring"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId int32
	peerMap map[string]int32
	peerPublicKeyMap map[*rsa.PublicKey]int32
	maxLength int32
	mux sync.Mutex
}

// Pair - data structure to hold a key/value pair - addr/id.
type Pair struct {
	addr string
	id   int32
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      {
	p[i], p[j] = p[j], p[i]
}

func (p PairList) Len() int           {
	return len(p)
}

func (p PairList) Less(i, j int) bool {
	return p[i].id < p[j].id
}

func (peers *PeerList) EncodePeerMapToJSON() (string, error) {
	jsonBytes, err := json.Marshal(peers.peerMap)
	return string(jsonBytes), err
}

func (f PeerList) GetPeerMap() map[string]int32{
	return f.peerMap
}

func (f PeerList) GetPeerPublicKeyMap() map[*rsa.PublicKey]int32{
	return f.peerPublicKeyMap
}

func (f PeerList) GetMux() sync.Mutex{
	return f.mux
}

func (f PeerList) GetMaxLength() int32{
	return f.maxLength
}

// A function to turn a map into a PairList, then sort and return it.
func sortMapByValue(m map[string]int32) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		//fmt.Println("in sortMapByValue : k, v :", k, v)
		p[i] = Pair{
			addr: k,
			id:   int32(v),
		}
		//fmt.Println("in sortMapByValue : p[i] :", p[i])
		i++
	}
	sort.Sort(p)
	return p
}

func NewPeerList(id int32, maxLength int32) PeerList {
	peerList := PeerList{
		peerMap: make(map[string]int32),
		peerPublicKeyMap: make(map[*rsa.PublicKey]int32),
		maxLength: maxLength}
	peerList.Register(id)
	return peerList
}

func NewPeerPublicKeyList(id int32, maxLength int32) PeerList {
	peerList := PeerList{peerPublicKeyMap: make(map[*rsa.PublicKey]int32), maxLength: maxLength}
	peerList.Register(id)
	return peerList
}

func(peers *PeerList) Add(addr string, id int32) {
	peers.peerMap[addr]=id;
}

func(peers *PeerList) AddPublicKey(publicKey *rsa.PublicKey, id int32) {
	peers.peerPublicKeyMap[publicKey]=id;
}

func(peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

//Rebalance func changes the PeerMap to contain take maxLength(32) closest peers (by Id)
func (peers *PeerList) Rebalance() {
	if int32(len(peers.peerMap)) > peers.maxLength {
		//fmt.Println("in Rebalance")
		//fmt.Println("in Rebalance : original peerMap length : ", len(peers.peerMap))
		peers.peerMap["selfAddr"] = peers.selfId //adding self id to peerMap
		sortedAddrIDList := sortMapByValue(peers.peerMap)
		//fmt.Println("in Rebalance : sortedAddrIDList : ", sortedAddrIDList)
		sortedAddrIDListLength := len(sortedAddrIDList)
		//fmt.Println("in Rebalance : sortedAddrIDListLength : ", sortedAddrIDListLength)
		peers.peerMap = peers.getBalancedPeerMap(sortedAddrIDListLength, sortedAddrIDList)
	}
}

func (peers *PeerList) getBalancedPeerMap(sortedAddrIDListLength int, sortedAddrIDList PairList) map[string]int32 {
	r := ring.New(sortedAddrIDListLength) // new ring
	useRingPtr := r
	//initialize ring with sortedAddrIDList values
	for i := 0; i < sortedAddrIDListLength; i++ {
		r.Value = sortedAddrIDList[i]
		//fmt.Println("in Rebalance : r.Value : ", r.Value)
		if sortedAddrIDList[i].id == peers.selfId {
			useRingPtr = r
			//fmt.Println("in Rebalance : useRingPtr : ", useRingPtr)
		}
		r = r.Next()
	}
	newPeerMap := make(map[string]int32)
	r = useRingPtr
	//fmt.Println("in Rebalance : useRingPtr : ", useRingPtr)
	for i := 1; i <= int(peers.maxLength/2); i++ {
		r = r.Prev()
		pair := r.Value.(Pair)
		newPeerMap[pair.addr] = pair.id
	}
	r = useRingPtr
	for i := 1; i <= int(peers.maxLength/2); i++ {
		r = r.Next()
		pair := r.Value.(Pair)
		newPeerMap[pair.addr] = pair.id
	}

	return newPeerMap
}

func(peers *PeerList) Show() string {
	rs := ""
	peers.mux.Lock()
	defer peers.mux.Unlock()
	for addr, id := range peers.peerMap {
		rs += fmt.Sprintf("addr= %s ", addr)
		rs += fmt.Sprintf(", id= %d \n", id)
	}
	rs += "\n"
	//rs = fmt.Sprintf("This is the PeerMap: %s\n", hex.EncodeToString(sum[:])) + rs
	rs = fmt.Sprintf("This is the PeerMap: \n") + rs
	fmt.Print(rs)
	return  rs
}

func(peers *PeerList) Show2() string {
	rs := ""
	peers.mux.Lock()
	defer peers.mux.Unlock()
	for addr, id := range peers.peerMap {
		rs += fmt.Sprintf("addr= %s ", addr)
		rs += fmt.Sprintf(", id= %d \n", id)
	}
	rs += "\n"
	for publicKey, id := range peers.peerPublicKeyMap {
		rs += fmt.Sprintf("public Key= %s ", publicKey)
		rs += fmt.Sprintf(", id= %d \n", id)
	}
	rs += "\n"
	//rs = fmt.Sprintf("This is the PeerMap: %s\n", hex.EncodeToString(sum[:])) + rs
	rs = fmt.Sprintf("This is the PeerMap: \n") + rs
	fmt.Print(rs)
	return  rs
}


func(peers *PeerList) ShowPublicMap() string {
	rs := ""
	peers.mux.Lock()
	defer peers.mux.Unlock()
	for publicKey, id := range peers.peerPublicKeyMap {
		rs += fmt.Sprintf("public Key= %s ", publicKey)
		rs += fmt.Sprintf(", id= %d \n", id)
	}
	rs += "\n"
	//rs = fmt.Sprintf("This is the PeerMap: %s\n", hex.EncodeToString(sum[:])) + rs
	rs = fmt.Sprintf("This is the PeerPublicKeyMap: \n") + rs
	fmt.Print(rs)
	return  rs
}

func(peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	copyOfPeerMap := make(map[string]int32)
	for k := range peers.peerMap {
		copyOfPeerMap[k] = peers.peerMap[k]
	}
	peers.mux.Unlock()
	return copyOfPeerMap
}

func(peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

//The "PeerMapJson" in HeartBeatData is the JSON format of "PeerList.peerMap". It is the result of "PeerList.PeerMapToJSON()" function.
// Sorry for the confused argument name "PeerMapBase64" in PerpareHeartBeatData().
// json.Marshal(peers.peerMap)
func(peers *PeerList) PeerMapToJson() (string, error) {
	jsonBytes, err := json.Marshal(peers.peerMap)
	return string(jsonBytes), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	var newPeerMap map[string]int32
	err := json.Unmarshal([]byte(peerMapJsonStr), &newPeerMap)
	if err == nil {
		peers.mux.Lock()

		for k := range newPeerMap {
			if k != selfAddr {
				peers.peerMap[k] = newPeerMap[k]
			}
		}

		peers.mux.Unlock()
	}
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}