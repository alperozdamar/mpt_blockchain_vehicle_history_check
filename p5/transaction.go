package p5

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	TransactionId	string 		`json:"transactionId"`
	Mileage  		int32		`json:"mileage"`
	Plate			string 		`json:"plate"` //car's pilate (Unique Value)
	TransactionFee	int		    `json:"transactionFee"`
	Balance			int	        `json:"balance"`
	Time			time.Time   `json:"time"`
	ServiceName     string	    `json:"service"`
}

func NewTransaction(transactionId string, mileage int32, plate string,transactionFee int,balance int,time time.Time,serviceName string) (Transaction)   {
	return Transaction{
			TransactionId:transactionId,
			Mileage:mileage,
			Plate:plate,
			TransactionFee:transactionFee,
			Balance:balance,
			Time:time,
			ServiceName:serviceName,
		}
}

func (transaction *Transaction) EncodeToJSON() (string, error) {
	//fmt.Println("test.transaction:",transaction)
	jsonBytes, error := json.Marshal(transaction)
	//fmt.Println("test.jsonBytes:",string(jsonBytes))
	return string(jsonBytes), error
}

func (transaction *Transaction) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), transaction)
}

//func (transaction *Transaction) TransactionVerification(transactionId string) error {
//
//	//Verify the signature. (Decrypt)
//
//
//
//	//Verify the balance.
//	return nil;
//}