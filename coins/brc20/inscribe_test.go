package brc20

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/joho/godotenv"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	addr := os.Getenv("Address")
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
		// `{"p":"brc-20","op":"deploy","tick":"ord1","max":"21000","lim":"1000","dec":"18"}`,
		// `{"p":"brc-20","op":"mint","tick":"ord1","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b002","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc-20","op":"mint","tick":"b002","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b003","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc20-s","op":"mint","tick":"aaab","pid":"b9429047e5#01","amt":"100"}`,
		//`{"p":"brc-20","op":"mint","tick":"b003","amt":"1000"}`,
		//`{"p":"brc-20","op":"deploy","tick":"b004","max":"21000000","lim":"1000","dec":"3"}`,
		//`{"p":"brc-20","op":"mint","tick":"b004","amt":"1000"}`,

		//`{"p":"brc20-s","op":"deploy","t":"fixed","pid":"c969bcd202#01","stake":"ord1","earn":"abce","erate":"1000","dec":"18","dmax":"21000","total":"21000","only":"1"}`,
		//`{"p":"brc20-s","op":"deposit","pid":"c969bcd202#01","amt":"18446744073709551615"}`,
		//`{"p":"brc20-s","op":"deposit","pid":"c969bcd202#01","amt":"1000"}`,
		//`{"p":"brc20-s","op":"withdraw","pid":"f2f838e203#02","amt":"1"}`,
		//`{"p":"brc20-s","op":"mint","tick":"abce","pid":"c969bcd202#01","amt":"100"}`,
		//`{"p":"brc20-s","op":"transfer","tid":"833814e8ba","tick":"aaab","amt":"100"}`,
		//`{"p":"brc-20","op":"transfer","tick":"b002","amt":"100"}`,
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	taddr := os.Getenv("Address")
	autoInscribe(t, taddr, inscriptions)
}

func TestCalculateHash(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	addr := os.Getenv("Address")
	name := os.Getenv("name")
	dec := os.Getenv("decimal")
	fmt.Println(dec)
	decimals, err := strconv.Atoi(dec)
	if err != nil {
		fmt.Println("Error during conversion")
		return
	}
	tmpTotal := os.Getenv("total")
	total, err := strconv.Atoi(tmpTotal)
	if err != nil {
		fmt.Println("Error during conversion")
		return
	}
	hash := calculateTickID(name, total, decimals, addr, addr)
	t.Log("TID:", hex.EncodeToString(hash)[0:10])
}

func TestGenBlock(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	addr := os.Getenv("Address")
	network := &chaincfg.RegressionNetParams
	client := modeRpcClient()
	defer client.Shutdown()
	address, err := btcutil.DecodeAddress(addr, network)
	if err != nil {
		t.Fatal("decode addr error:", err.Error())
	}
	genrateBlock(t, client, address)
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

	utxos, err := client.ListUnspentMinMaxAddresses(1, 100000, []btcutil.Address{address})
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
	isfoundUtxo := false
	for i, _ := range utxos {
		if utxos[i].Amount > 20 {
			x := new(big.Int)
			x.SetString(fmt.Sprintf("%.0f", utxos[i].Amount*1e8), 10)
			prevout.Amount = x.Int64()
			prevout.TxId = utxos[i].TxID
			prevout.VOut = utxos[i].Vout
			isfoundUtxo = true
			break
		}
	}
	if !isfoundUtxo {
		t.Fatal("not found uxto.amount > 20 of addr:", address.EncodeAddress(), "\n please change addr")
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

func genrateBlock(t *testing.T, client *rpcclient.Client, address btcutil.Address) {
	maxTri := int64(1)
	blockHash, err := client.GenerateToAddress(1, address, &maxTri)
	if err != nil {
		t.Fatal("generate block failed", err.Error())
	}
	for i := 0; i < len(blockHash); i++ {
		t.Log(blockHash[i])
	}
}

func modeRpcClient() *rpcclient.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	Host := os.Getenv("Host")
	User := os.Getenv("User")
	Pass := os.Getenv("Pass")

	connCfg := &rpcclient.ConnConfig{
		Host:         Host,
		User:         User,
		Pass:         Pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}
	return client
}

func calculateTickID(tick string, supply int, dec int, from, to string) []byte {
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
