package data

import "encoding/json"

type RegisterData struct {
	AssignedId int32 `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	r:=RegisterData{AssignedId:id,PeerMapJson:peerMapJson};
	return r
}

func (data *RegisterData) EncodeToJson() (string, error) {

	jsonBytes, error := json.Marshal(data)
	return string(jsonBytes), error

}

func (data *RegisterData) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), data)
}