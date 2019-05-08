package authn

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/vapor/accesstoken"
	dbm "github.com/vapor/database/db"
	_ "github.com/vapor/database/leveldb"
	"github.com/vapor/database/orm"
	_ "github.com/vapor/database/sqlite"
	"github.com/vapor/errors"
)

func TestAuthenticate(t *testing.T) {
	tokenDB := dbm.NewSqlDB("sql", "sqlitedb", "temp")
	defer func() {
		tokenDB.Db().Close()
		os.RemoveAll("temp")
	}()

	tokenDB.Db().AutoMigrate(&orm.AccessToken{})
	tokenStore := accesstoken.NewStore(tokenDB)
	token, err := tokenStore.Create("alice", "test")
	if err != nil {
		t.Errorf("create token error")
	}

	cases := []struct {
		id, tok string
		want    error
	}{
		{"alice", token.Token, nil},
		{"alice", "alice:abcsdsdfassdfsefsfsfesfesfefsefa", ErrInvalidToken},
	}

	api := NewAPI(tokenStore, false)

	for _, c := range cases {
		var username, password string
		toks := strings.SplitN(c.tok, ":", 2)
		if len(toks) > 0 {
			username = toks[0]
		}
		if len(toks) > 1 {
			password = toks[1]
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.SetBasicAuth(username, password)

		_, err := api.Authenticate(req)
		if errors.Root(err) != c.want {
			t.Errorf("Authenticate(%s) error = %s want %s", c.id, err, c.want)
		}
	}
}
