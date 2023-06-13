package brc20

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type TxInput struct {
	TxId           string
	VOut           uint32
	Amount         int64
	Address        string
	PrivateKey     string
	NonWitnessUtxo string // legacy address need
}

type TxOutput struct {
	Address string
	Amount  int64
}

const (
	txVersion = 2
	nLockTime = 0
)

func Transfer(ins []*TxInput, outs []*TxOutput, network *chaincfg.Params) (string, error) {
	var inputs []*wire.OutPoint
	var nSequences []uint32
	prevOuts := make(map[wire.OutPoint]*wire.TxOut)
	for _, in := range ins {
		txHash, err := chainhash.NewHashFromStr(in.TxId)
		if err != nil {
			return "", err
		}
		prevOut := wire.NewOutPoint(txHash, in.VOut)
		inputs = append(inputs, prevOut)

		prevPkScript, err := AddrToPkScript(in.Address, network)
		if err != nil {
			return "", err
		}
		witnessUtxo := wire.NewTxOut(in.Amount, prevPkScript)
		prevOuts[*prevOut] = witnessUtxo

		nSequences = append(nSequences, wire.MaxTxInSequenceNum)
	}

	var outputs []*wire.TxOut
	for _, out := range outs {
		pkScript, err := AddrToPkScript(out.Address, network)
		if err != nil {
			return "", err
		}
		outputs = append(outputs, wire.NewTxOut(out.Amount, pkScript))
	}

	bp, err := psbt.New(inputs, outputs, txVersion, nLockTime, nSequences)
	if err != nil {
		return "", err
	}

	updater, err := psbt.NewUpdater(bp)
	if err != nil {
		return "", err
	}

	prevOutputFetcher := txscript.NewMultiPrevOutFetcher(prevOuts)

	for i, in := range ins {
		if err = signInput(updater, i, in, prevOutputFetcher, txscript.SigHashAll, network); err != nil {
			return "", err
		}
		if err = psbt.Finalize(bp, i); err != nil {
			return "", err
		}
	}

	buyerSignedTx, err := psbt.Extract(bp)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = buyerSignedTx.Serialize(&buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

func signInput(updater *psbt.Updater, i int, in *TxInput, prevOutFetcher *txscript.MultiPrevOutFetcher, hashType txscript.SigHashType, network *chaincfg.Params) error {
	wif, err := btcutil.DecodeWIF(in.PrivateKey)
	if err != nil {
		return err
	}
	privKey := wif.PrivKey

	prevPkScript, err := AddrToPkScript(in.Address, network)
	if err != nil {
		return err
	}

	if txscript.IsPayToPubKeyHash(prevPkScript) {
		prevTx := wire.NewMsgTx(txVersion)
		var txBytes []byte
		if txBytes, err = hex.DecodeString(in.NonWitnessUtxo); err != nil {
			return err
		}
		if err = prevTx.Deserialize(bytes.NewReader(txBytes)); err != nil {
			return err
		}
		if err = updater.AddInNonWitnessUtxo(prevTx, i); err != nil {
			return err
		}
	} else {
		witnessUtxo := wire.NewTxOut(in.Amount, prevPkScript)
		err = updater.AddInWitnessUtxo(witnessUtxo, i)
		if err != nil {
			return err
		}
	}

	if err = updater.AddInSighashType(hashType, i); err != nil {
		return err
	}

	if txscript.IsPayToTaproot(prevPkScript) {
		internalPubKey := schnorr.SerializePubKey(privKey.PubKey())
		updater.Upsbt.Inputs[i].TaprootInternalKey = internalPubKey

		sigHashes := txscript.NewTxSigHashes(updater.Upsbt.UnsignedTx, prevOutFetcher)
		if hashType == txscript.SigHashAll {
			hashType = txscript.SigHashDefault
		}
		witness, err := txscript.TaprootWitnessSignature(updater.Upsbt.UnsignedTx, sigHashes,
			i, in.Amount, prevPkScript, hashType, privKey)
		if err != nil {
			return err
		}

		updater.Upsbt.Inputs[i].TaprootKeySpendSig = witness[0]
	} else if txscript.IsPayToPubKeyHash(prevPkScript) {
		signature, err := txscript.RawTxInSignature(updater.Upsbt.UnsignedTx, i, prevPkScript, hashType, privKey)
		if err != nil {
			return err
		}
		_, err = updater.Sign(i, signature, privKey.PubKey().SerializeCompressed(), nil, nil)
		if err != nil {
			return err
		}
	} else {
		pubKeyBytes := privKey.PubKey().SerializeCompressed()
		sigHashes := txscript.NewTxSigHashes(updater.Upsbt.UnsignedTx, prevOutFetcher)

		script, err := PayToPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
		if err != nil {
			return err
		}
		signature, err := txscript.RawTxInWitnessSignature(updater.Upsbt.UnsignedTx, sigHashes, i, in.Amount, script, hashType, privKey)
		if err != nil {
			return err
		}

		if txscript.IsPayToScriptHash(prevPkScript) {
			redeemScript, err := PayToWitnessPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
			if err != nil {
				return err
			}
			err = updater.AddInRedeemScript(redeemScript, i)
			if err != nil {
				return err
			}
		}

		_, err = updater.Sign(i, signature, pubKeyBytes, nil, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddrToPkScript(addr string, network *chaincfg.Params) ([]byte, error) {
	address, err := btcutil.DecodeAddress(addr, network)
	if err != nil {
		return nil, err
	}

	return txscript.PayToAddrScript(address)
}

func PayToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
}

func PayToWitnessPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(pubKeyHash).
		Script()
}
