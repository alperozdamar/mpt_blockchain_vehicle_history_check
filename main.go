package main

import (
	"reflect"

	"./p5"
	"fmt"
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

	//KEY GENERATION
	key:=p5.GenerateKeyPair(2048);

	fmt.Println("publicKey:",&key.PublicKey)
	fmt.Println("privateKey::",key)

	plainText := "AlperWillNeverGiveUp"
	fmt.Println("Before Encrpytion:\n",plainText)
	plainTextByteArray:=[]byte(plainText)
	fmt.Println("Before Encrpytion plainTextByteArray:\n",plainTextByteArray)

	//ENCRYPTION
	ciphertext:=p5.EncryptWithPublicKey(plainTextByteArray,&key.PublicKey);
	fmt.Println("CipherText:\n",ciphertext)
	//DECRYPTION
	decryptedPlaintextByteArray :=p5.DecryptWithPrivateKey(ciphertext,key)
	fmt.Println("resultPlaintextByteArray:\n",decryptedPlaintextByteArray)
	resultPlaintext := string(decryptedPlaintextByteArray);
	fmt.Println("resultPlaintext:\n",resultPlaintext)

	if(reflect.DeepEqual(plainTextByteArray, decryptedPlaintextByteArray)){
		fmt.Println("YES THEY ARE SAME PERFECT")
	}else{
		fmt.Errorf("NO THEY ARE NOT SAME.THERE IS A PROBLEM")
	}


	//router := p3.NewRouter()
	//if len(os.Args) > 1 {
	//	log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	//} else {
	//	log.Fatal(http.ListenAndServe(":6686", router))
	//}
}






