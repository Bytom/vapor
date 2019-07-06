package chainmgr

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/vapor/consensus"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/protocol"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/test/mock"
)

const txsNumber = 2000

func getTransactions() []*types.Tx {
	txs := []*types.Tx{}
	for i := 0; i < txsNumber; i++ {
		txInput := types.NewSpendInput(nil, bc.NewHash([32]byte{0x01}), *consensus.BTMAssetID, uint64(i), 1, []byte{0x51})
		txInput.CommitmentSuffix = []byte{0, 1, 2}
		txInput.WitnessSuffix = []byte{0, 1, 2}

		tx := &types.Tx{

			TxData: types.TxData{
				//SerializedSize: uint64(i * 10),
				Inputs: []*types.TxInput{
					txInput,
				},
				Outputs: []*types.TxOutput{
					types.NewIntraChainOutput(*consensus.BTMAssetID, uint64(i), []byte{0x6a}),
				},
			},
			Tx: &bc.Tx{
				ID: bc.Hash{V0: uint64(i), V1: uint64(i), V2: uint64(i), V3: uint64(i)},
			},
		}
		txs = append(txs, tx)
	}
	return txs
}

func TestSyncMempool(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatalf("failed to create temporary data folder: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	testDBA := dbm.NewDB("testdba", "leveldb", "tmpDir")
	testDBB := dbm.NewDB("testdbb", "leveldb", "tmpDir")

	blocks := mockBlocks(nil, 5)
	a := mockSync(blocks, &mock.Mempool{}, testDBA)
	b := mockSync(blocks, &mock.Mempool{}, testDBB)

	netWork := NewNetWork()
	netWork.Register(a, "192.168.0.1", "test node A", consensus.SFFullNode)
	netWork.Register(b, "192.168.0.2", "test node B", consensus.SFFullNode)
	if B2A, A2B, err := netWork.HandsShake(a, b); err != nil {
		t.Errorf("fail on peer hands shake %v", err)
	} else {
		go B2A.postMan()
		go A2B.postMan()
	}

	go a.syncMempoolLoop()
	a.syncMempool("test node B")
	wantTxs := getTransactions()
	a.txSyncCh <- &txSyncMsg{"test node B", wantTxs}

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	gotTxs := []*protocol.TxDesc{}
	for {
		select {
		case <-ticker.C:
			gotTxs = b.mempool.GetTransactions()
			if len(gotTxs) >= txsNumber {
				goto out
			}
		case <-timeout.C:
			t.Fatalf("mempool sync timeout")
		}
	}

out:
	if len(gotTxs) != txsNumber {
		t.Fatalf("mempool sync txs num err. got:%d want:%d", len(gotTxs), txsNumber)
	}

	for i, gotTx := range gotTxs {
		index := gotTx.Tx.Inputs[0].Amount()
		if !reflect.DeepEqual(gotTx.Tx.Inputs[0].Amount(), wantTxs[index].Inputs[0].Amount()) {
			t.Fatalf("mempool tx err. index:%d\n,gotTx:%s\n,wantTx:%s", i, spew.Sdump(gotTx.Tx.Inputs), spew.Sdump(wantTxs[0].Inputs))
		}

		if !reflect.DeepEqual(gotTx.Tx.Outputs[0].AssetAmount(), wantTxs[index].Outputs[0].AssetAmount()) {
			t.Fatalf("mempool tx err. index:%d\n,gotTx:%s\n,wantTx:%s", i, spew.Sdump(gotTx.Tx.Outputs), spew.Sdump(wantTxs[0].Outputs))
		}

	}

}
