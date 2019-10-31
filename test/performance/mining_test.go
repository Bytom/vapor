package performance

import (
	"os"
	"testing"
	"time"

	"github.com/vapor/account"
	"github.com/vapor/database"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/proposal"
	"github.com/vapor/test"
)

// Function NewBlockTemplate's benchmark - 0.05s
func BenchmarkNewBlockTpl(b *testing.B) {
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		b.Fatal(err)
	}
	accountStore := database.NewAccountStore(testDB)
	accountManager := account.NewManager(accountStore, chain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proposal.NewBlockTemplate(chain, accountManager, uint64(time.Now().UnixNano()/1e6))
	}
}
