package node

import (
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/elastos/Elastos.ELA.SPV/log"
	"github.com/elastos/Elastos.ELA.SPV/sdk"
	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA/bloom"
	"github.com/elastos/Elastos.ELA/core"
)

// single instance of the SPV node
var Instance *SPVNode
var AssetEla = getElaId()

type SPVNode struct {
	sdk.SPVService
	*HeaderStore
	*DataStore
	waitChan chan byte
}

func Init(seeds []string) error {
	var err error
	node := new(SPVNode)
	node.HeaderStore, err = NewHeaderStore()
	if err != nil {
		return err
	}

	node.DataStore, err = NewDataStore()
	if err != nil {
		return err
	}

	var clientId [8]byte
	rand.Read(clientId[:])
	spvClient, err := sdk.GetSPVClient(sdk.TypeMainNet, binary.LittleEndian.Uint64(clientId[:]), seeds)
	if err != nil {
		return err
	}

	node.SPVService, err = sdk.GetSPVService(spvClient, node.HeaderStore, node)
	if err != nil {
		return err
	}

	Instance = node

	return nil
}

func (n *SPVNode) GetData() ([]*common.Uint168, []*core.OutPoint) {
	ops, err := n.DataStore.GetOps()
	if err != nil {
		log.Error("[SPV_NODE] GetData error ", err)
	}

	return n.DataStore.GetAddrs(), ops
}

func (n *SPVNode) OnCommitTx(tx core.Transaction, height uint32) (bool, error) {
	return n.DataStore.PutTx(NewStoreTx(&tx, height))
}

func (n *SPVNode) OnBlockCommitted(bloom.MerkleBlock, []core.Transaction) {
}

func (n *SPVNode) OnRollback(height uint32) error {
	return n.DataStore.Rollback(height)
}

func (n *SPVNode) Start() {
	log.Debug("Wait for register addresses...")
	n.waitChan = make(chan byte)
	<-n.waitChan
	close(n.waitChan)
	n.waitChan = nil

	n.SPVService.Start()
}

func (n *SPVNode) Stop() {
	if n.waitChan != nil {
		close(n.waitChan)
	}
	n.DataStore.Close()
	n.SPVService.Stop()
}

// Interface implements
func (n *SPVNode) RegisterAddresses(addresses []string) error {
	for _, address := range addresses {
		if _, err := n.DataStore.PutAddr(address); err != nil {
			return err
		}
	}
	if n.waitChan == nil {
		return errors.New("RegisterAddresses can only call once on SPV node start," +
			" please use RegisterAddress to register new addresses")
	}
	n.waitChan <- 1
	return nil
}

func (n *SPVNode) RegisterAddress(address string) error {
	ok, err := n.DataStore.PutAddr(address)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("address has already registered")
	}
	n.SPVService.ReloadFilter()
	return nil
}

func (n *SPVNode) BestHeight() uint32 {
	tip, err := n.HeaderStore.GetBestHeader()
	if err != nil {
		return 0
	}
	return tip.Height
}

func getElaId() common.Uint256 {
	// ELA coin
	elaCoin := &core.Transaction{
		TxType:         core.RegisterAsset,
		PayloadVersion: 0,
		Payload: &core.PayloadRegisterAsset{
			Asset: core.Asset{
				Name:      "ELA",
				Precision: 0x08,
				AssetType: 0x00,
			},
			Amount:     0 * 100000000,
			Controller: common.Uint168{},
		},
		Attributes: []*core.Attribute{},
		Inputs:     []*core.Input{},
		Outputs:    []*core.Output{},
		Programs:   []*core.Program{},
	}
	return elaCoin.Hash()
}
