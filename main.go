/*

DECODED HYPERLEDGER APPLICATION

DecodedChainCode functions:
- main, Init, Invoke, Query (standard and required)
- read: reads the contents of a specific key.
- readAll: reads all the primary keys.
- getDataArrayStrings - private function
- saveStringToDataArray
- saveLedger

*/

package main


import (
    "encoding/json"
    "errors"

    "github.com/hyperledger/fabric/core/chaincode/shim"

    utils "github.com/DecodedCo/blockchain-golang-chaincode/utils"
)


// ============================================================================================================================


// Create the struct to tie the methods to.
type DecodedChainCode struct {
}

var PRIMARYKEY = [4]string{ "Owners", "Assets", "Transactions", "PendingTransactions" }


// ============================================================================================================================
// Main


func main() {
    err := shim.Start(new(DecodedChainCode))
    if err != nil {
        utils.PrintErrorFull("Initialisation error", err)
    } else {
        utils.PrintStatus("Started chaincode...")
    }
    // Set logging level
    logLevel, _ := shim.LogLevel("ERROR")
    shim.SetLoggingLevel(logLevel)
}


// ============================================================================================================================


// Init resets all the things
func (dcc *DecodedChainCode) Init(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Initialise the empty datastores. The owners and assets datastores are going to store an array of all keys/ids.
    var blank []string
    blankBytes, err := json.Marshal(&blank)
    if err != nil {
        utils.PrintErrorFull("Init - Marshal", err)
        return nil, err
    }
    //
    if err = stub.PutState(PRIMARYKEY[0], blankBytes); err != nil {
        utils.PrintErrorFull("Init", err)
        return nil, err
    }
    if err = stub.PutState(PRIMARYKEY[1], blankBytes); err != nil {
        utils.PrintErrorFull("Init", err)
        return nil, err
    }
    if err = stub.PutState(PRIMARYKEY[2], blankBytes); err != nil {
        utils.PrintErrorFull("Init", err)
        return nil, err
    }
    if err = stub.PutState(PRIMARYKEY[3], blankBytes); err != nil {
        utils.PrintErrorFull("Init", err)
        return nil, err
    }
    // Done.
    utils.PrintSuccess("Initialisation complete")
    return nil, nil
} 


// Invoke is our entry point to invoke a chaincode function
func (dcc *DecodedChainCode) Invoke(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    // Handle different functions
    if fn == "init" { //initialize the chaincode state, used as reset
        return dcc.Init(stub, fn, args)
    } else if fn == "addOwner" {
        return dcc.addOwner(stub, fn, args)
    } else if fn == "addOwnerString" {
        return dcc.addOwnerString(stub, fn, args)
    } else if fn == "updateOwner" {
        return dcc.updateOwner(stub, fn, args)
    } else if fn == "changeOwnerValidationStatus" { // read all owners and return full data for them.
        return dcc.changeOwnerValidationStatus(stub, fn, args)
    } else if fn == "addAssetString" {
        return dcc.addAssetString(stub, fn, args)
    } else if fn == "updateAsset" {
        return dcc.updateAsset(stub, fn, args)
    } else if fn == "transactAsset" {
        return dcc.transactAsset(stub, fn, args)
    } else if fn == "approveTransaction" {
        return dcc.approveTransaction(stub, fn, args)
    } else if fn == "declineTransaction" {
        return dcc.declineTransaction(stub, fn, args)
    }
    // In any other case.
    utils.PrintError("ERROR: Invoke function did not find ChainCode function: " + fn)
    return nil, errors.New(" --- INVOKE ERROR: Received unknown function invocation")
}


// Query is our entry point for queries
func (dcc *DecodedChainCode) Query(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    // Handle different functions
    if fn == "read" { // read a variable
        return dcc.read(stub, fn, args)
    } else if fn == "readAll" { // read all ledgers, returns the ids
        return dcc.readAll(stub, fn, args)
    } else if fn == "readAllOwners" { // read all owners and return full data for them.
        return dcc.readAllOwners(stub, fn, args)
    } else if fn == "readAllAssets" { // read all assets and return full data for them.
        return dcc.readAllAssets(stub, fn, args)
    } else if fn == "readAllTransactions" { // read all transactions and return full data for them.
        return dcc.readAllTransactions(stub, fn, args)
    }
    utils.PrintError("ERROR: Query function did not find ChainCode function: " + fn)
    return nil, errors.New(" --- QUERY ERROR: Received unknown function query")
}


// ============================================================================================================================


// Function that reads the bytes associated with a data-key and returns the byte-array.
func (dcc *DecodedChainCode) read(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 1 { // needs a data key to read.
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    dataKey := args[0]
    dataBytes, err := stub.GetState(dataKey)
    if dataBytes == nil { // deals with non existing data keys.
        err = errors.New("{\"Error\":\"State " + dataKey + " does not exist\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    if err != nil {
        err = errors.New("{\"Error\":\"Failed to get state for " + dataKey + "\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    } 
    utils.PrintSuccess("Read the ledger: " + dataKey)
    return dataBytes, nil
}


func (dcc *DecodedChainCode) readAll(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // get all owners - returns an array of strings.
    ownersLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[0], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAll - getDataArrayStrings", err)
        return nil, err
    }
    // get all assets
    assetsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[1], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAll - getDataArrayStrings", err)
        return nil, err
    }
    // get all transactions
    transactionsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[2], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAll - getDataArrayStrings", err)
        return nil, err
    }
    // get all transactions
    pendingTransactionsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[3], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAll - getDataArrayStrings", err)
        return nil, err
    }
    // Create a map of all the ledgers.
    m := map[string][]string{ 
        PRIMARYKEY[0]: ownersLedger, 
        PRIMARYKEY[1]: assetsLedger, 
        PRIMARYKEY[2]: transactionsLedger, 
        PRIMARYKEY[3]: pendingTransactionsLedger,
    }
    // Cast to JSON
    mStr, err := json.Marshal(m)
    if err != nil {
        utils.PrintErrorFull("readAll - Marshal", err)
        return nil, err
    }
    // Return as bytes.
    utils.PrintSuccess("Successfully read all the main ledgers ")
    out := []byte(string(mStr))
    return out, nil
}


func (dcc *DecodedChainCode) getDataArrayStrings(stub shim.ChaincodeStubInterface, dataKey string, args []string) ([]string, error) {
    var err error
    var empty []string
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"getDataArrayStrings\"}")
        utils.PrintErrorFull("", err)
        return empty, err
    }
    arrayBytes, err := stub.GetState(dataKey)
    if err != nil {
        utils.PrintErrorFull("getDataArrayStrings - GetState", err)
        return empty, err
    }
    var outputArray []string
    if err = json.Unmarshal(arrayBytes, &outputArray); err != nil {
        utils.PrintErrorFull("getDataArrayStrings - Unmarshal", err)
        return empty, err
    }
    return outputArray, nil
}


func (dcc *DecodedChainCode) saveStringToDataArray(stub shim.ChaincodeStubInterface, dataKey string, addString string, ledger []string) ([]byte, error) {
    var err error
    // Add the string to the array
    ledger = append(ledger, addString)
    if err = dcc.saveLedger(stub, dataKey, ledger); err != nil {
        utils.PrintErrorFull("saveStringToDataArray - saveLedger", err)
        return nil, err
    }
    return nil, nil
}


func (dcc *DecodedChainCode) saveLedger(stub shim.ChaincodeStubInterface, dataKey string, ledger []string) (error) {
    var err error
    // Marshall the ledger to bytes
    bytesToWrite, err := json.Marshal(&ledger)
    if err != nil {
        return err
    }
    // Save the array.
    if err = stub.PutState(dataKey, bytesToWrite); err != nil {
        return err
    }
    return nil
}


// ============================================================================================================================

