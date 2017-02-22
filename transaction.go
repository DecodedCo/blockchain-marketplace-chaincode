/*

DECODED HYPERLEDGER APPLICATION

Types of available transactions:
    - Straight-through-processed: effective immediately.
    - Pending-approval: transaction processed but put in a queue for validation.
                        funds are taken from buyer and put into its escrow, asset is taken from owner and put in its escrow.
                        once the transaction is approved everything is finalised.

DecodedChainCode functions:
- createTransaction
- getTransaction
- transactAsset
- approveTransaction
- declineTransaction
- readAllTransactions

Transaction functions:
- save
- approve
- rollback
- removeFromPendingLedger

*/


package main


import (
    "encoding/json"
    "errors"
    "strconv"
    "time"

    "github.com/hyperledger/fabric/core/chaincode/shim"

    utils "github.com/DecodedCo/blockchain-golang-chaincode/utils"
)



// ============================================================================================================================


type Transaction struct {
    Id              string      `json:"transactionId"`
    AssetId         string      `json:"assetId"`
    // Counterparties
    SellerId        string      `json:"sellerId"`
    BuyerId         string      `json:"buyerId"`
    // Specifics
    Quantity        int         `json:"quantity"`
    Price           float64     `json:"price"`
    Discount        float64     `json:"discount"`
    // Metadata
    Created         int64       `json:"createdAt"`
    // Status
    Status          string      `json:"status"`
    // API related.
    APIFixing       string      `json:"apifixing"`
}


// ============================================================================================================================


func (tx *Transaction) save(stub shim.ChaincodeStubInterface) (error) {
    var err error
    transactionBytesToWrite, err := json.Marshal(&tx)
    if err != nil {
        return err
    }
    err = stub.PutState(tx.Id, transactionBytesToWrite)
    if err != nil {
        return err
    }
    return nil
}


func (tx *Transaction) approve(stub shim.ChaincodeStubInterface, dcc *DecodedChainCode) (error) {
    var err error
    var buyer, seller Owner
    var asset Asset
    if tx.Status != "Pending" {
        err = errors.New("{\"Error\":\"Trying to roll back a non-pending transaction (" + tx.Id + ")\"}")
        return err
    }
    // Get the structs needed
    buyer, err = dcc.getOwner(stub, []string{ tx.BuyerId })
    if err != nil {
        return err
    }
    seller, err = dcc.getOwner(stub, []string{ tx.SellerId })
    if err != nil {
        return err
    }
    asset, err = dcc.getAsset(stub, []string{ tx.AssetId })
    if err != nil {
        return err
    }
    // Process the approval.
    // 1. Buyer: Take the buyer escrow money and add the asset if needed.
    buyer.approveBuyTransaction(tx.AssetId, tx.Price * float64(tx.Quantity))
    err = buyer.save(stub)
    if err != nil {
        return err
    }
    // 2. Seller, give him the funds and take the asset off his ledger.
    seller.approveSellTransaction(&asset, tx.Price * float64(tx.Quantity), tx.Quantity)
    err = seller.save(stub)
    if err != nil {
        return err
    }
    // 3. Asset: Update the escrow for the seller. Remove asset if needed.
    //           Update/add everything for the buyer.
    asset.addOwner(tx.BuyerId, tx.Quantity)
    asset.removeOwner(tx.SellerId, tx.Quantity, true)
    err = asset.save(stub)
    if err != nil {
        return err
    }
    // Update the status
    tx.Status = "Approved"
    return nil
}


func (tx *Transaction) rollback(stub shim.ChaincodeStubInterface, dcc *DecodedChainCode) (error) {
    var err error
    var buyer Owner
    var asset Asset
    if tx.Status != "Pending" {
        err = errors.New("{\"Error\":\"Trying to roll back a non-pending transaction (" + tx.Id + ")\"}")
        return err
    }
    // Rollback the buyer escrow money
    buyer, err = dcc.getOwner(stub, []string{ tx.BuyerId })
    if err != nil {
        return err
    }
    buyer.rollbackBuyTransaction(tx.Price * float64(tx.Quantity))
    err = buyer.save(stub)
    if err != nil {
        return err
    }
    // Rollback the asset-seller escrow quantity
    asset, err = dcc.getAsset(stub, []string{ tx.AssetId })
    if err != nil {
        return err
    }
    asset.rollbackTransaction(tx.SellerId, tx.Quantity)
    err = asset.save(stub)
    if err != nil {
        return err
    }
    // Update the status
    tx.Status = "Declined"
    return nil
}


func (tx *Transaction) removeFromPendingLedger(stub shim.ChaincodeStubInterface, dcc *DecodedChainCode) (error) {
    var err error
    var pendingTransactionsLedger, emptyArgs []string
    // Get the pending ledger
    pendingTransactionsLedger, err = dcc.getDataArrayStrings(stub, PRIMARYKEY[3], emptyArgs)
    // Remove transaction id
    pendingTransactionsLedger = utils.DeleteElementFromSlice(pendingTransactionsLedger, tx.Id)
    // Save.
    err = dcc.saveLedger(stub, PRIMARYKEY[3], pendingTransactionsLedger)
    if err != nil {
        return err
    }
    return nil
}


// ============================================================================================================================


func (dcc *DecodedChainCode) createTransaction(quantity int, price float64, args []string) (Transaction, error) {
    var err error
    var transaction Transaction
    if len(args) != 3 { // assetId, sellerId, buyerId
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"createTransaction\"}")
        utils.PrintErrorFull("", err)
        return transaction, err
    }
    assetId := args[0]
    sellerId := args[1]
    buyerId := args[2]
    // Create the transaction
    timestamp := time.Now().Unix()
    transactionId := utils.HashSHA256( assetId + "-" + sellerId + "-" + buyerId + "-" + strconv.Itoa(int(timestamp)) )
    transaction = Transaction{
        Id: transactionId,
        AssetId: assetId, 
        SellerId: sellerId, 
        BuyerId: buyerId, 
        Quantity: quantity, 
        Price: price, 
        Discount: 0.0,
        Created: timestamp,
        Status: "Pending",
        APIFixing: "",
    }
    return transaction, err
}


func (dcc *DecodedChainCode) getTransaction(stub shim.ChaincodeStubInterface, args []string) (Transaction, error) {
    var transaction Transaction
    var err error
    if len(args) != 1 { // Only needs an transaction id.
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"getTransaction\"}")
        utils.PrintErrorFull("", err)
        return transaction, err
    }
    transactionId := args[0]
    transactionBytes, err := stub.GetState(transactionId)
    if transactionBytes == nil {
        err = errors.New("{\"Error\":\"State " + transactionId + " does not exist\", \"Function\":\"getTransaction\"}")
        utils.PrintErrorFull("", err)
        return transaction, err
    }
    if err != nil {
        utils.PrintErrorFull("getTransaction - GetState", err)
        return transaction, err
    }
    if err = json.Unmarshal(transactionBytes, &transaction); err != nil {
        utils.PrintErrorFull("getTransaction - Unmarshal", err)
        return transaction, err
    }
    return transaction, nil
}


func (dcc *DecodedChainCode) transactAsset(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    var transaction Transaction
    // Check for the appropriate number of inputs: assetName, fromName, toName, quantity, forAmount, approvalNeeded
    if len(args) != 6 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // ----------------------------------------------
    // Handle the inputs.
    assetId := args[0]
    sellerId := args[1]
    buyerId := args[2]
    quantity, err := strconv.Atoi(args[3]) // Convert string to int.
    if err != nil {
        utils.PrintErrorFull("transactAsset - Atoi", err)
        return nil, err
    }
    price, err := strconv.ParseFloat(args[4], 64) // Convert string to float64.
    if err != nil {
        utils.PrintErrorFull("transactAsset - ParseFloat", err)
        return nil, err
    }
    forAmount := price * float64(quantity);
    approvalRequired := args[5]
    // ----------------------------------------------
    // Check the existence of the asset and owners.
    asset, err := dcc.getAsset(stub, []string{ assetId })
    if err != nil {
        utils.PrintErrorFull("transactAsset - getAsset", err)
        return nil, err
    }
    seller, err := dcc.getOwner(stub, []string{ sellerId })
    if err != nil {
        utils.PrintErrorFull("transactAsset - getOwner", err)
        return nil, err
    }
    buyer, err := dcc.getOwner(stub, []string{ buyerId })
    if err != nil {
        utils.PrintErrorFull("transactAsset - getOwner", err)
        return nil, err
    }
    // Trigger for approval...
    if asset.Triggers.Approval == true && quantity > asset.Triggers.ApprovalQty {
        approvalRequired = "TRUE"
    }
    // ----------------------------------------------
    // Check the requirements for trading.
    // 1. Check if both owners are validated to trade.
    if err = seller.isValidated(fn); err != nil {
        utils.PrintErrorFull("transactAsset - isValidated", err)
        return nil, err
    }
    if err = buyer.isValidated(fn); err != nil {
        utils.PrintErrorFull("transactAsset - isValidated", err)
        return nil, err
    }
    // 2. Check if the current owner actually owns the asset.
    checkAsset := utils.IsElementInSlice(asset.Owners, sellerId)
    checkOwner := utils.IsElementInSlice(seller.Assets, asset.Id)
    if checkAsset == false || checkOwner == false {
        err = errors.New("Ownership issues.")
        utils.PrintErrorFull("transactAsset", err)
        return nil, err
    }
    // 3. Check the balance is enough to pay the forAmount.
    if err = buyer.verifyBalance(forAmount); err != nil {
        utils.PrintErrorFull("transactAsset - verifyBalance", err)
        return nil, err
    }
    // 4. Check if the owner owns enough of the asset.
    if err = asset.verifyHoldings(sellerId, quantity); err != nil {
        utils.PrintErrorFull("transactAsset - verifyHoldings", err)
        return nil, err
    }
    // 5. Check if the asset price is right
    if err = asset.verifyPrice(price); err != nil {
        utils.PrintErrorFull("transactAsset - verifyPrice", err)
        return nil, err
    }
    // ----------------------------------------------
    // Everything checks out.
    transaction, err = dcc.createTransaction(quantity, price, []string{ assetId, sellerId, buyerId })
    if err != nil {
        utils.PrintErrorFull("transactAsset - createTransaction", err)
        return nil, err
    }
    // Load the current transactionsLedger
    transactionsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[2], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("transactAsset - getDataArrayStrings", err)
        return nil, err
    }
    // ----------------------------------------------
    // Get the API fixing if needed...
    if asset.Contract.URL != "" {
        fixing := utils.GetEndpoint(asset.Contract.URL , asset.Contract.Keys)
        // Save the fixing in the transaction...
        transaction.APIFixing = fixing
        // Parse the value and fixing
        val, err := strconv.ParseFloat(asset.Contract.Value, 64)
        if err != nil {
            utils.PrintErrorFull("transactAsset - ParseFloat", err)
            return nil, err
        }
        fix, err := strconv.ParseFloat(fixing, 64)
        if err != nil {
            utils.PrintErrorFull("transactAsset - ParseFloat", err)
            return nil, err
        }
        // Check the condition.
        switch cond := asset.Contract.Condition; cond {
            case "less":
                if fix < val {
                    transaction.Discount = asset.Contract.Discount
                }
            case "more":
                if fix > val {
                    transaction.Discount = asset.Contract.Discount
                }
            case "equal":
                if fix == val {
                    transaction.Discount = asset.Contract.Discount
                }
            default:
                utils.PrintError("Critical error... smart contract code failing...")
        }
        // Update the price.
        price = price * (1.0 - transaction.Discount / 100.0)
        transaction.Price = price
    }
    // ----------------------------------------------
    // Some things have to happen regardless of the transaction requires approval
    buyer.Balance = buyer.Balance - forAmount
    buyer.addTransaction(transaction.Id)
    seller.addTransaction(transaction.Id)
    if sellerId == asset.Issuer {
        asset.Quantity = asset.Quantity - quantity
    }
    // ----------------------------------------------
    if approvalRequired == "TRUE" { // Process the pending transaction
        // Escrow the funds
        buyer.EscrowBalance = buyer.EscrowBalance + forAmount
        // Update the asset. Change the ownership to escrow for the quantity.
        asset.escrowOwner(sellerId, quantity)
        // Save in pending transactions
        pendingTransactionsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[3], emptyArgs)
        if err != nil {
            utils.PrintErrorFull("transactAsset - getDataArrayStrings", err)
            return nil, err
        }
        _, err = dcc.saveStringToDataArray(stub, PRIMARYKEY[3], transaction.Id, pendingTransactionsLedger)
        if err != nil {
            utils.PrintErrorFull("transactAsset - saveStringToDataArray", err)
            return nil, err
        }
        transaction.Status = "Pending"
    } else { // Process full transaction
        seller.removeAsset(&asset, quantity)
        seller.Balance = seller.Balance + forAmount
        buyer.addAsset(assetId)
        asset.addOwner(buyerId, quantity)
        asset.removeOwner(sellerId, quantity, false)
        // Update the transaction. from pending to accepted.
        transaction.Status = "Validated"
    }
    // ----------------------------------------------
    // Save the owners, asset, transaction and the general transaction ledger.
    if err = buyer.save(stub); err != nil {
        utils.PrintErrorFull("transactAsset - save", err)
        return nil, err
    }
    if err = seller.save(stub); err != nil {
        utils.PrintErrorFull("transactAsset - save", err)
        return nil, err
    }
    if err = asset.save(stub); err != nil {
        utils.PrintErrorFull("transactAsset - save", err)
        return nil, err
    }
    if err = transaction.save(stub); err != nil {
        utils.PrintErrorFull("transactAsset - save", err)
        return nil, err
    }
    _, err = dcc.saveStringToDataArray(stub, PRIMARYKEY[2], transaction.Id, transactionsLedger)
    if err != nil {
        utils.PrintErrorFull("transactAsset - saveStringToDataArray", err)
        return nil, err
    }
    // ----------------------------------------------
    utils.PrintSuccess("Transacted asset `" + assetId + "` from owner `" + sellerId + "` to owner `" + buyerId + "`")
    return nil, nil
}


// Approve pending transaction
func (dcc *DecodedChainCode) approveTransaction(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 1 { // Only needs a transactionId
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get the transaction
    transaction, err := dcc.getTransaction(stub, []string{ args[0] })
    if err != nil {
        utils.PrintErrorFull("approveTransaction - getTransaction", err)
        return nil, err
    }
    // Finalise the escrow for buyer, seller, and asset.
    if err = transaction.approve(stub, dcc); err != nil {
        utils.PrintErrorFull("approveTransaction - approve", err)
        return nil, err
    }
    // Update status, save transaction
    if err = transaction.save(stub); err != nil {
        utils.PrintErrorFull("approveTransaction - save", err)
        return nil, err
    }
    // Remove from pending ledger.
    if err = transaction.removeFromPendingLedger(stub, dcc); err != nil {
        utils.PrintErrorFull("approveTransaction - removeFromPendingLedger", err)
        return nil, err
    }
    utils.PrintSuccess("Approved transaction (" + transaction.Id + ") of asset `" + transaction.AssetId + "` from owner `" + transaction.SellerId + "` to owner `" + transaction.BuyerId + "`")
    return nil, nil
}


// Decline a pending transaction
func (dcc *DecodedChainCode) declineTransaction(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 1 { // Only needs a transactionId
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get the transaction
    transaction, err := dcc.getTransaction(stub, []string{ args[0] })
    if err != nil {
        utils.PrintErrorFull("declineTransaction - getTransaction", err)
        return nil, err
    }
    // Rollback the full transaction
    if err = transaction.rollback(stub, dcc); err != nil {
        utils.PrintErrorFull("declineTransaction - rollback", err)
        return nil, err
    }
    // Save transaction
    if err = transaction.save(stub); err != nil {
        utils.PrintErrorFull("declineTransaction - save", err)
        return nil, err
    }
    // Remove from pending ledger.
    if err = transaction.removeFromPendingLedger(stub, dcc); err != nil {
        utils.PrintErrorFull("declineTransaction - removeFromPendingLedger", err)
        return nil, err
    }
    // ----------------------------------------------
    utils.PrintSuccess("Declined transaction (" + transaction.Id + ") of asset `" + transaction.AssetId + "` from owner `" + transaction.SellerId + "` to owner `" + transaction.BuyerId + "`")
    return nil, nil
}


func (dcc *DecodedChainCode) readAllTransactions(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get all transactions - returns an slice of strings - transactionIds
    transactionsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[2], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAllTransactions - getDataArrayStrings", err)
        return nil, err
    }
    if len(transactionsLedger) > 0 {
        // Initialise an empty slice for the output
        var fullTransactionsLedger []Transaction
        // Iterate over all transactions and return the transaction object.
        for _, transactionId := range transactionsLedger {
            thisTransaction, err := dcc.getTransaction(stub, []string{ transactionId })
            if err != nil {
                utils.PrintErrorFull("readAllTransactions - getTransaction", err)
                return nil, err
            }
            fullTransactionsLedger = append(fullTransactionsLedger, thisTransaction)
        }
        // This gives us an slice with transactions. Translate to bytes and return
        fullTransactionsLedgerBytes, err := json.Marshal(&fullTransactionsLedger)
        if err != nil {
            utils.PrintErrorFull("readAllTransactions - Marshal", err)
            return nil, err
        }
        utils.PrintSuccess("Retrieved full information for all transactions.")
        return fullTransactionsLedgerBytes, nil
    } else {
        return nil, nil
    }
    return nil, nil // redundancy
}


// ============================================================================================================================

