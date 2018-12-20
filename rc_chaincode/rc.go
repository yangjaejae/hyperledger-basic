package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {

}

// ----- Wallet ----- //
type Wallet struct {
	Value 		uint64 			`json:"value"`		// Balance
	Transfer	TransferInfo	`json:"transfer`	// Transfer Information
}

// ----- Transfer information ----- //
type TransferInfo struct {
	FromOrTo	string 	`json:"fromOrTo"`	// Collaborator
	Value 		uint64 	`json:"value"`		// Remittance amount
	Date 		string 	`json:"date"`		// Transfer Date
	TxType 		string 	`json:"type"`		// Transfer Type	0: Publish(By Admin)
											// 					1: Payment(By Sender) 				2: Payment(By Recipient)
											// 					3: Cancel Payment(By Sender) 		4: Cancel Payment(By Recipient)	
											// 					5: Remittance(By Sender), 			6: Remittance(By Recipient)
											// 					7: Cancel Remittance(By Sender) 	8: Cancel Remittance(By Recipient)	
}

// ============================================================================================================================
// 	Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

// ============================================================================================================================
// 	Init
// ============================================================================================================================
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// ============================================================================================================================
// 	Invoke
//	init_wallet	:	invoke '{"Args":["init_wallet", "1"]}'
//	publish		:	invoke '{"Args":["publish", "1", "10", "10000", "20181212"]}'
//	transfer	:	invoke '{"Args":["transfer", "1", "2", "1000", "3", "20181212"]}'
//	get_account	:	query '{"Args":["get_account", "1"]}'
//	get_txList	:	query '{"Args":["get_txList", "1"]}'
// ============================================================================================================================
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "init" {
		return s.Init(stub)
	} else if function == "init_wallet" {
		return init_wallet(stub, args)
	} else if function == "publish" {
		return publish(stub, args)
	} else if function == "transfer" {
		return transfer(stub, args)
	} else if function == "get_account" {
		return get_account(stub, args)
	} else if function == "get_txList" {
		return get_txList(stub, args)
	}

	return shim.Error(fmt.Sprintf("Received unknown invoke function name: %s", function));
}

// ============================================================================================================================
//	init_wallet
//	- params: key
//	- return: walletAsBytes
// ============================================================================================================================
func init_wallet(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var newWallet = Wallet {
		Value 		: 0,
	}

	walletAsBytes, _ := json.Marshal(newWallet)
	err := stub.PutState(args[0], walletAsBytes)

	if (err != nil) {
		return shim.Error(fmt.Sprintf("Failed to create Wallet: %s", args[0]));
	}

	return shim.Success(walletAsBytes)
}

// ============================================================================================================================
//	publish
//	- params: key, from, value, date
//	- return: walletAsBytes
// ============================================================================================================================
func publish(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var wallet Wallet
	
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	walletAsBytes, _ := stub.GetState(args[0])
	if walletAsBytes == nil {
		return shim.Error("Not Found wallet : %s", )
	}
	
	json.Unmarshal(walletAsBytes, &wallet)
	value, _ := strconv.ParseUint(args[2], 10, 32)
	
	wallet.Value += value
	wallet.Transfer.FromOrTo = args[1]
	wallet.Transfer.Value = value
	wallet.Transfer.TxType = "0"	// 0 is publish
	wallet.Transfer.Date = args[3]

	walletAsBytes, _ = json.Marshal(wallet)
	err := stub.PutState(args[0], walletAsBytes)
	if (err != nil) {
		return shim.Error("Failed to publish");
	}

	return shim.Success(walletAsBytes)
}

// ============================================================================================================================
//	transfer
//	- params: key, Collaborator, value, transfer_type, date
//	- return: txid
// ============================================================================================================================
func transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	from := Wallet{}
	to := Wallet{}

	fromAsBytes, _ := stub.GetState(args[0])
	toAsBytes, _ := stub.GetState(args[1])

	value, _ := strconv.ParseUint(args[2], 10, 32)
	toType, _ := strconv.Atoi(args[3])
	toType += 1
	fromType := strconv.Itoa(toType)

	if fromAsBytes == nil || toAsBytes == nil {
		return shim.Error("Not found wallet")
	}

	json.Unmarshal(fromAsBytes, &from)
	json.Unmarshal(toAsBytes, &to)
	
	if from.Value < value {
		return shim.Error(fmt.Sprintf("%s is not enough balance.", args[0]))
	}
	
	from.Value -= value
	from.Transfer.FromOrTo = args[1]
	from.Transfer.Value = value
	from.Transfer.TxType = args[3]
	from.Transfer.Date = args[4]

	to.Value += value
	to.Transfer.FromOrTo = args[0]
	to.Transfer.Value = value
	to.Transfer.TxType = fromType
	to.Transfer.Date = args[4]

	fromAsBytes, _ = json.Marshal(from)
	toAsBytes, _ = json.Marshal(to)

	err := stub.PutState(args[0], fromAsBytes)
	if (err != nil) {
		return shim.Error(fmt.Sprintf("Failed to transfer: %s", err.Error));
	}

	txid := stub.GetTxID()

	err = stub.PutState(args[1], toAsBytes)
	if (err != nil) {
		return shim.Error(fmt.Sprintf("Failed to transfer: %s", err.Error));
	}

	return shim.Success([]byte(txid))
}

// ============================================================================================================================
// 	get_account
//	- params: key
//	- return: value
// ============================================================================================================================
func get_account(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var wallet Wallet
	
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	 }
  
	 walletAsBytes, _ := stub.GetState(args[0]);
	 if walletAsBytes == nil {
		return shim.Error("Could not locate Wallet")
	 }

	 json.Unmarshal(walletAsBytes, &wallet)
	 value := fmt.Sprint(wallet.Value)

	 return shim.Success([]byte(value))
}

// ============================================================================================================================
// 	get_txList
//	- params: key
//	- return: []historyAsBytes
// ============================================================================================================================
func get_txList(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	type get_History struct {
		TxId    string   	`json:"txId"`
		Value   Wallet   	`json:"value"`
	 }
	 var history []get_History;
	 var wallet Wallet
  
	 if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	 }
  
	 transferId := args[0]
	 fmt.Printf("- start getHistoryForMarble: %s\n", transferId)
  
	 resultsIterator, err := stub.GetHistoryForKey(transferId)
	 if err != nil {
		return shim.Error(err.Error())
	 }
	 defer resultsIterator.Close()
  
	 for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
		   return shim.Error(err.Error())
		}
  
		var tx get_History
		tx.TxId = historyData.TxId                     
		json.Unmarshal(historyData.Value, &wallet)    
		if historyData.Value == nil {                 
		   var emptyWalletHistory Wallet
		   tx.Value = emptyWalletHistory                
		} else {
		   json.Unmarshal(historyData.Value, &wallet) 
		   tx.Value = wallet                      
		}
		history = append(history, tx)   
	 }
	 
	 fmt.Printf("- getHistoryForMarble returning:\n%s", history)
  
	 historyAsBytes, _ := json.Marshal(history)     //convert to array of bytes
	 return shim.Success(historyAsBytes)  
}