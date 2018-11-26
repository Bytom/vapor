package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http/httptest"
	"os"
	"testing"

	dbm "github.com/tendermint/tmlibs/db"

	"github.com/vapor/accesstoken"
	"github.com/vapor/blockchain/rpc"
	"github.com/vapor/blockchain/txbuilder"
	"github.com/vapor/consensus"
	"github.com/vapor/testutil"
)

func TestAPIHandler(t *testing.T) {
	a := &API{}
	response := &Response{}

	// init httptest server
	a.buildHandler()
	server := httptest.NewServer(a.handler)
	defer server.Close()

	// create accessTokens
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	a.accessTokens = accesstoken.NewStore(testDB)

	client := &rpc.Client{
		BaseURL:     server.URL,
		AccessToken: "test-user:test-secret",
	}

	cases := []struct {
		path     string
		request  interface{}
		respWant *Response
	}{
		{
			path: "/create-key",
			request: struct {
				Alias    string `json:"alias"`
				Password string `json:"password"`
			}{Alias: "alice", Password: "123456"},
			respWant: &Response{
				Status: "fail",
				Msg:    "wallet not found, please check that the wallet is open",
			},
		},
		{
			path:    "/error",
			request: nil,
			respWant: &Response{
				Status: "fail",
				Msg:    "wallet not found, please check that the wallet is open",
			},
		},
		{
			path:    "/",
			request: nil,
			respWant: &Response{
				Status: "",
				Msg:    "",
			},
		},
		{
			path: "/create-access-token",
			request: struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			}{ID: "test-access-id", Type: "test-access-type"},
			respWant: &Response{
				Status: "success",
				Msg:    "",
				Data:   map[string]interface{}{"id": "test-access-id", "type": "test-access-type", "token": "test-access-id:440d87ae0d625a7fcf076275b18372e09a0899e37ec86398879388de90cb0c67"},
			},
		},
		{
			path:    "/gas-rate",
			request: nil,
			respWant: &Response{
				Status: "success",
				Msg:    "",
				Data:   map[string]interface{}{"gasRate": 1000},
			},
		},
	}

	for _, c := range cases {
		response = &Response{}
		client.Call(context.Background(), c.path, c.request, &response)

		if !testutil.DeepEqual(response.Status, c.respWant.Status) {
			t.Errorf(`got=%#v; want=%#v`, response.Status, c.respWant.Status)
		}
	}
}

func TestEstimateTxGas(t *testing.T) {
	tmplStr := `{"allow_additional_actions":false,"raw_transaction":"0701a8d30201010060015ef39ac9f5a6b0db3eb4b2a54a8d012ef5626c1da5462bc97c7a0a1c11bd8e39bdffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0d6e0ce120001160014b29b9e1b31018d5161e33d0c465bbb6dd1df1556010002013dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffacfc9bc512011600142a38b1e022d42414b76a11b7b63075d08fe90b77000139ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff14011600145427f2318811979c908eb2f06d439d134aa156fe00","signing_instructions":[{"position":0,"witness_components":[{"type":"raw_tx_signature","quorum":1,"keys":[{"xpub":"6c420aa025610d323a55c29a8692e2f909b176e88c3bfc8b78cb64ead1bd5db2c6d8492346d12acea177ed0fa4a4579c4bdf020c8cf10fa99cad72f3d5b7de04","derivation_path":["010100000000000000","0e00000000000000"]}],"signatures":null},{"type":"data","value":"512d84b99c93d51729215de3d796390f762f74692306863e3f3bcb0090b399f4"}]}]}`
	template := txbuilder.Template{}
	err := json.Unmarshal([]byte(tmplStr), &template)
	if err != nil {
		t.Fatal(err)
	}

	estimateResult, err := EstimateTxGas(template)
	if err != nil {
		t.Fatal(err)
	}

	baseRate := float64(100000)
	totalNeu := float64(estimateResult.StorageNeu+estimateResult.VMNeu+flexibleGas*consensus.VMGasRate) / baseRate
	roundingNeu := math.Ceil(totalNeu)
	estimateNeu := int64(roundingNeu) * int64(baseRate)

	if estimateResult.TotalNeu != estimateNeu {
		t.Errorf(`got=%#v; want=%#v`, estimateResult.TotalNeu, estimateNeu)
	}
}

func TestEstimateTxGasRange(t *testing.T) {

	cases := []struct {
		path     string
		tmplStr  string
		respWant *EstimateTxGasResp
	}{
		{
			path:    "/estimate-transaction-gas",
			tmplStr: `{"allow_additional_actions":false,"raw_transaction":"0701a8d30201010060015ef39ac9f5a6b0db3eb4b2a54a8d012ef5626c1da5462bc97c7a0a1c11bd8e39bdffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0d6e0ce120001160014b29b9e1b31018d5161e33d0c465bbb6dd1df1556010002013dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffacfc9bc512011600142a38b1e022d42414b76a11b7b63075d08fe90b77000139ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff14011600145427f2318811979c908eb2f06d439d134aa156fe00","signing_instructions":[{"position":0,"witness_components":[{"type":"raw_tx_signature","quorum":1,"keys":[{"xpub":"6c420aa025610d323a55c29a8692e2f909b176e88c3bfc8b78cb64ead1bd5db2c6d8492346d12acea177ed0fa4a4579c4bdf020c8cf10fa99cad72f3d5b7de04","derivation_path":["010100000000000000","0e00000000000000"]}],"signatures":null},{"type":"data","value":"512d84b99c93d51729215de3d796390f762f74692306863e3f3bcb0090b399f4"}]}]}`,
			respWant: &EstimateTxGasResp{
				TotalNeu: (flexibleGas + 2095) * consensus.VMGasRate,
			},
		},
		/*
			{
				path:    "/estimate-transaction-gas",
				tmplStr: `{"raw_transaction":"0701a8d30201010060015ef39ac9f5a6b0db3eb4b2a54a8d012ef5626c1da5462bc97c7a0a1c11bd8e39bdffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0d6e0ce120001160014b29b9e1b31018d5161e33d0c465bbb6dd1df1556010002013dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc092ebc4120116001468c873fd87f05dc1e6ac5d43cc6889a0338d9ad200013bffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80ea30011600145427f2318811979c908eb2f06d439d134aa156fe00","signing_instructions":[{"position":0,"witness_components":[{"type":"raw_tx_signature","quorum":1,"keys":[{"xpub":"6c420aa025610d323a55c29a8692e2f909b176e88c3bfc8b78cb64ead1bd5db2c6d8492346d12acea177ed0fa4a4579c4bdf020c8cf10fa99cad72f3d5b7de04","derivation_path":["010100000000000000","0e00000000000000"]}],"signatures":null},{"type":"data","value":"512d84b99c93d51729215de3d796390f762f74692306863e3f3bcb0090b399f4"}]}],"allow_additional_actions":false}`,
				respWant: &EstimateTxGasResp{
					TotalNeu: (flexibleGas + 3305) * consensus.VMGasRate,
				},
			},
			{
				path:    "/estimate-transaction-gas",
				tmplStr: `{"raw_transaction":"0701a8d30201010060015ef39ac9f5a6b0db3eb4b2a54a8d012ef5626c1da5462bc97c7a0a1c11bd8e39bdffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0d6e0ce120001160014b29b9e1b31018d5161e33d0c465bbb6dd1df1556010002013dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0cfb9c01201160014e75b5b89f8398b214c64d0621a19f25d716c2f4700013cffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80ade204011600145427f2318811979c908eb2f06d439d134aa156fe00","signing_instructions":[{"position":0,"witness_components":[{"type":"raw_tx_signature","quorum":1,"keys":[{"xpub":"6c420aa025610d323a55c29a8692e2f909b176e88c3bfc8b78cb64ead1bd5db2c6d8492346d12acea177ed0fa4a4579c4bdf020c8cf10fa99cad72f3d5b7de04","derivation_path":["010100000000000000","0e00000000000000"]}],"signatures":null},{"type":"data","value":"512d84b99c93d51729215de3d796390f762f74692306863e3f3bcb0090b399f4"}]}],"allow_additional_actions":false}`,
				respWant: &EstimateTxGasResp{
					TotalNeu: (flexibleGas + 13556) * consensus.VMGasRate,
				},
			},
		*/
	}
	for _, c := range cases {
		template := txbuilder.Template{}
		err := json.Unmarshal([]byte(c.tmplStr), &template)
		if err != nil {
			t.Fatal(err)
		}
		estimateTxGasResp, err := EstimateTxGas(template)
		realTotalNeu := float64(c.respWant.TotalNeu)
		rate := math.Abs((float64(estimateTxGasResp.TotalNeu) - realTotalNeu) / float64(estimateTxGasResp.TotalNeu))
		if rate > 0.2 {
			t.Errorf(`the estimateNeu over realNeu 20%%`)
		}
	}
}
