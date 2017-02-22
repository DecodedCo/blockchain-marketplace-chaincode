/*

DECODED HYPERLEDGER APPLICATION

DecodedChainCode functions:
- createOwner - private functions
- getOwner - private functions
- addOwner, addOwnerString, addOwnerJSON: adds an owner using a number of string arguments or a single JSON-encoded argument.
- assignAssetToOwner - private function
- readAllOwners
- changeOwnerValidationStatus

Owner functions:
- save
- addAsset
- removeAsset
- deleteAsset
- addTransaction
- approveBuyTransaction
- approveSellTransaction
- rollbackBuyTransaction
- verifyBalance
- isValidated
- toggleValidation

*/


package main


import (
    "encoding/json"
    "strconv"
    "errors"

    "github.com/hyperledger/fabric/core/chaincode/shim"

    utils "github.com/DecodedCo/blockchain-golang-chaincode/utils"
)


// ============================================================================================================================


type Owner struct {
    OwnerId         string      `json:"username"`
    Name            string      `json:"name"`
    Information     OwnerInfo   `json:"information"`    
    Balance         float64     `json:"balance"`
    Validated       bool        `json:"validated"`
    Tag             string      `json:"tag"`
    // 
    EscrowBalance   float64     `json:"escrowBalance"`
    //
    Assets          []string    `json:"assetIds"` // For now the asset name is the id.
    //
    Issued          []string    `json:"issuedIds"` // Assets that this company issued.
    //
    Transactions    []string    `json:"transactions"` // transaction ids
}


type OwnerInfo struct {
    Description     string      `json:"description"`
    Logo            string      `json:"logo"` // url to image
    Background      string      `json:"background"` // url to image
}


// ============================================================================================================================


func (o *Owner) save(stub shim.ChaincodeStubInterface) (error) {
    var err error
    ownerBytesToWrite, err := json.Marshal(&o)
    if err != nil {
        utils.PrintErrorFull("save - Marshal", err)
        return err
    }
    if err = stub.PutState(o.OwnerId, ownerBytesToWrite); err != nil {
        utils.PrintErrorFull("save - PutState", err)
        return err
    }
    return nil
} // end of o.save


func (o *Owner) addAsset(assetId string) {
    // If the owner already owns nothing changes, otherwise add to slice.
    if utils.IsElementInSlice(o.Assets, assetId) == false {
        o.Assets = append(o.Assets, assetId)
    }
} // end of o.addAsset


func (o *Owner) removeAsset(asset *Asset, quantity int) {
    // If it is the full holding, delete from the slice.
    ownsQuantity := asset.OwnedBy[o.OwnerId].Quantity
    if ownsQuantity == quantity {
        o.deleteAsset(asset.Id)
    }
} // end of o.removeAsset


func (o *Owner) deleteAsset(assetId string) {
    o.Assets = utils.DeleteElementFromSlice(o.Assets, assetId)
} // end of o.deleteAsset


func (o *Owner) addTransaction(transactionId string) {
    o.Transactions = append(o.Transactions, transactionId)
} // end of o.addTransaction


func (o *Owner) approveBuyTransaction(assetId string, amount float64) {
    o.addAsset(assetId)
    o.EscrowBalance = o.EscrowBalance - amount
} // end of o.approveBuyTransaction


func (o *Owner) approveSellTransaction(asset *Asset, amount float64, quantity int) {
    if asset.OwnedBy[o.OwnerId].EscrowQty == quantity && asset.OwnedBy[o.OwnerId].Quantity == 0 {
        o.deleteAsset(asset.Id)
    }
    o.Balance = o.Balance + amount
} // end of o.approveSellTransaction


func (o *Owner) rollbackBuyTransaction(amount float64) {
    o.EscrowBalance = o.EscrowBalance - amount
    o.Balance = o.Balance + amount
} // end of o.rollbackBuyTransaction


func (o *Owner) verifyBalance(amount float64) (error) {
    var err error
    if o.Balance < amount {
        err = errors.New("Insufficient balance.")
        return err
    }
    return nil
} // end of o.verifyBalance


func (o *Owner) isValidated(fn string) (error) {
    var err error
    if o.Validated == false {
        err = errors.New("{\"Error\":\"Owner " + o.OwnerId + " is not validated\", \"Function\":\"" + fn + "\"}")
        return err
    }
    return nil
} // end of o.isValidated


func (o *Owner) toggleValidation() {
    if o.Validated == true {
        o.Validated = false
    } else {
        o.Validated = true
    }
} // end of o.toggleValidation


// ============================================================================================================================


func (dcc *DecodedChainCode) createOwner(args []string) (Owner, error) {
    var err error
    var information OwnerInfo
    var emptyArgs []string
    var owner Owner // We need to have an empty owner ready to return in case of an error.
    isValidated := false // Initialise as true for now.
    if len(args) != 6 { // OwnerId, fullname, balance, description, logo-url, tag
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"createOwner\"}")
        utils.PrintErrorFull("", err)
        return owner, err
    }
    balance, err := strconv.ParseFloat(args[2], 64)
    if err != nil {
        utils.PrintErrorFull("createOwner - Atoi", err)
        return owner, err
    }
    information = OwnerInfo{ Description: args[3], Logo: args[4], Background: "" }
    owner = Owner{ 
        OwnerId: args[0], 
        Name: args[1], 
        Information: information,
        Tag: args[5],
        Balance: balance, 
        EscrowBalance: 0.0,
        Assets: emptyArgs, 
        Issued: emptyArgs, 
        Validated: isValidated,
        Transactions: emptyArgs,
    }
    // Done.
    utils.PrintSuccess("Created the new owner: " + args[1])
    return owner, nil
} // end of dcc.createOwner


func (dcc *DecodedChainCode) getOwner(stub shim.ChaincodeStubInterface, args []string) (Owner, error) {
    var owner Owner // We need to have an empty owner ready to return in case of an error.
    var err error
    if len(args) != 1 { // Only needs an owner id.
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"getOwner\"}")
        utils.PrintErrorFull("", err)
        return owner, err
    }
    ownerId := args[0]
    ownerBytes, err := stub.GetState(ownerId)
    if ownerBytes == nil {
        err = errors.New("{\"Error\":\"State " + ownerId + " does not exist\", \"Function\":\"getOwner\"}")
        utils.PrintErrorFull("", err)
        return owner, err
    }
    if err != nil {
        utils.PrintErrorFull("getOwner - GetState", err)
        return owner, err
    }
    if err = json.Unmarshal(ownerBytes, &owner); err != nil {
        utils.PrintErrorFull("getOwner - Unmarshal", err)
        return owner, err
    }
    utils.PrintSuccess("Successfully retrieved the owner: " + owner.Name)
    return owner, nil
} // end of dcc.getOwner


func (dcc *DecodedChainCode) addOwner(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    if len(args) != 6 { // OwnerId, fullname, balance, description, logo-url, tag
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // The OwnerId needs to be unique. Check if the owner does not already exist.
    ownerId := args[0]
    // Get all the owners that are currently in the system.
    ownersLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[0], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("addOwner - getDataArrayStrings", err)
        return nil, err
    }
    // Check if the ownerId exists in the current ledger of owners.
    ownerExists := utils.IsElementInSlice(ownersLedger, ownerId)
    if ownerExists == false {
        // Create a new owner
        newOwner, err := dcc.createOwner(args)
        if err != nil {
            utils.PrintErrorFull("addOwner - createOwner", err)
            return nil, err
        }
        // Save new owner
        if err = newOwner.save(stub); err != nil {
            utils.PrintErrorFull("addOwner - save", err)
            return nil, err
        }
        // Add owner to the list.
        _, err = dcc.saveStringToDataArray(stub, PRIMARYKEY[0], ownerId, ownersLedger)
        if err != nil {
            utils.PrintErrorFull("addOwner - saveStringToDataArray", err)
            return nil, err
        }
        // Done!
        utils.PrintSuccess("Successfully added a new owner: " + newOwner.Name)
        return nil, nil
    } else {
        err = errors.New("Owner `" + ownerId + "` already exists.")
        utils.PrintErrorFull("addOwner", err)
        return nil, err
    }
    // Redundancy.
    return nil, nil
} // end of dcc.addOwner


// Wrapper. Pass everything on to `addOwner`
func (dcc *DecodedChainCode) addOwnerString(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    return dcc.addOwner(stub, fn, args)
} // end of dcc.addOwnerString


func (dcc *DecodedChainCode) updateOwner(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 3 { // OwnerId, Description, Logo
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Load the current data
    ownerId := args[0]
    owner, err := dcc.getOwner(stub, []string{ ownerId })
    if err != nil {
        utils.PrintErrorFull("updateOwner - getOwner", err)
        return nil, err
    }
    // Update the fields
    owner.Information.Description = args[1]
    owner.Information.Logo = args[2]
    // Save the new owner.
    if err = owner.save(stub); err != nil {
        utils.PrintErrorFull("updateOwner - save", err)
        return nil, err
    }
    utils.PrintSuccess("Successfully updated owner: " + owner.Name)
    return nil, nil
} // end of dcc.updateOwner


func (dcc *DecodedChainCode) assignAssetToOwner(stub shim.ChaincodeStubInterface, fn string, assetId string, owner Owner, isIssuer bool) ([]byte, error) {
    var err error
    // Add the asset to the list of assets this owner owns. 
    // It does not hold information about how many, that is stored in the asset.
    owner.Assets = append(owner.Assets, assetId)
    if isIssuer {
        owner.Issued = append(owner.Issued, assetId)
    }
    // Save the owner.
    if err = owner.save(stub); err != nil {
        utils.PrintErrorFull("assignAssetToOwner - save", err)
        return nil, err
    }
    utils.PrintSuccess("Assigned asset `" + assetId + "` to owner `" + owner.OwnerId + "`")
    return nil, nil
} // end of dcc.assignAssetToOwner


func (dcc *DecodedChainCode) readAllOwners(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get all owners - returns an slice of strings - ownerIds
    ownersLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[0], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAllOwners - getDataArrayStrings", err)
        return nil, err
    }
    if len(ownersLedger) > 0 {
        // Initialise an empty slice for the output
        var fullOwnersLedger []Owner
        // Iterate over all owners and return the owner object.
        for _, ownerId := range ownersLedger {
            thisOwner, err := dcc.getOwner(stub, []string{ ownerId })
            if err != nil {
                utils.PrintErrorFull("readAllOwners - getOwner", err)
                return nil, err
            }
            fullOwnersLedger = append(fullOwnersLedger, thisOwner)
        }
        // This gives us an slice with owners. Translate to bytes and return
        fullOwnersLedgerBytes, err := json.Marshal(&fullOwnersLedger)
        if err != nil {
            utils.PrintErrorFull("readAllOwners - Marshal", err)
            return nil, err
        }
        utils.PrintSuccess("Retrieved full information for all owners.")
        return fullOwnersLedgerBytes, nil
    } else {
        return nil, nil
    }
    return nil, nil // redundancy
} // end of dcc.readAllOwners


// Function just flips the validation status of the owner.
func (dcc *DecodedChainCode) changeOwnerValidationStatus(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 1 { // Only needs the ownerId.
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    ownerId := args[0]
    owner, err := dcc.getOwner(stub, []string{ ownerId })
    if err != nil {
        utils.PrintErrorFull("changeOwnerValidationStatus - getOwner", err)
        return nil, err
    }
    // Flip the validation status of the owner.
    owner.toggleValidation()
    if err = owner.save(stub); err != nil {
        utils.PrintErrorFull("changeOwnerValidationStatus - save", err)
        return nil, err
    }
    utils.PrintSuccess("Changed the owner validation status for " + ownerId)
    return nil, nil
} // end of dcc.changeOwnerValidationStatus


// ============================================================================================================================

