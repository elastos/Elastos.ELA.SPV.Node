package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/elastos/Elastos.ELA.SPV.Node/node"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA/auxpow"
	"github.com/elastos/Elastos.ELA/core"
)

func RegisterAddresses(params Params) (Result, error) {
	addresses, ok := params["addresses"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("[RegisterAddresses] parameter addresses not exist")
	}
	addrs := make([]string, 0, len(addresses))
	for _, address := range addresses {
		switch addr := address.(type) {
		case string:
			addrs = append(addrs, addr)
		default:
			return nil, fmt.Errorf("[RegisterAddresses] address not in string format")
		}
	}

	err := node.Instance.RegisterAddresses(addrs)
	if err != nil {
		return nil, fmt.Errorf("[RegisterAddresses] register addresses failed %s", err.Error())
	}
	return nil, nil
}

func RegisterAddress(params Params) (Result, error) {
	address, ok := params.String("address")
	if !ok {
		return nil, fmt.Errorf("[RegisterAddress] parameter address not exist")
	}

	err := node.Instance.RegisterAddress(address)
	if err != nil {
		return nil, fmt.Errorf("[RegisterAddress] register address %s error %s", address, err.Error())
	}
	return nil, err
}

func GetBlockCount(params Params) (Result, error) {
	tip, err := node.Instance.GetBestHeader()
	if err != nil {
		return 0, err
	}
	return tip.Height, nil
}

func GetBestBlockHash(params Params) (Result, error) {
	tip, err := node.Instance.GetBestHeader()
	if err != nil {
		return nil, err
	}
	return tip.Hash().String(), err
}

func GetBlockHash(params Params) (Result, error) {
	height, ok := params.Uint("index")
	if !ok {
		return nil, fmt.Errorf("[GetBlockHash] parameter index not exist")
	}
	hash, err := node.Instance.GetHeaderHash(height)
	if err != nil {
		return nil, err
	}
	return hash.String(), nil
}

func GetBlock(params Params) (Result, error) {
	hex, ok := params.String("hash")
	if !ok {
		return nil, fmt.Errorf("[GetBlock] parameter hash not exist")
	}
	data, err := common.HexStringToBytes(hex)
	if err != nil {
		return nil, fmt.Errorf("[GetBlock] convert hex string failed %s", err.Error())
	}
	hash, err := common.Uint256FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("[GetBlock] parse Uint256 failed %s", err.Error())
	}

	format, ok := params.Uint("format")
	if !ok {
		format = 1
	}

	return getBlock(hash, format)
}

func GetBlockByHeight(params Params) (Result, error) {
	height, ok := params.Uint("height")
	if !ok {
		return nil, fmt.Errorf("[GetBlockByHeight] parameter height not exist")
	}
	hash, err := node.Instance.GetHeaderHash(height)
	if err != nil {
		return nil, fmt.Errorf("[GetBlockByHeight] query block at height %d failed %s", height, err.Error())
	}

	format, ok := params.Uint("format")
	if !ok {
		format = 1
	}

	return getBlock(hash, format)
}

func GetRawTransaction(params Params) (Result, error) {
	hex, ok := params.String("hash")
	if !ok {
		return nil, fmt.Errorf("[GetRawTransaction] parameter hash not exist")
	}

	data, err := common.HexStringToBytes(hex)
	if err != nil {
		return nil, fmt.Errorf("[GetRawTransaction] convert hash hex string failed %s", err.Error())
	}

	hash, err := common.Uint256FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("[GetRawTransaction] parse hash bytes failed %s", err.Error())
	}
	tx, err := node.Instance.GetTx(hash)
	if err != nil {
		return nil, fmt.Errorf("[GetRawTransaction] query transaction %s failed %s",
			hash.String(), err.Error())
	}
	headerHash, err := node.Instance.GetHeaderHash(tx.Height)
	if err != nil {
		return nil, fmt.Errorf("[GetRawTransaction] query header on height %d failed %s",
			tx.Height, err.Error())
	}
	header, err := node.Instance.GetHeader(headerHash)
	if err != nil {
		return nil, fmt.Errorf("[GetRawTransaction] query header %s failed %s",
			headerHash.String(), err.Error())
	}

	format, ok := params.String("format")
	if ok {
		switch format {
		case "btc":
			buf := new(bytes.Buffer)
			if err := elaTxToBtcTx(&tx.Transaction).Serialize(buf); err != nil {
				return nil, err
			}
			return common.BytesToHexString(buf.Bytes()), nil
		case "ela":
			buf := new(bytes.Buffer)
			if err := tx.Serialize(buf); err != nil {
				return nil, err
			}
			return common.BytesToHexString(buf.Bytes()), nil
		case "json":
			return getTransactionInfo(&header.Header, &tx.Transaction), nil
		default:
			return nil, fmt.Errorf("[GetRawTransaction] unspported format %s", format)
		}
	}

	decoded, ok := params.Bool("format")
	if ok && decoded {
		return getTransactionInfo(&header.Header, &tx.Transaction), nil
	}

	buf := new(bytes.Buffer)
	if err := elaTxToBtcTx(&tx.Transaction).Serialize(buf); err != nil {
		return nil, err
	}
	return common.BytesToHexString(buf.Bytes()), nil
}

func SendRawTransaction(params Params) (Result, error) {
	data, ok := params.String("data")
	if !ok {
		return nil, fmt.Errorf("[SendRawTransaction] parameter data not exist")
	}
	txBytes, err := common.HexStringToBytes(data)
	if err != nil {
		return nil, fmt.Errorf("[SendRawTransaction] parse data hex string failed %s", err.Error())
	}

	format, ok := params.String("format")
	if !ok {
		format = "btc"
	}
	switch format {
	case "btc":
		var btcTx auxpow.BtcTx
		err = btcTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			return nil, fmt.Errorf("[SendRawTransaction] transaction deserialize failed %s", err.Error())
		}
		tx, err := btcTxToElaTx(&btcTx)
		if err != nil {
			return nil, fmt.Errorf(
				"[SendRawTransaction] convert btc transaction to ela transaction failed %s", err.Error())
		}
		node.Instance.SendTransaction(*tx)
		return tx.Hash().String(), nil
	case "ela":
		var tx core.Transaction
		err = tx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			return nil, fmt.Errorf("[SendRawTransaction] transaction deserialize failed %s", err.Error())
		}
		node.Instance.SendTransaction(tx)
		return tx.Hash().String(), nil
	}
	return nil, fmt.Errorf("[SendRawTransaction] unknown transaction format %s", format)
}

func getBlock(hash *common.Uint256, format uint32) (Result, error) {
	storeHeader, err := node.Instance.GetHeader(hash)
	if err != nil {
		return nil, fmt.Errorf("[GetBlock] unknown block with hash %s", hash.String())
	}
	switch format {
	case 0:
		return nil, fmt.Errorf("[GetBlock] serialized format not support in SPV node")
	case 2:
		return getBlockInfo(storeHeader.Header, true)
	}
	return getBlockInfo(storeHeader.Header, false)
}

func getBlockInfo(header core.Header, verbose bool) (*BlockInfo, error) {
	hash := header.Hash()

	txIds, err := node.Instance.GetTxIds(header.Height)
	if err != nil {
		return nil, fmt.Errorf("[GetBlockInfo] query block transactions failed %s", err.Error())
	}
	var txs []interface{}
	if verbose {
		for _, txId := range txIds {
			tx, err := node.Instance.GetTx(txId)
			if err != nil {
				return nil, fmt.Errorf("[GetBlockInfo] query transaction %s failed %s",
					txId.String(), err.Error())
			}
			txs = append(txs, getTransactionInfo(&header, &tx.Transaction))
		}
	} else {
		for _, txId := range txIds {
			txs = append(txs, common.BytesToHexString(txId.Bytes()))
		}
	}
	var versionBytes [4]byte
	binary.BigEndian.PutUint32(versionBytes[:], header.Version)

	var chainWork [4]byte
	binary.BigEndian.PutUint32(chainWork[:], node.Instance.BestHeight()-header.Height)

	nextBlockHash, err := node.Instance.GetHeaderHash(header.Height + 1)
	if err != nil {
		nextBlockHash = new(common.Uint256)
	}

	auxPow := new(bytes.Buffer)
	header.AuxPow.Serialize(auxPow)

	return &BlockInfo{
		Hash:              hash.String(),
		Confirmations:     node.Instance.BestHeight() - header.Height + 1,
		StrippedSize:      0,
		Size:              0,
		Weight:            0,
		Height:            header.Height,
		Version:           header.Version,
		VersionHex:        common.BytesToHexString(versionBytes[:]),
		MerkleRoot:        header.MerkleRoot.String(),
		Tx:                txs,
		Time:              header.Timestamp,
		MedianTime:        header.Timestamp,
		Nonce:             header.Nonce,
		Bits:              header.Bits,
		Difficulty:        "",
		ChainWork:         common.BytesToHexString(chainWork[:]),
		PreviousBlockHash: header.Previous.String(),
		NextBlockHash:     nextBlockHash.String(),
		AuxPow:            common.BytesToHexString(auxPow.Bytes()),
	}, nil
}

func getTransactionInfo(header *core.Header, tx *core.Transaction) *TransactionInfo {
	inputs := make([]InputInfo, len(tx.Inputs))
	for i, v := range tx.Inputs {
		inputs[i].TxID = common.BytesToHexString(v.Previous.TxID.Bytes())
		inputs[i].VOut = v.Previous.Index
		inputs[i].Sequence = v.Sequence
	}

	outputs := make([]OutputInfo, len(tx.Outputs))
	for i, v := range tx.Outputs {
		outputs[i].Value = v.Value.String()
		outputs[i].Index = uint32(i)
		address, _ := v.ProgramHash.ToAddress()
		outputs[i].Address = address
		outputs[i].AssetID = common.BytesToHexString(v.AssetID.Bytes())
		outputs[i].OutputLock = v.OutputLock
	}

	attributes := make([]AttributeInfo, len(tx.Attributes))
	for i, v := range tx.Attributes {
		attributes[i].Usage = v.Usage
		attributes[i].Data = common.BytesToHexString(v.Data)
	}

	programs := make([]ProgramInfo, len(tx.Programs))
	for i, v := range tx.Programs {
		programs[i].Code = common.BytesToHexString(v.Code)
		programs[i].Parameter = common.BytesToHexString(v.Parameter)
	}

	var txHash = tx.Hash()
	var txHashStr = txHash.String()
	var size = uint32(tx.GetSize())

	return &TransactionInfo{
		TxId:           txHashStr,
		Hash:           txHashStr,
		Size:           size,
		VSize:          size,
		Version:        header.Version,
		LockTime:       tx.LockTime,
		Inputs:         inputs,
		Outputs:        outputs,
		BlockHash:      header.Hash().String(),
		Confirmations:  node.Instance.BestHeight() - header.Height + 1,
		Time:           header.Timestamp,
		BlockTime:      header.Timestamp,
		TxType:         tx.TxType,
		PayloadVersion: tx.PayloadVersion,
		Payload:        nil,
		Attributes:     attributes,
		Programs:       programs,
	}
}

func elaTxToBtcTx(elaTx *core.Transaction) *auxpow.BtcTx {
	btcTx := new(auxpow.BtcTx)
	inputs := make([]*auxpow.BtcTxIn, 0, len(elaTx.Inputs))
	for _, txIn := range elaTx.Inputs {
		var input auxpow.BtcTxIn
		input.PreviousOutPoint.Hash = txIn.Previous.TxID
		input.PreviousOutPoint.Index = uint32(txIn.Previous.Index)
		input.Sequence = txIn.Sequence
		inputs = append(inputs, &input)
	}
	btcTx.TxIn = inputs

	outputs := make([]*auxpow.BtcTxOut, 0, len(elaTx.Outputs))
	for _, txOut := range elaTx.Outputs {
		var output auxpow.BtcTxOut
		output.PkScript = txOut.ProgramHash.Bytes()
		output.Value = txOut.Value.IntValue()
		outputs = append(outputs, &output)
	}
	btcTx.TxOut = outputs
	btcTx.LockTime = elaTx.LockTime

	return btcTx
}

func btcTxToElaTx(btcTx *auxpow.BtcTx) (*core.Transaction, error) {
	elaTx := new(core.Transaction)
	elaTx.TxType = core.TransferAsset
	elaTx.Payload = new(core.PayloadTransferAsset)
	attr := core.NewAttribute(core.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	elaTx.Attributes = []*core.Attribute{&attr}

	inputs := make([]*core.Input, 0, len(btcTx.TxIn))
	programMap := make(map[string]*core.Program, 0)
	for _, txIn := range btcTx.TxIn {
		var input core.Input
		input.Previous.TxID = txIn.PreviousOutPoint.Hash
		input.Previous.Index = uint16(txIn.PreviousOutPoint.Index)
		input.Sequence = txIn.Sequence
		inputs = append(inputs, &input)

		var program core.Program
		program.Deserialize(bytes.NewReader(txIn.SignatureScript))
		key := common.BytesToHexString(program.Code)
		programMap[key] = &program
	}

	elaTx.Inputs = inputs
	elaTx.Programs = make([]*core.Program, 0, len(programMap))
	for _, program := range programMap {
		elaTx.Programs = append(elaTx.Programs, program)
	}

	outputs := make([]*core.Output, 0, len(btcTx.TxOut))
	for _, out := range btcTx.TxOut {
		var output core.Output
		output.AssetID = node.AssetEla
		output.Value = common.Fixed64(out.Value)
		hash, err := common.Uint168FromBytes(out.PkScript)
		if err != nil {
			return nil, err
		}
		output.ProgramHash = *hash
		outputs = append(outputs, &output)
	}

	elaTx.Outputs = outputs
	elaTx.LockTime = btcTx.LockTime

	return elaTx, nil
}
