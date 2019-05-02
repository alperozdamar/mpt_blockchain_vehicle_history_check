package p5

import (
	"encoding/json"
)

type Transaction struct {
	transactionId	string 		`json:"transactionId"`
	carId 			string		`json:"carId"` 		//can be car's pilate (Unique Value)
	mileage  		int32		`json:"mileage"`
	plate			string 		`json:"plate"`
	transactionFee	int32		`json:"transactionFee"`
}

func (transaction *Transaction) EncodeToJson() (string, error) {
	jsonBytes, error := json.Marshal(transaction)
	return string(jsonBytes), error
}

func (transaction *Transaction) DecodeFromJson(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), transaction)
}

func (transaction *Transaction) EncodeToJSON() (string, error) {
	jsonBytes, error := json.Marshal(transaction)
	return string(jsonBytes), error
}

func (transaction *Transaction) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), transaction)
}
