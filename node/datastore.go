package node

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"sync"

	"github.com/elastos/Elastos.ELA.SPV/sdk"
	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA/core"
)

var (
	BKTAddrs     = []byte("Addrs")
	BKTTxs       = []byte("Txs")
	BKTHeightTxs = []byte("HeightTxs")
	BKTOps       = []byte("Ops")
)

type DataStore struct {
	*sync.RWMutex
	*bolt.DB
	filter *sdk.AddrFilter
}

func NewDataStore() (*DataStore, error) {
	db, err := bolt.Open("data_store.bin", 0644, &bolt.Options{InitialMmapSize: 5000000})
	if err != nil {
		return nil, err
	}
	store := new(DataStore)
	store.RWMutex = new(sync.RWMutex)
	store.DB = db

	db.Update(func(btx *bolt.Tx) error {
		_, err = btx.CreateBucketIfNotExists(BKTAddrs)
		if err != nil {
			return err
		}
		_, err = btx.CreateBucketIfNotExists(BKTTxs)
		if err != nil {
			return err
		}
		_, err = btx.CreateBucketIfNotExists(BKTHeightTxs)
		if err != nil {
			return err
		}
		_, err = btx.CreateBucketIfNotExists(BKTOps)
		if err != nil {
			return err
		}
		return nil
	})

	addrs, err := store.getAddrs()
	if err != nil {
		return nil, err
	}
	store.filter = sdk.NewAddrFilter(addrs)

	return store, nil
}

func (t *DataStore) PutAddr(address string) error {
	t.Lock()
	defer t.Unlock()

	hash, err := common.Uint168FromAddress(address)
	if err != nil {
		return err
	}

	t.filter.AddAddr(hash)

	return t.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BKTAddrs).Put([]byte(address), hash.Bytes())
	})
}

func (t *DataStore) GetAddrs() []*common.Uint168 {
	t.RLock()
	defer t.RUnlock()

	return t.filter.GetAddrs()
}

func (t *DataStore) getAddrs() (addrs []*common.Uint168, err error) {
	err = t.View(func(tx *bolt.Tx) error {
		return tx.Bucket(BKTAddrs).ForEach(func(k, v []byte) error {
			addr, err := common.Uint168FromBytes(v)
			if err != nil {
				return err
			}
			addrs = append(addrs, addr)
			return nil
		})
	})

	return addrs, err
}

func (t *DataStore) PutTx(txn *StoreTx) (match bool, err error) {
	t.Lock()
	defer t.Unlock()

	err = t.Update(func(tx *bolt.Tx) error {

		for index, output := range txn.Outputs {
			if t.filter.ContainAddr(output.ProgramHash) {
				match = true
				op := core.NewOutPoint(txn.Hash(), uint16(index)).Bytes()
				if err := tx.Bucket(BKTOps).Put(op, op); err != nil {
					return err
				}
			}
		}

		for _, input := range txn.Inputs {
			outpoint := tx.Bucket(BKTOps).Get(input.Previous.Bytes())
			if outpoint != nil {
				match = true
			}
		}

		if match {
			buf := new(bytes.Buffer)
			if err = txn.Serialize(buf); err != nil {
				return err
			}

			if err = tx.Bucket(BKTTxs).Put(txn.Hash().Bytes(), buf.Bytes()); err != nil {
				return err
			}

			var key [4]byte
			binary.LittleEndian.PutUint32(key[:], txn.Height)
			data := tx.Bucket(BKTHeightTxs).Get(key[:])

			var txMap = make(map[common.Uint256]uint32)
			if err = gob.NewDecoder(bytes.NewReader(data)).Decode(&txMap); err != nil {
				return err
			}

			txMap[txn.Hash()] = txn.Height

			buf = new(bytes.Buffer)
			if err = gob.NewEncoder(buf).Encode(txMap); err != nil {
				return err
			}

			return tx.Bucket(BKTHeightTxs).Put(key[:], buf.Bytes())
		}

		return nil
	})

	return match, err
}

func (t *DataStore) GetTx(hash *common.Uint256) (txn *StoreTx, err error) {
	t.RLock()
	defer t.RUnlock()

	err = t.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(BKTTxs).Get(hash.Bytes())
		txn = new(StoreTx)
		return txn.Deserialize(bytes.NewReader(data))
	})

	return txn, err
}

func (t *DataStore) GetTxIds(height uint32) (txIds []*common.Uint256, err error) {
	t.RLock()
	defer t.RUnlock()

	err = t.View(func(tx *bolt.Tx) error {
		var key [4]byte
		binary.LittleEndian.PutUint32(key[:], height)
		data := tx.Bucket(BKTHeightTxs).Get(key[:])

		var txMap = make(map[common.Uint256]uint32)
		err = gob.NewDecoder(bytes.NewReader(data)).Decode(&txMap)
		if err != nil {
			return err
		}

		txIds = make([]*common.Uint256, 0, len(txMap))
		for hash := range txMap {
			txIds = append(txIds, &hash)
		}
		return nil
	})

	return txIds, err
}

func (t *DataStore) GetOps() (ops []*core.OutPoint, err error) {
	t.RLock()
	defer t.RUnlock()

	err = t.View(func(tx *bolt.Tx) error {
		return tx.Bucket(BKTOps).ForEach(func(k, v []byte) error {
			op, err := core.OutPointFromBytes(v)
			if err != nil {
				return err
			}
			ops = append(ops, op)
			return nil
		})
	})

	return ops, err
}

func (t *DataStore) Rollback(height uint32) error {
	t.Lock()
	defer t.Unlock()

	return t.Update(func(tx *bolt.Tx) error {
		var key [4]byte
		binary.LittleEndian.PutUint32(key[:], height)
		data := tx.Bucket(BKTHeightTxs).Get(key[:])

		var txMap = make(map[common.Uint256]uint32)
		err := gob.NewDecoder(bytes.NewReader(data)).Decode(&txMap)
		if err != nil {
			return err
		}

		for hash := range txMap {
			var txn core.Transaction
			data := tx.Bucket(BKTTxs).Get(hash.Bytes())
			if err = txn.Deserialize(bytes.NewReader(data)); err != nil {
				return err
			}
			for index, output := range txn.Outputs {
				if t.filter.ContainAddr(output.ProgramHash) {
					outpoint := core.NewOutPoint(txn.Hash(), uint16(index)).Bytes()
					tx.Bucket(BKTOps).Delete(outpoint)
				}
			}
			if err = tx.Bucket(BKTTxs).Delete(hash.Bytes()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *DataStore) Reset() error {
	t.Lock()
	defer t.Unlock()

	return t.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(BKTTxs)
		if err != nil {
			return err
		}

		return tx.DeleteBucket(BKTHeightTxs)
	})
}

// Close db
func (t *DataStore) Close() {
	t.Lock()
	t.DB.Close()
}
