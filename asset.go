/*

DECODED HYPERLEDGER APPLICATION

DecodedChainCode functions:
- createAsset - private functions
- getAsset - private functions
- addAssetString
- readAllAssets

Asset functions
- save
- addOwner
- subtractOwner
- removeOwner
- deleteOwner
- escrowOwner
- rollbackTransaction
- verifyHoldings
- verifyPrice

*/


package main


import (
    "encoding/json"
    "errors"
    "strconv"
    "time"
    "strings"

    "github.com/hyperledger/fabric/core/chaincode/shim"

    utils "github.com/DecodedCo/blockchain-marketplace-chaincode/utils"
)


// ============================================================================================================================


type Asset struct {
    // Specifics around the asset.
    Id          string                  `json:"assetId"`
    Name        string                  `json:"name"`
    Information AssetInfo               `json:"information"`
    Tag         string                  `json:"tag"`
    //
    Quantity    int                     `json:"quantity"` // available quantity
    Price       float64                 `json:"price"`
    //
    Issuer      string                  `json:"issuer"`
    IssuedTS    int64                   `json:"issued"`
    IssuedQty   int                     `json:"issuedQty"`
    // It is always initialised with the issuer owning the whole Quantity.
    Owners      []string                `json:"ownerIds"`
    OwnedBy     map[string]OwnedBy      `json:"ownership"`
    //
    Triggers    Trigger                 `json:"triggers"`
    Contract    API                     `json:"apiTrigger"`
}


type AssetInfo struct {
    Description string                  `json:"description"`
    Logo        string                  `json:"logo"` // url to image
}


type OwnedBy struct {
    OwnerId     string                  `json:"ownerId"`
    Quantity    int                     `json:"quantity"`
    EscrowQty   int                     `json:"escrowQty"`
}


// Smart contract features. Texting is deprecated.


type Trigger struct {
    Approval    bool                    `json:"approval"`
    ApprovalQty int                     `json:"approvalQty"`
}


type API struct {
    URL         string                  `json:"url"`
    Keys        []string                `json:"keys"`
    Condition   string                  `json:"isBelow"`
    Value       string                  `json:"value"`
    Discount    float64                 `json:"discount"`
}


// ============================================================================================================================


func (a *Asset) save(stub shim.ChaincodeStubInterface) (error) {
    var err error
    assetBytesToWrite, err := json.Marshal(&a)
    if err != nil {
        utils.PrintErrorFull("save - Marshal", err)
        return err
    }
    if err = stub.PutState(a.Id, assetBytesToWrite); err != nil {
        utils.PrintErrorFull("save - PutState", err)
        return err
    }
    return nil
} // end of a.save


func (a *Asset) addOwner(ownerId string, quantity int) {
    var ownedBy OwnedBy
    if utils.IsElementInSlice(a.Owners, ownerId) { // Check if this owner already owns the asset
        ownedBy = a.OwnedBy[ownerId]
        ownedBy.Quantity = ownedBy.Quantity + quantity
    } else { // new asset for this ownerId
        a.Owners = append(a.Owners, ownerId)
        ownedBy = OwnedBy{ OwnerId: ownerId, Quantity: quantity, EscrowQty: 0 }
    }
    a.OwnedBy[ownerId] = ownedBy
} // end of a.addOwner


func (a *Asset) subtractOwner(ownerId string, quantity int, isPending bool) {
    // Owner stays in the asset.Owners. Needs to update its holdings though.
    ownedBy := a.OwnedBy[ownerId]
    if isPending {
        ownedBy.EscrowQty = ownedBy.EscrowQty - quantity
    } else {
        ownedBy.Quantity = ownedBy.Quantity - quantity
    }
    a.OwnedBy[ownerId] = ownedBy
} // end of a.subtractOwner


func (a *Asset) removeOwner(ownerId string, quantity int, isPending bool) {
    ownedBy := a.OwnedBy[ownerId]
    if isPending {
        if ownedBy.Quantity == 0 && ownedBy.EscrowQty == quantity { // If this owner has just sold its entire quantity, delete.
            a.deleteOwner(ownerId)
        } else {
            a.subtractOwner(ownerId, quantity, isPending) // Owner still holds some of this asset in some form
        }
    } else {
        if ownedBy.Quantity == quantity && ownedBy.EscrowQty == 0 { // If this owner has just sold its entire quantity, delete.
            a.deleteOwner(ownerId)
        } else {
            a.subtractOwner(ownerId, quantity, isPending) // Owner still holds some of this asset in some form
        }
    }
} // end of a.removeOwner


func (a *Asset) deleteOwner(ownerId string) {
    a.Owners = utils.DeleteElementFromSlice(a.Owners, ownerId)
    delete(a.OwnedBy, ownerId)
} // end of a.deleteOwner


func (a *Asset) escrowOwner(ownerId string, quantity int) {
    // Leave the owner in the owners slice, Update the OwnedBy from quantity to escrow
    ownedBy := a.OwnedBy[ownerId]
    ownedBy.Quantity = ownedBy.Quantity - quantity
    ownedBy.EscrowQty = ownedBy.EscrowQty + quantity
    a.OwnedBy[ownerId] = ownedBy
} // end of a.escrowOwner


func (a *Asset) rollbackTransaction(ownerId string, quantity int) {
    a.escrowOwner(ownerId, quantity) // We can call it with a negative number
    // Also update the Quantity available if the owner is the issuer.
    if ownerId == a.Issuer {
        a.Quantity = a.Quantity + quantity
    }
} // end of a.rollbackTransaction


func (a *Asset) verifyHoldings(ownerId string, quantity int) (error) {
    var err error
    if a.OwnedBy[ownerId].Quantity < quantity {
        err = errors.New("Insufficient quantity of ownership.")
        return err
    }
    return nil
} // end of a.verifyHoldings


func (a *Asset) verifyPrice(price float64) (error) {
    var err error
    if a.Price != price {
        err = errors.New("Price does not agree with specifications.")
        return err
    }
    return nil
} // end of a.verifyPrice


// ============================================================================================================================


func (dcc *DecodedChainCode) createAsset(args []string) (Asset, error) {
    var err error
    var asset Asset // We need to have an empty asset ready to return in case of an error.
    var information AssetInfo
    var ownedBy OwnedBy // Initialised empty.
    var trigger Trigger
    var api API
    ownedByMap := make(map[string]OwnedBy)
    var quantity, approvalQty int
    var price float64
    // Check inputs. Only requires an asset id, asset name, ownerId, quantity, price, description, logo
    // smart contract: approval, approvalqty
    // 10 = tag
    if len(args) != 10 { 
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"createAsset\"}")
        utils.PrintErrorFull("", err)
        return asset, err
    }
    quantity, err = strconv.Atoi(args[3])
    if err != nil {
        utils.PrintErrorFull("createAsset - Atoi", err)
        return asset, err
    }
    price, err = strconv.ParseFloat(args[4], 64)
    if err != nil {
        utils.PrintErrorFull("createAsset - ParseFloat", err)
        return asset, err
    }
    approval, err := strconv.ParseBool(args[7])
    if err != nil {
        utils.PrintErrorFull("createAsset - ParseBool", err)
        return asset, err
    }
    approvalQty, err = strconv.Atoi(args[8])
    if err != nil {
        utils.PrintErrorFull("createAsset - Atoi", err)
        return asset, err
    }
    timestamp := time.Now().Unix()
    // Populate the structs
    information = AssetInfo{ Description: args[5], Logo: args[6]}
    ownedBy = OwnedBy{ OwnerId: args[2], Quantity: quantity, EscrowQty: 0 }
    ownedByMap[args[2]] = ownedBy
    trigger = Trigger{ Approval: approval, ApprovalQty: approvalQty }
    asset = Asset{
        Id: args[0],
        Name: args[1],
        Information: information,
        Tag: args[9],
        Quantity: quantity, 
        Price: price,
        Issuer: args[2], 
        IssuedTS: timestamp, 
        IssuedQty: quantity, 
        Owners: []string{ args[2] },
        OwnedBy: ownedByMap,
        Triggers: trigger,
        Contract: api, // initialised empty currently.
    }
    // Done
    utils.PrintSuccess("Successfully created a new asset `" + args[0] + "` for owner `" + args[2] + "`")
    return asset, nil
} // end of dcc.createAsset


func (dcc *DecodedChainCode) getAsset(stub shim.ChaincodeStubInterface, args []string) (Asset, error) {
    var asset Asset
    var err error
    if len(args) != 1 { // Only needs an asset id.
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"getAsset\"}")
        utils.PrintErrorFull("", err)
        return asset, err
    }
    assetId := args[0]
    assetBytes, err := stub.GetState(assetId)
    if assetBytes == nil {
        err = errors.New("{\"Error\":\"State " + assetId + " does not exist\", \"Function\":\"getAsset\"}")
        utils.PrintErrorFull("getAsset - GetState", err)
        return asset, err
    }
    if err != nil {
        utils.PrintErrorFull("getAsset - GetState", err)
        return asset, err
    }
    if err = json.Unmarshal(assetBytes, &asset); err != nil {
        utils.PrintErrorFull("getAsset - Unmarshal", err)
        return asset, err
    }
    return asset, nil
} // end of dcc.getAsset


func (dcc *DecodedChainCode) addAssetString(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var empty []string
    if len(args) != 10 { // Id, Name, issuerId, Quantity, Price, description, logo, approval, approvalQty, tag
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    assetId := args[0]
    issuerId := args[2]
    // Check if the issuer exists.
    issuer, err := dcc.getOwner(stub, []string{ issuerId })
    if err != nil {
        utils.PrintErrorFull("addAssetString - getOwner", err)
        return nil, err
    }
    // Check if the issuer is validated! - if not the owner cannot create new assets.
    if issuer.Validated == false {
        err = errors.New("{\"Error\":\"Owner " + issuerId + " is not validated\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get a list of all assets.
    assetsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[1], empty)
    if err != nil {
        utils.PrintErrorFull("addAssetString - getDataArrayStrings", err)
        return nil, err
    }
    // Check if the assetId exists in the current ledger of assets.
    assetExists := utils.IsElementInSlice(assetsLedger, assetId)
    if assetExists == false {
        // Create a new asset. This is initialised without the issuer associated.
        newAsset, err := dcc.createAsset(args) // Args has the assetId, assetName, ownerName and quantity
        if err != nil {
            utils.PrintErrorFull("addAssetString - createAsset", err)
            return nil, err
        }
        // Save new asset
        if err = newAsset.save(stub); err != nil {
            utils.PrintErrorFull("addAssetString - save", err)
            return nil, err
        }
        // Add asset to the list.
        _, err = dcc.saveStringToDataArray(stub, PRIMARYKEY[1], assetId, assetsLedger)
        if err != nil {
            utils.PrintErrorFull("addAssetString - saveStringToDataArray", err)
            return nil, err
        }
        // ----------------------------------------------
        // Assign the asset to the owner.
        _, err = dcc.assignAssetToOwner(stub, "assignAssetToOwner", assetId, issuer, true)
        if err != nil {
            utils.PrintErrorFull("addAssetString - assignAssetToOwner", err)
            return nil, err
        }
        utils.PrintSuccess("Successfully added a new asset: " + assetId + " to the owner: " + issuerId)
        return nil, nil
    } else {
        err = errors.New("Asset `" + assetId + "` already exists.")
        utils.PrintErrorFull("addAssetString", err)
        return nil, err
    }
    return nil, nil // Redundancy.
} // end of dcc.addAssetString


func (dcc *DecodedChainCode) updateAsset(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    if len(args) != 9 { // assetId, Name, Description, Logo, | url, keys, condition, value, discount
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Load the current data
    assetId := args[0]
    asset, err := dcc.getAsset(stub, []string{ assetId })
    if err != nil {
        utils.PrintErrorFull("updateAsset - getAsset", err)
        return nil, err
    }
    // Deal with the string discount
    discount, err := strconv.ParseFloat(args[8], 64)
    if err != nil {
        utils.PrintErrorFull("updateAsset - ParseFloat", err)
        return nil, err
    }
    // Update the fields.
    asset.Name = args[1]
    asset.Information.Description = args[2]
    asset.Information.Logo = args[3]
    asset.Contract.URL = args[4]
    asset.Contract.Keys = strings.Split(args[5], ",")
    asset.Contract.Condition = args[6]
    asset.Contract.Value = args[7]
    asset.Contract.Discount = discount
    // Save the new asset.
    if err = asset.save(stub); err != nil {
        utils.PrintErrorFull("updateAsset - save", err)
        return nil, err
    }
    return nil, nil
} // end of dcc.updateAsset


// Function to read all available assets and their information.
// It has an optional parameter that can filter out assets owned by a specific owner.
func (dcc *DecodedChainCode) readAllAssets(stub shim.ChaincodeStubInterface, fn string, args []string) ([]byte, error) {
    var err error
    var emptyArgs []string
    if len(args) != 0 {
        err = errors.New("{\"Error\":\"Incorrect number of arguments\", \"Function\":\"" + fn + "\"}")
        utils.PrintErrorFull("", err)
        return nil, err
    }
    // Get all assets - returns an slice of strings.
    assetsLedger, err := dcc.getDataArrayStrings(stub, PRIMARYKEY[1], emptyArgs)
    if err != nil {
        utils.PrintErrorFull("readAllAssets - getDataArrayStrings", err)
        return nil, err
    }
    if len(assetsLedger) > 0 {
        // Initialise an empty slice for the output
        var fullAssetsLedger []Asset
        // Iterate over all assets and return the asset object.
        for _, assetId := range assetsLedger {
            thisAsset, err := dcc.getAsset(stub, []string{ assetId })
            if err != nil {
                utils.PrintErrorFull("readAllAssets - getAsset", err)
                return nil, err
            }
            fullAssetsLedger = append(fullAssetsLedger, thisAsset)
        }
        // This gives us an slice with assets. Translate to bytes and return
        fullAssetsLedgerBytes, err := json.Marshal(&fullAssetsLedger)
        if err != nil {
            utils.PrintErrorFull("readAllAssets - Marshal", err)
            return nil, err
        }
        utils.PrintSuccess("Retrieved full information for all assets.")
        return fullAssetsLedgerBytes, nil
    } else {
        return nil, nil
    }
    return nil, nil // redundancy
} // end of dcc.readAllAssets


// ============================================================================================================================

