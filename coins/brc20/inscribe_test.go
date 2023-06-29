package brc20

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

func TestInscribe(t *testing.T) {
	network := &chaincfg.TestNet3Params

	commitTxPrevOutputList := make([]*PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "fcd1a1c33df653427e20159a799e6c1ba28421fd168fe353a54508c956fb382e",
		VOut:       1,
		Amount:     546,
		Address:    "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "fcd1a1c33df653427e20159a799e6c1ba28421fd168fe353a54508c956fb382e",
		VOut:       0,
		Amount:     252198,
		Address:    "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "fcd1a1c33df653427e20159a799e6c1ba28421fd168fe353a54508c956fb382e",
		VOut:       2,
		Amount:     100000,
		Address:    "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "fcd1a1c33df653427e20159a799e6c1ba28421fd168fe353a54508c956fb382e",
		VOut:       3,
		Amount:     796800,
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
}

func TestBrc30(t *testing.T) {
	network := &chaincfg.RegressionNetParams
	client := modeRpcClient()
	defer client.Shutdown()
	addr := "bcrt1qvd26a8c26d4mu5fzyh74pvcp9ykgutxt9fktqf"
	address, err := btcutil.DecodeAddress(addr, network)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}
	priv, err := client.DumpPrivKey(address)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}

	utxos, err := client.ListUnspentMinMaxAddresses(1, 100, []btcutil.Address{address})
	if err != nil {
		t.Fatal("get utxo error:", err.Error())
	}
	var prevout = PrevOutput{
		TxId:       "64d8369309e3c1be0010a8b19d892c394013a1d813f780559d715bee73645c42",
		VOut:       1,
		Amount:     4999947860,
		Address:    addr,
		PrivateKey: priv.String(),
	}
	for i, _ := range utxos {
		if utxos[i].Amount > 20 {
			prevout.Amount = int64(utxos[i].Amount * 100000000)
			prevout.TxId = utxos[i].TxID
			prevout.VOut = utxos[i].Vout
			break
		}
	}
	commitTxPrevOutputList := make([]*PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &prevout)

	inscriptionDataList := make([]InscriptionData, 0)
	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc-20","op":"deploy","tick":"lf06","max":"21000000","lim":"1000","dec":"3"}`),
	//	RevealAddr:  addr,
	//})
	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc-20","op":"mint","tick":"lf06","amt":"1000"}`),
	//	RevealAddr:  addr,
	//})
	//
	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"deploy","t":"pool","pid":"d2fcc4c3fc#01","stake":"lf06","earn":"abcd","erate":"1000","dec":"2","dmax":"1000000","total":"40000000","only":"1"}`),
	//	RevealAddr:  addr,
	//})
	//
	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"stake","pid":"d2fcc4c3fc#01","amt":"1000"}`),
	//	RevealAddr:  addr,
	//})

	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"unstake","pid":"d2fcc4c3fc#01","amt":"100"}`),
	//	RevealAddr:  addr,
	//})

	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"mint","tick":"abcd","tid":"d2fcc4c3fc","amt":"100"}`),
	//	RevealAddr:  addr,
	//})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc20-s","op":"transfer","tid":"d2fcc4c3fc","tick":"abcd","amt":"1"}`),
		RevealAddr:  addr,
	})

	request := &InscriptionRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          2,
		RevealFeeRate:          2,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          addr,
	}

	requestBytes, _ := json.Marshal(request)
	log.Println(string(requestBytes))

	tool, err := newInscriptionTool(network, request)
	if err != nil {
		t.Fatal("build inscribe error:", err.Error())
	}

	//send commit tx
	commitTXID, err := client.SendRawTransaction(tool.CommitTx, true)
	if err != nil {
		t.Fatal("send commit error:", err.Error(), "commit tx", *tool.CommitTx)
	}
	t.Log("Commit TXID:", commitTXID.String())
	genrateBlock(t, client, address)
	for i, _ := range tool.RevealTx {
		revealTXID, err := client.SendRawTransaction(tool.RevealTx[i], true)
		if err != nil {
			t.Fatal("send reveal error:", err.Error(), "commit tx", tool.CommitTx)
		}
		t.Log("Reveal TXID:", revealTXID.String())
		genrateBlock(t, client, address)
	}
}

func TestAutoBRC30(t *testing.T) {
	// brc20 -- a001 ,a002 ,a003 ,a004
	// brc30 --

	// c41b7cf376#01 brc20(b001) ->  brc30(b001) only
	// c41b7cf376#02 brc20(b002) -> brc30(b001) share
	// f2f838e203#01 brc20(b001) ->  brc30(b001) only
	// f2f838e203#02 brc20(b002) -> brc30(b001) share
	// c2dce1ef8#01 brc20(b002) -> brc30(b003) share
	// 7087ee2f6b#01 brc20(b002) -> brc30(b005) share fixed
	inscriptions := []string{
		//`{"p":"brc-20","op":"deploy","tick":"b001","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc-20","op":"mint","tick":"b001","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b002","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc-20","op":"mint","tick":"lf08","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b003","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc20-s","op":"mint","tick":"aaab","pid":"b9429047e5#01","amt":"100"}`,
		//`{"p":"brc-20","op":"mint","tick":"b003","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b004","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc-20","op":"mint","tick":"b004","amt":"1000"}`,
		//`{"p":"brc20-s","op":"deploy","t":"fixed","pid":"7087ee2f6b#01","stake":"b002","earn":"b005","erate":"1000","dec":"2","dmax":"1000000","total":"21000000","only":""}`,
		//`{"p":"brc20-s","op":"deposit","pid":"c2dce1ef8e#01","amt":"18446744073709551615"}`,
		//`{"p":"brc20-s","op":"deposit","pid":"f2f838e203#02","amt":"500"}`,
		//`{"p":"brc20-s","op":"withdraw","pid":"f2f838e203#02","amt":"1"}`,
		//`{"p":"brc20-s","op":"mint","tick":"aaab","tid":"833814e8ba","amt":"100"}`,
		//`{"p":"brc20-s","op":"transfer","tid":"833814e8ba","tick":"aaab","amt":"100"}`,
		//`{"p":"brc-20","op":"transfer","tick":"b002","amt":"100"}`,

		//`{"p":"brc-20","op":"deploy","tick":"b20a","max":"21000000","lim":"1000","dec":"18"}`,
		//`{"p":"brc-20","op":"mint","tick":"b20a","amt":"1000"}`,
		`{"p":"brc-20","op":"transfer","tick":"b20a","amt":"200"}`,

		//`{"p":"brc20-s","op":"deploy","t":"fixed","pid":"22f4b3f9fd#01","stake":"b20a","earn":"b30a","erate":"100","dec":"18","dmax":"1000000","total":"21000000","only":"true"}`,
		//`{"p":"brc20-s","op":"deposit","pid":"22f4b3f9fd#01","amt":"20"}`,
		//`{"p":"brc20-s","op":"mint","tick":"b30a","pid":"22f4b3f9fd#01","amt":"200000"}`,
		//`{"p":"brc20-s","op":"transfer","tid":"22f4b3f9fd","tick":"b30a","amt":"500"}`,

	}

	/* send inscribe transaction */
	//autoInscribe(t, "bcrt1qvd26a8c26d4mu5fzyh74pvcp9ykgutxt9fktqf", inscriptions)

	/* send transfer transaction */
	_ = inscriptions
	txID := "637cf4481d284de1cb407711566f2d1b53d64ccf63f9d03ee30e141f7483ad84" // tx id of inscribe transfer transaction
	fromAddr := "bcrt1qvd26a8c26d4mu5fzyh74pvcp9ykgutxt9fktqf"
	toAddr := "bcrt1qdk34rmzke023v04xxvpxz0fwauctsxk42ue2zj"
	autoTransfer(t, txID, fromAddr, toAddr)
}

func TestCaculateHash(t *testing.T) {
	addr := "bcrt1qvd26a8c26d4mu5fzyh74pvcp9ykgutxt9fktqf"
	hash := caculateTickID("b30a", 21000000, 18, addr, addr)
	t.Log(hex.EncodeToString(hash))
}

func autoInscribe(t *testing.T, addr string, inscriptions []string) {
	network := &chaincfg.RegressionNetParams
	client := modeRpcClient()
	defer client.Shutdown()

	address, err := btcutil.DecodeAddress(addr, network)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}
	priv, err := client.DumpPrivKey(address)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}

	utxos, err := client.ListUnspentMinMaxAddresses(1, 100, []btcutil.Address{address})
	if err != nil {
		t.Fatal("get utxo error:", err.Error())
	}
	var prevout = PrevOutput{
		TxId:       "64d8369309e3c1be0010a8b19d892c394013a1d813f780559d715bee73645c42",
		VOut:       1,
		Amount:     4999947860,
		Address:    addr,
		PrivateKey: priv.String(),
	}
	for i, _ := range utxos {
		if utxos[i].Amount > 20 {
			prevout.Amount = int64(utxos[i].Amount * 100000000)
			prevout.TxId = utxos[i].TxID
			prevout.VOut = utxos[i].Vout
			break
		}
	}
	commitTxPrevOutputList := make([]*PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &prevout)

	inscriptionDataList := make([]InscriptionData, 0)
	for i, _ := range inscriptions {
		inscriptionDataList = append(inscriptionDataList, InscriptionData{
			ContentType: "text/plain;charset=utf-8",
			Body:        []byte(inscriptions[i]),
			RevealAddr:  addr,
		})
	}

	request := &InscriptionRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          2,
		RevealFeeRate:          2,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          addr,
	}

	requestBytes, _ := json.Marshal(request)
	log.Println(string(requestBytes))

	tool, err := newInscriptionTool(network, request)
	if err != nil {
		t.Fatal("build inscribe error:", err.Error())
	}

	//send commit tx
	commitTXID, err := client.SendRawTransaction(tool.CommitTx, true)
	if err != nil {
		t.Fatal("send commit error:", err.Error(), "commit tx", *tool.CommitTx)
	}
	t.Log("Commit TXID:", commitTXID.String())
	genrateBlock(t, client, address)
	for i, _ := range tool.RevealTx {
		revealTXID, err := client.SendRawTransaction(tool.RevealTx[i], true)
		if err != nil {
			t.Fatal("send reveal error:", err.Error(), "commit tx", tool.CommitTx)
		}
		t.Log("Reveal TXID:", revealTXID.String())
		genrateBlock(t, client, address)
	}
}

func autoTransfer(t *testing.T, txID string, fromAddr string, toAddr string) {
	network := &chaincfg.RegressionNetParams
	client := modeRpcClient()
	defer client.Shutdown()

	fromAddress, err := btcutil.DecodeAddress(fromAddr, network)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}

	//toAddress, err := btcutil.DecodeAddress(toAddr, network)
	//if err != nil {
	//	t.Fatal("decode addr error:", err.Error())
	//}
	privateKey, err := client.DumpPrivKey(fromAddress)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}

	//utxos, err := client.ListUnspentMinMaxAddresses(1, 100, []btcutil.Address{address})
	utxos, err := client.ListUnspent()
	if err != nil {
		t.Fatal("get utxo error:", err.Error())
	}
	var prevOutput = PrevOutput{
		TxId:       txID,
		VOut:       0,
		Amount:     4999947860,
		Address:    fromAddr,
		PrivateKey: privateKey.String(),
	}
	txUnspent := false
	for i, _ := range utxos {
		if utxos[i].TxID == txID {
			prevOutput.Amount = int64(utxos[i].Amount * 100000000)
			prevOutput.TxId = utxos[i].TxID
			prevOutput.VOut = utxos[i].Vout
			txUnspent = true
			break
		}
	}
	if !txUnspent {
		t.Fatal("tx has been spent:", txID)
	}
	//inputs := []btcjson.TransactionInput{
	//	{
	//		Txid: txID,
	//		Vout: prevout.VOut,
	//	},
	//}
	//amounts := map[btcutil.Address]btcutil.Amount {
	//	toAddress: btcutil.Amount(prevout.Amount),
	//}
	//msgTx, err := client.CreateRawTransaction(inputs, amounts, nil)
	//if err != nil {
	//	t.Fatal("CreateRawTransaction error:", err.Error())
	//}
	//signedMsgTx, isSigned, err := client.SignRawTransaction(msgTx)
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)
	privateKeyList := []*btcec.PrivateKey{privateKey.PrivKey}
	tx := wire.NewMsgTx(DefaultTxVersion)
	changePkScript, err := AddrToPkScript(toAddr, network)
	if err != nil {
		t.Fatal("AddrToPkScript error:", err.Error())
	}

	txHash, err := chainhash.NewHashFromStr(prevOutput.TxId)
	if err != nil {
		t.Fatal("NewHashFromStr error:", err.Error())
	}
	outPoint := wire.NewOutPoint(txHash, prevOutput.VOut)
	pkScript, err := AddrToPkScript(prevOutput.Address, network)
	if err != nil {
		t.Fatal("AddrToPkScript error:", err.Error())
	}
	txOut := wire.NewTxOut(prevOutput.Amount, pkScript)
	prevOutFetcher.AddPrevOut(*outPoint, txOut)

	in := wire.NewTxIn(outPoint, nil, nil)
	in.Sequence = DefaultSequenceNum
	tx.AddTxIn(in)

	tx.AddTxOut(wire.NewTxOut(prevOutput.Amount, changePkScript))

	txForEstimate := wire.NewMsgTx(DefaultTxVersion)
	txForEstimate.TxIn = tx.TxIn
	txForEstimate.TxOut = tx.TxOut
	if err := sign(txForEstimate, privateKeyList, prevOutFetcher); err != nil {
		t.Fatal("SignRawTransaction error:", err.Error())
	}
	commitFeeRate := 2
	fee := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(txForEstimate))) * btcutil.Amount(commitFeeRate)
	totalSenderAmount := btcutil.Amount(prevOutput.Amount)
	changeAmount := totalSenderAmount - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			t.Fatal("insufficient balance")
		}
	}

	err = sign(tx, privateKeyList, prevOutFetcher)
	if err != nil {
		t.Fatal("SignRawTransaction error:", err.Error())
	}
	//if !isSigned {
	//	t.Fatal("SignRawTransaction failed")
	//}
	commitTXID, err := client.SendRawTransaction(tx, true)

	if err != nil {
		t.Fatal("send commit error:", err.Error(), "commit tx", tx)
	}
	t.Log("Commit TXID:", commitTXID.String())
	genrateBlock(t, client, fromAddress)
}

func genrateBlock(t *testing.T, client *rpcclient.Client, address btcutil.Address) {
	maxTri := int64(3)
	_, err := client.GenerateToAddress(1, address, &maxTri)
	if err != nil {
		t.Fatal("generate block failed", err.Error())
	}
}

func modeRpcClient() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "<ip>:<port>",
		User:         "<username>",
		Pass:         "<password>",
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}
	return client
}

func caculateTickID(tick string, supply int, dec int, from, to string) []byte {
	builder := strings.Builder{}
	builder.Write([]byte(tick))
	builder.Write([]byte(fmt.Sprintf("%d", supply)))
	builder.Write([]byte(fmt.Sprintf("%d", dec)))
	builder.Write([]byte(from))
	builder.Write([]byte(to))
	fmt.Println("builder", builder.String())
	hasher := sha256.New()
	hasher.Write([]byte(builder.String()))
	b := hasher.Sum(nil)
	return b
}
