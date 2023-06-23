package brc20

import (
	"encoding/json"
	"github.com/btcsuite/btcd/btcutil"
	"log"
	"testing"

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
	//	Body:        []byte(`{"p":"brc20-s","op":"deploy","t":"pool","pid":"ddc8c2547a#01","stake":"lf06","earn":"abcd","erate":"1000","dec":"2","dmax":"1000000","total":"40000000","only":"1"}`),
	//	RevealAddr:  addr,
	//})
	//
	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"stake","pid":"ddc8c2547a#01","amt":"1000"}`),
	//	RevealAddr:  addr,
	//})

	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"unstake","pid":"ddc8c2547a#01","amt":"100"}`),
	//	RevealAddr:  addr,
	//})

	//inscriptionDataList = append(inscriptionDataList, InscriptionData{
	//	ContentType: "text/plain;charset=utf-8",
	//	Body:        []byte(`{"p":"brc20-s","op":"mint","tick":"abcd","tid":"ddc8c2547a","amt":"100"}`),
	//	RevealAddr:  addr,
	//})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc20-s","op":"transfer","tid":"ddc8c2547a","tick":"abcd","amt":"1"}`),
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
	inscriptions := []string{
		`{"p":"brc-20","op":"deploy","tick":"lf06","max":"21000000","lim":"1000","dec":"3"}`,
		`{"p":"brc-20","op":"mint","tick":"lf06","amt":"1000"}`,
		`{"p":"brc20-s","op":"deploy","t":"pool","pid":"ddc8c2547a#01","stake":"lf06","earn":"abcd","erate":"1000","dec":"2","dmax":"1000000","total":"40000000","only":"1"}`,
		`{"p":"brc20-s","op":"stake","pid":"ddc8c2547a#01","amt":"1000"}`,
		`{"p":"brc20-s","op":"unstake","pid":"ddc8c2547a#01","amt":"100"}`,
		`{"p":"brc20-s","op":"mint","tick":"abcd","tid":"ddc8c2547a","amt":"100"}`,
		`{"p":"brc20-s","op":"transfer","tid":"ddc8c2547a","tick":"abcd","amt":"1"}`,
	}
	autoInscribe(t, "bcrt1qvd26a8c26d4mu5fzyh74pvcp9ykgutxt9fktqf", inscriptions)
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
	genrateBlock(t, client, address)
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
		Host:         "18.167.77.79:18443",
		User:         "bitcoinrpc",
		Pass:         "bitcoinrpc",
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}
	return client
}
