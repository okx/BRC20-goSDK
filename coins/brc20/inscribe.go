package brc20

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	DefaultTxVersion      = 2
	DefaultSequenceNum    = mempool.MaxRBFSequence
	DefaultRevealOutValue = int64(546)
	MaxStandardTxWeight   = blockchain.MaxBlockWeight / 10
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type InscriptionData struct {
	ContentType string `json:"contentType"`
	Body        []byte `json:"body"`
	RevealAddr  string `json:"revealAddr"`
}

type PrevOutput struct {
	TxId       string `json:"txId"`
	VOut       uint32 `json:"vOut"`
	Amount     int64  `json:"amount"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type InscriptionRequest struct {
	CommitTxPrevOutputList []*PrevOutput     `json:"commitTxPrevOutputList"`
	CommitFeeRate          int64             `json:"commitFeeRate"`
	RevealFeeRate          int64             `json:"revealFeeRate"`
	InscriptionDataList    []InscriptionData `json:"inscriptionDataList"`
	RevealOutValue         int64             `json:"revealOutValue"`
	ChangeAddress          string            `json:"changeAddress"`
}

type InscribeTxs struct {
	CommitTx     string   `json:"commitTx"`
	RevealTxs    []string `json:"revealTxs"`
	CommitTxFee  int64    `json:"commitTxFee"`
	RevealTxFees []int64  `json:"revealTxFees"`
}

type inscriptionTxCtxData struct {
	PrivateKey              *btcec.PrivateKey
	InscriptionScript       []byte
	CommitTxAddressPkScript []byte
	ControlBlockWitness     []byte
	RevealTxPrevOutput      *wire.TxOut
}

type InscriptionTool struct {
	Network                   *chaincfg.Params
	CommitTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	CommitTxPrivateKeyList    []*btcec.PrivateKey
	InscriptionTxCtxDataList  []*inscriptionTxCtxData
	RevealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	CommitTxPrevOutputList    []*PrevOutput
	RevealTx                  []*wire.MsgTx
	CommitTx                  *wire.MsgTx
	MustCommitTxFee           int64
	MustRevealTxFees          []int64
}

func Inscribe(network *chaincfg.Params, request *InscriptionRequest) (*InscribeTxs, error) {
	tool, err := newInscriptionTool(network, request)
	if err != nil && errors.Is(err, ErrInsufficientBalance) {
		return &InscribeTxs{
			CommitTx:     "",
			RevealTxs:    []string{},
			CommitTxFee:  tool.MustCommitTxFee,
			RevealTxFees: tool.MustRevealTxFees,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	commitTx, err := tool.getCommitTxHex()
	if err != nil {
		return nil, err
	}
	revealTxs, err := tool.getRevealTxHexList()
	if err != nil {
		return nil, err
	}

	commitTxFee, revealTxFees := tool.calculateFee()

	return &InscribeTxs{
		CommitTx:     commitTx,
		RevealTxs:    revealTxs,
		CommitTxFee:  commitTxFee,
		RevealTxFees: revealTxFees,
	}, nil
}

func newInscriptionTool(network *chaincfg.Params, request *InscriptionRequest) (*InscriptionTool, error) {
	var commitTxPrivateKeyList []*btcec.PrivateKey
	for _, prevOutput := range request.CommitTxPrevOutputList {
		privateKeyWif, err := btcutil.DecodeWIF(prevOutput.PrivateKey)
		if err != nil {
			return nil, err
		}
		commitTxPrivateKeyList = append(commitTxPrivateKeyList, privateKeyWif.PrivKey)
	}
	tool := &InscriptionTool{
		Network:                   network,
		CommitTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		CommitTxPrivateKeyList:    commitTxPrivateKeyList,
		InscriptionTxCtxDataList:  make([]*inscriptionTxCtxData, len(request.InscriptionDataList)),
		RevealTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		CommitTxPrevOutputList:    request.CommitTxPrevOutputList,
	}
	if err := tool.initTool(network, request); err != nil {
		return tool, err
	}
	return tool, nil
}

func (tool *InscriptionTool) initTool(network *chaincfg.Params, request *InscriptionRequest) error {
	destinations := make([]string, len(request.InscriptionDataList))
	revealOutValue := DefaultRevealOutValue
	if request.RevealOutValue > 0 {
		revealOutValue = request.RevealOutValue
	}
	for i := 0; i < len(request.InscriptionDataList); i++ {
		inscriptionTxCtxData, err := createInscriptionTxCtxData(network, request, i)
		if err != nil {
			return err
		}
		tool.InscriptionTxCtxDataList[i] = inscriptionTxCtxData
		destinations[i] = request.InscriptionDataList[i].RevealAddr
	}
	totalRevealPrevOutputValue, err := tool.buildEmptyRevealTx(destinations, revealOutValue, request.RevealFeeRate)
	if err != nil {
		return err
	}
	err = tool.buildCommitTx(request.CommitTxPrevOutputList, request.ChangeAddress, totalRevealPrevOutputValue, request.CommitFeeRate)
	if err != nil {
		return err
	}
	err = tool.signCommitTx()
	if err != nil {
		return errors.New("sign commit tx error")
	}
	err = tool.completeRevealTx()
	if err != nil {
		return err
	}
	return nil
}

func createInscriptionTxCtxData(network *chaincfg.Params, inscriptionRequest *InscriptionRequest, indexOfInscriptionDataList int) (*inscriptionTxCtxData, error) {
	// use commitTx first input privateKey
	privateKeyWif, err := btcutil.DecodeWIF(inscriptionRequest.CommitTxPrevOutputList[0].PrivateKey)
	if err != nil {
		return nil, err
	}
	privateKey := privateKeyWif.PrivKey

	inscriptionBuilder := txscript.NewScriptBuilder().
		AddData(schnorr.SerializePubKey(privateKey.PubKey())).
		AddOp(txscript.OP_CHECKSIG).
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte("ord")).
		AddOp(txscript.OP_DATA_1).
		AddOp(txscript.OP_DATA_1).
		AddData([]byte(inscriptionRequest.InscriptionDataList[indexOfInscriptionDataList].ContentType)).
		AddOp(txscript.OP_0)

	maxChunkSize := 520
	bodySize := len(inscriptionRequest.InscriptionDataList[indexOfInscriptionDataList].Body)
	for i := 0; i < bodySize; i += maxChunkSize {
		end := i + maxChunkSize
		if end > bodySize {
			end = bodySize
		}
		// to skip txscript.MaxScriptSize 10000
		inscriptionBuilder.AddFullData(inscriptionRequest.InscriptionDataList[indexOfInscriptionDataList].Body[i:end])
	}
	inscriptionScript, err := inscriptionBuilder.Script()
	if err != nil {
		return nil, err
	}
	// to skip txscript.MaxScriptSize 10000
	inscriptionScript = append(inscriptionScript, txscript.OP_ENDIF)

	proof := &txscript.TapscriptProof{
		TapLeaf:  txscript.NewBaseTapLeaf(schnorr.SerializePubKey(privateKey.PubKey())),
		RootNode: txscript.NewBaseTapLeaf(inscriptionScript),
	}

	controlBlock := proof.ToControlBlock(privateKey.PubKey())
	controlBlockWitness, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	tapHash := proof.RootNode.TapHash()
	commitTxAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootOutputKey(privateKey.PubKey(), tapHash[:])), network)
	if err != nil {
		return nil, err
	}
	commitTxAddressPkScript, err := txscript.PayToAddrScript(commitTxAddress)
	if err != nil {
		return nil, err
	}

	return &inscriptionTxCtxData{
		PrivateKey:              privateKey,
		InscriptionScript:       inscriptionScript,
		CommitTxAddressPkScript: commitTxAddressPkScript,
		ControlBlockWitness:     controlBlockWitness,
	}, nil
}

func (tool *InscriptionTool) buildEmptyRevealTx(destination []string, revealOutValue, revealFeeRate int64) (int64, error) {
	addTxInTxOutIntoRevealTx := func(tx *wire.MsgTx, index int) error {
		in := wire.NewTxIn(&wire.OutPoint{Index: uint32(index)}, nil, nil)
		in.Sequence = DefaultSequenceNum
		tx.AddTxIn(in)
		scriptPubKey, err := AddrToPkScript(destination[index], tool.Network)
		if err != nil {
			return err
		}
		out := wire.NewTxOut(revealOutValue, scriptPubKey)
		tx.AddTxOut(out)
		return nil
	}

	totalPrevOutputValue := int64(0)
	total := len(tool.InscriptionTxCtxDataList)
	revealTx := make([]*wire.MsgTx, total)
	mustRevealTxFees := make([]int64, total)
	for i := 0; i < total; i++ {
		tx := wire.NewMsgTx(DefaultTxVersion)
		if err := addTxInTxOutIntoRevealTx(tx, i); err != nil {
			return 0, err
		}
		prevOutputValue := revealOutValue + int64(tx.SerializeSize())*revealFeeRate
		emptySignature := make([]byte, 64)
		emptyControlBlockWitness := make([]byte, 33)
		fee := (int64(wire.TxWitness{
			emptySignature,
			tool.InscriptionTxCtxDataList[i].InscriptionScript,
			emptyControlBlockWitness,
		}.SerializeSize()+2+3) / 4) * revealFeeRate // +2 encoding bytes , +3 for rounding up, /4 divide by weight
		prevOutputValue += fee
		tool.InscriptionTxCtxDataList[i].RevealTxPrevOutput = &wire.TxOut{
			PkScript: tool.InscriptionTxCtxDataList[i].CommitTxAddressPkScript,
			Value:    prevOutputValue,
		}
		totalPrevOutputValue += prevOutputValue
		revealTx[i] = tx
		mustRevealTxFees[i] = int64(tx.SerializeSize())*revealFeeRate + fee
	}
	tool.RevealTx = revealTx
	tool.MustRevealTxFees = mustRevealTxFees

	return totalPrevOutputValue, nil
}

func (tool *InscriptionTool) buildCommitTx(commitTxPrevOutputList []*PrevOutput, changeAddress string, totalRevealPrevOutputValue, commitFeeRate int64) error {
	totalSenderAmount := btcutil.Amount(0)
	tx := wire.NewMsgTx(DefaultTxVersion)
	changePkScript, err := AddrToPkScript(changeAddress, tool.Network)
	if err != nil {
		return err
	}
	for _, prevOutput := range commitTxPrevOutputList {
		txHash, err := chainhash.NewHashFromStr(prevOutput.TxId)
		if err != nil {
			return err
		}
		outPoint := wire.NewOutPoint(txHash, prevOutput.VOut)
		pkScript, err := AddrToPkScript(prevOutput.Address, tool.Network)
		if err != nil {
			return err
		}
		txOut := wire.NewTxOut(prevOutput.Amount, pkScript)
		tool.CommitTxPrevOutputFetcher.AddPrevOut(*outPoint, txOut)

		in := wire.NewTxIn(outPoint, nil, nil)
		in.Sequence = DefaultSequenceNum
		tx.AddTxIn(in)

		totalSenderAmount += btcutil.Amount(prevOutput.Amount)
	}
	for i := range tool.InscriptionTxCtxDataList {
		tx.AddTxOut(tool.InscriptionTxCtxDataList[i].RevealTxPrevOutput)
	}

	tx.AddTxOut(wire.NewTxOut(0, changePkScript))

	txForEstimate := wire.NewMsgTx(DefaultTxVersion)
	txForEstimate.TxIn = tx.TxIn
	txForEstimate.TxOut = tx.TxOut
	if err := sign(txForEstimate, tool.CommitTxPrivateKeyList, tool.CommitTxPrevOutputFetcher); err != nil {
		return err
	}

	fee := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(txForEstimate))) * btcutil.Amount(commitFeeRate)
	changeAmount := totalSenderAmount - btcutil.Amount(totalRevealPrevOutputValue) - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			txForEstimate.TxOut = txForEstimate.TxOut[:len(txForEstimate.TxOut)-1]
			feeWithoutChange := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(txForEstimate))) * btcutil.Amount(commitFeeRate)
			if totalSenderAmount-btcutil.Amount(totalRevealPrevOutputValue)-feeWithoutChange < 0 {
				tool.MustCommitTxFee = int64(fee)
				return ErrInsufficientBalance
			}
		}
	}
	tool.CommitTx = tx
	return nil
}

func (tool *InscriptionTool) completeRevealTx() error {
	for i := range tool.InscriptionTxCtxDataList {
		tool.RevealTxPrevOutputFetcher.AddPrevOut(
			wire.OutPoint{
				Hash:  tool.CommitTx.TxHash(),
				Index: uint32(i),
			},
			tool.InscriptionTxCtxDataList[i].RevealTxPrevOutput,
		)
		tool.RevealTx[i].TxIn[0].PreviousOutPoint.Hash = tool.CommitTx.TxHash()
	}
	for i := range tool.InscriptionTxCtxDataList {
		revealTx := tool.RevealTx[i]
		witnessArray, err := txscript.CalcTapscriptSignaturehash(
			txscript.NewTxSigHashes(revealTx, tool.RevealTxPrevOutputFetcher),
			txscript.SigHashDefault, revealTx, 0, tool.RevealTxPrevOutputFetcher,
			txscript.NewBaseTapLeaf(tool.InscriptionTxCtxDataList[i].InscriptionScript),
		)
		if err != nil {
			return err
		}
		signature, err := schnorr.Sign(tool.InscriptionTxCtxDataList[i].PrivateKey, witnessArray)
		if err != nil {
			return err
		}
		witness := wire.TxWitness{
			signature.Serialize(),
			tool.InscriptionTxCtxDataList[i].InscriptionScript,
			tool.InscriptionTxCtxDataList[i].ControlBlockWitness,
		}
		tool.RevealTx[i].TxIn[0].Witness = witness
	}
	// check tx max tx wight
	for i, tx := range tool.RevealTx {
		revealWeight := blockchain.GetTransactionWeight(btcutil.NewTx(tx))
		if revealWeight > MaxStandardTxWeight {
			return fmt.Errorf("reveal(index %d) transaction weight greater than %d (MAX_STANDARD_TX_WEIGHT): %d", i, MaxStandardTxWeight, revealWeight)
		}
	}
	return nil
}

func (tool *InscriptionTool) signCommitTx() error {
	return sign(tool.CommitTx, tool.CommitTxPrivateKeyList, tool.CommitTxPrevOutputFetcher)
}

func sign(tx *wire.MsgTx, privateKeys []*btcec.PrivateKey, prevOutFetcher *txscript.MultiPrevOutFetcher) error {
	for i, in := range tx.TxIn {
		prevOut := prevOutFetcher.FetchPrevOutput(in.PreviousOutPoint)
		txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		privKey := privateKeys[i]
		if txscript.IsPayToTaproot(prevOut.PkScript) {
			witness, err := txscript.TaprootWitnessSignature(tx, txSigHashes, i, prevOut.Value, prevOut.PkScript, txscript.SigHashDefault, privKey)
			if err != nil {
				return err
			}
			in.Witness = witness
		} else if txscript.IsPayToPubKeyHash(prevOut.PkScript) {
			sigScript, err := txscript.SignatureScript(tx, i, prevOut.PkScript, txscript.SigHashAll, privKey, true)
			if err != nil {
				return err
			}
			in.SignatureScript = sigScript
		} else {
			pubKeyBytes := privKey.PubKey().SerializeCompressed()
			script, err := PayToPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
			if err != nil {
				return err
			}
			amount := prevOut.Value
			witness, err := txscript.WitnessSignature(tx, txSigHashes, i, amount, script, txscript.SigHashAll, privKey, true)
			if err != nil {
				return err
			}
			in.Witness = witness

			if txscript.IsPayToScriptHash(prevOut.PkScript) {
				redeemScript, err := PayToWitnessPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
				if err != nil {
					return err
				}
				in.SignatureScript = append([]byte{byte(len(redeemScript))}, redeemScript...)
			}
		}
	}

	return nil
}

func (tool *InscriptionTool) getCommitTxHex() (string, error) {
	return getTxHex(tool.CommitTx)
}

func (tool *InscriptionTool) getRevealTxHexList() ([]string, error) {
	txHexList := make([]string, len(tool.RevealTx))
	for i := range tool.RevealTx {
		txHex, err := getTxHex(tool.RevealTx[i])
		if err != nil {
			return nil, err
		}
		txHexList[i] = txHex
	}
	return txHexList, nil
}

func (tool *InscriptionTool) calculateFee() (int64, []int64) {
	commitTxFee := int64(0)
	for _, in := range tool.CommitTx.TxIn {
		commitTxFee += tool.CommitTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
	}
	for _, out := range tool.CommitTx.TxOut {
		commitTxFee -= out.Value
	}
	revealTxFees := make([]int64, 0)
	for _, tx := range tool.RevealTx {
		revealTxFee := int64(0)
		for i, in := range tx.TxIn {
			revealTxFee += tool.RevealTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
			revealTxFee -= tx.TxOut[i].Value
			revealTxFees = append(revealTxFees, revealTxFee)
		}
	}
	return commitTxFee, revealTxFees
}

func getTxHex(tx *wire.MsgTx) (string, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}
