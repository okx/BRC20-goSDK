# Brc20 sdk

This sdk provides functions of inscribing inscriptions and transferring inscriptions, including btc nft and brc20 tokens, etc.

## Inscribe inscription

In order to inscribe inscriptions, you can use the Inscribe function, just provide the input UTXOs and inscription data, please see the example for details.

### Example

```go
network := &chaincfg.TestNet3Params

commitTxPrevOutputList := make([]*PrevOutput, 0)
commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
    TxId:       "42323b006e7433c7010cb8025de3f018621c5bce882e35f97009d9313ce4b840",
    VOut:       0,
    Amount:     546,
    Address:    "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})
commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
    TxId:       "78d81df15795206560c5f4f49824a38deb0a63941c6d593ca12739b2d940c8cd",
    VOut:       2,
    Amount:     100000,
    Address:    "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})
commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
    TxId:       "a48400e329681c2524b939ffbafda3a5a0f5f6329015375298fc6703e3bed552",
    VOut:       0,
    Amount:     546,
    Address:    "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})
commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
    TxId:       "78d81df15795206560c5f4f49824a38deb0a63941c6d593ca12739b2d940c8cd",
    VOut:       3,
    Amount:     246500,
    Address:    "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})

inscriptionDataList := make([]InscriptionData, 0)
inscriptionDataList = append(inscriptionDataList, InscriptionData{
    ContentType: "text/plain;charset=utf-8",
    Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1000"}`),
    RevealAddr:  "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
})
inscriptionDataList = append(inscriptionDataList, InscriptionData{
    ContentType: "text/plain;charset=utf-8",
    Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1000"}`),
    RevealAddr:  "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
})
inscriptionDataList = append(inscriptionDataList, InscriptionData{
    ContentType: "text/plain;charset=utf-8",
    Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1000"}`),
    RevealAddr:  "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
})
inscriptionDataList = append(inscriptionDataList, InscriptionData{
    ContentType: "text/plain;charset=utf-8",
    Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1000"}`),
    RevealAddr:  "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
})

request := &InscriptionRequest{
    CommitTxPrevOutputList: commitTxPrevOutputList,
    CommitFeeRate:          2,
    RevealFeeRate:          2,
    RevealOutValue:         546,
    InscriptionDataList:    inscriptionDataList,
    ChangeAddress:          "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
}

requestBytes, _ := json.Marshal(request)
log.Println(string(requestBytes))

txs, _ := Inscribe(network, request)

txsBytes, _ := json.Marshal(txs)
t.Log(string(txsBytes))
```

### Parameters

Name | Type                | Description                                   | Notes
------------- |---------------------|-----------------------------------------------| -------------
**CommitTxPrevOutputList** | **[]\*PrevOutput**  | List of utxo's used by inscribed inscriptions |
**CommitFeeRate** | **int64**           | CommitTx fee rate                             |
**RevealFeeRate** | **int64**           | RevealTx fee rate                             |
**RevealOutValue** | **int64**           | RevealTx output amount                        | [optional] default 546
**InscriptionDataList** | **[]InscriptionData** | Inscription content list                      |
**ChangeAddress** | **string**          | Address to receive change                     |

**PrevOutput**

Name | Type       | Description                               | Notes
------------- |------------|-------------------------------------------| -------------
**TxId** | **string** | The transaction id where the utxo is located   |
**VOut** | **uint32** | Output offset                             |
**Amount** | **int64**  | Output amount                             |
**Address** | **string** | Output address                            |
**PrivateKey** | **string** | WIF encoded private key                   |

**InscriptionData**

Name | Type       | Description              | Notes
------------- |------------|--------------------------| -------------
**ContentType** | **string** | Inscription Content Type |
**Body** | **[]byte** | Inscription Data         |
**RevealAddr** | **string** | Inscription binding address           |

### Return value

Transactions to be broadcast.

## Transfer inscription

In order to transfer the inscription, you can use the Transfer function to transfer the inscription, which supports 4 types of address input, please see the example for details.
**Pay attention to the position change rules of sat, and ensure that the inscription bound to sat is transferred correctly.**

### Example

```go
network := &chaincfg.TestNet3Params

var inputs []*TxInput
inputs = append(inputs, &TxInput{
    TxId:       "25b9d08a26c8d47795301dd47a861cff0459d14f27fbd41cffaca17d9aa20f87",
    VOut:       uint32(0),
    Amount:     int64(249352),
    Address:    "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})
inputs = append(inputs, &TxInput{
    TxId:           "6d59aa50447c0d55e6f9535c3e56d7014b4ca8070ee57ce2199219790cfd5815",
    VOut:           uint32(0),
    Amount:         int64(499356),
    Address:        "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
    PrivateKey:     "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
    NonWitnessUtxo: "02000000010a6b13715c8effde51dac60d572358005a589cd80413a88e0912e4c6d275abbe010000006a473044022019e34aa16cf55eb9c7a8627f61bcd671525a3818a23ab8a78af13c35121ea3c8022055a5bfb3e8486f6e83707660f1fca3da06f140f449902a63900625f43fadf10501210357bbb2d4a9cb8a2357633f201b9c518c2795ded682b7913c6beef3fe23bd6d2fffffffff019c9e0700000000001976a9145c005c5532ce810ddf20f9d1d939631b47089ecd88ac00000000",
})
// inscription
inputs = append(inputs, &TxInput{
    TxId:       "46e3ce050474e6da80760a2a0b062836ff13e2a42962dc1c9b17b8f962444206",
    VOut:       uint32(0),
    Amount:     int64(546),
    Address:    "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})
inputs = append(inputs, &TxInput{
    TxId:       "d1696c10046ec8b2d938924f1923f1f2e1588095fbf3ea0f8cd640b51da51ba2",
    VOut:       uint32(0),
    Amount:     int64(400),
    Address:    "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
    PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
})

var outputs []*TxOutput
outputs = append(outputs, &TxOutput{
    Address: "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
    Amount:  int64(200000),
})
outputs = append(outputs, &TxOutput{
    Address: "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
    Amount:  int64(200000),
})
outputs = append(outputs, &TxOutput{
    Address: "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
    Amount:  int64(100000),
})
outputs = append(outputs, &TxOutput{
    Address: "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
    Amount:  int64(246500),
})
outputs = append(outputs, &TxOutput{
    Address: "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
    Amount:  int64(1000),
})
outputs = append(outputs, &TxOutput{
    Address: "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
    Amount:  int64(1000),
})

signedTx, err := Transfer(inputs, outputs, network)
if err != nil {
    t.Fatal(err)
}
t.Log(signedTx)
```

### Parameters

#### Inputs

Name | Type       | Description                               | Notes
------------- |------------|-------------------------------------------| -------------
**TxId** | **string** | The transaction id where the utxo is located   |
**VOut** | **uint32** | Output offset                             |
**Amount** | **int64**  | Output amount                             |
**Address** | **string** | Output address                            |
**PrivateKey** | **string** | WIF encoded private key                   |
**NonWitnessUtxo** | **string** | The transaction hex where utxo is located | [optional] p2pkh address required

#### Outputs

Name | Type        | Description                                   | Notes
------------- |-------------|-----------------------------------------------| -------------
**Address** | **string**  | Output address |
**Amount** | **int64**   | Output amount |

### Return value

Transactions to be broadcast.
