package p5

import (
	"encoding/json"
)

type Car struct {
	plate string        					//`json:"header"`
	mileage int32					 		//`json:"value"`
}

func (car *Car) EncodeToJson() (string, error) {
	jsonBytes, error := json.Marshal(car)
	return string(jsonBytes), error
}

func (car *Car) DecodeFromJson(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), car)
}

func (car *Car) EncodeToJSON() (string, error) {
	jsonBytes, error := json.Marshal(car)
	return string(jsonBytes), error
}

func (car *Car) DecodeFromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), car)
}
