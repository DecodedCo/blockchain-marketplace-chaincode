# blockchain-golang-chaincode

The blockchain application in Golang

# Running the application

You need to add the hyperledger GO code to your local installation.

so in your GO/src/github.com folder:

```
mkdir hyperledger
cd hyperledger
git clone git@github.com:hyperledger/fabric.git
cd fabric
git fetch
git checkout v.06
```

Then return to this repo and run:

```
go build .
```

The run and register it with the blockchain using

```
CORE_CHAINCODE_ID_NAME=<applicationname> CORE_PEER_ADDRESS=0.0.0.0:7051 ./blockchain-golang-chaincode
```

```
CORE_CHAINCODE_ID_NAME=DecodedBlockChain CORE_PEER_ADDRESS=0.0.0.0:7051 ./blockchain-golang-chaincode
```

# Operations overview.

### Deploy

Now **deploy** the chaincode using:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "deploy",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"init", 
            "args": [] 
        } 
    },
    "id": 46664
}' "http://0.0.0.0:7050/chaincode"
```

**Note: This will be different in the future when using a cluster of peers since you cannot attach a process to all of them it seems**

### Invoke and Query OWNERS

To create a new owner we need to use the **invoke** method (this adds a `transaction` to the blockchain).

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "invoke",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"addOwnerString", 
            "args": [ "dcd", "Decoded", "100000", "description", "logo-url", "tag" ] 
        } 
    },
    "id": 2600
}' "http://0.0.0.0:7050/chaincode"
```

and create another one

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "invoke",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"addOwnerString", 
            "args": [ "bc", "BlockChain", "100000", "description", "logo-url", "tag" ] 
        } 
    },
    "id": 2600
}' "http://0.0.0.0:7050/chaincode"
```

Now we have two companies available. We can read the owners in three ways:

1. To now **query** a specific owner:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "query",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"read", 
            "args": [ "dcd" ] 
        } 
    },
    "id": 1337
}' "http://0.0.0.0:7050/chaincode"
```

To **query** all owners without the owners full information

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "query",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"read", 
            "args": [ "Owners" ] 
        } 
    },
    "id": 1337
}' "http://0.0.0.0:7050/chaincode"
```

To **query** all Owners with full information you can use the function `readAllOwners`

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "query",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"readAllOwners", 
            "args": [] 
        } 
    },
    "id": 1337
}' "http://0.0.0.0:7050/chaincode"
```

### Invoke and Query ASSETS

To add an asset:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "invoke",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"addAssetString", 
            "args": [ "appleId", "Apples", "dcd", "100", "13", "description", "logo", "approval", "approvalqty", "phonenr", "phonemsg", "tag" ] 
        } 
    },
    "id": 2600
}' "http://0.0.0.0:7050/chaincode"
```

The quantity and price are no optional

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "invoke",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"addAssetString", 
            "args": [ "bananaId", "Banana", "dcd", "100", "23", "description", "logo", "approval", "approvalqty", "phonenr", "phonemsg", "tag" ] 
        } 
    },
    "id": 2600
}' "http://0.0.0.0:7050/chaincode"
```

To read all the Assets you can use a similar function as for the Owners: `readAllAssets`

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "query",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"readAllAssets", 
            "args": [ ] 
        } 
    },
    "id": 0
}' "http://0.0.0.0:7050/chaincode"
```

### Invoke and Query TRANSACTIONS

There are two types of Transactions possible at this moment: straight-through (no approval needed) and pending (approval needed).

To transact an asset straight away without approval:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0",
    "method": "invoke",
    "params": {
        "type": 1,
        "chaincodeID": {
            "name": "DecodedBlockChain"
        },
        "ctorMsg": {
            "function": "transactAsset",
            "args": [
                "appleId", "dcd", "bc", "33", "12345", "FALSE"
            ]
        }
    },
    "id": 1
}' "http://0.0.0.0:7050/chaincode"
```

The input arguments are: `assetId`, `sellerId`, `buyerId`, `quantity`, `price`, `approvalRequired`.

You can check the transactions using

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "query",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"readAllTransactions", 
            "args": [] 
        } 
    },
    "id": 0
}' "http://0.0.0.0:7050/chaincode"
```

To create a transaction that needs to be approved use:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0",
    "method": "invoke",
    "params": {
        "type": 1,
        "chaincodeID": {
            "name": "DecodedBlockChain"
        },
        "ctorMsg": {
            "function": "transactAsset",
            "args": [
                "appleId", "dcd", "bc", "47", "8888", "TRUE"
            ]
        }
    },
    "id": 1
}' "http://0.0.0.0:7050/chaincode"
```

To approve a transaction

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0",
    "method": "invoke",
    "params": {
        "type": 1,
        "chaincodeID": {
            "name": "DecodedBlockChain"
        },
        "ctorMsg": {
            "function": "approveTransaction",
            "args": [
                "<TRANSACTION-ID>"
            ]
        }
    },
    "id": 1
}' "http://0.0.0.0:7050/chaincode"
```

To decline a transaction

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0",
    "method": "invoke",
    "params": {
        "type": 1,
        "chaincodeID": {
            "name": "DecodedBlockChain"
        },
        "ctorMsg": {
            "function": "declineTransaction",
            "args": [
                "<TRANSACTION-ID>"
            ]
        }
    },
    "id": 1
}' "http://0.0.0.0:7050/chaincode"
```

### Other

To change the validation status of an owner:

```
curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "jsonrpc": "2.0", 
    "method": "invoke",  
    "params": {
        "type":1, 
        "chaincodeID": {
            "name":"DecodedBlockChain"
        }, 
        "ctorMsg": { 
            "function":"changeOwnerValidationStatus", 
            "args": [ "bc" ] 
        } 
    },
    "id": 2600
}' "http://0.0.0.0:7050/chaincode"
```

