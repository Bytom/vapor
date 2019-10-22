package node

import (
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"reflect"

	"github.com/prometheus/prometheus/util/flock"
	log "github.com/sirupsen/logrus"
	cmn "github.com/tendermint/tmlibs/common"
	browser "github.com/toqueteos/webbrowser"

	"github.com/vapor/accesstoken"
	"github.com/vapor/account"
	"github.com/vapor/api"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/pseudohsm"
	cfg "github.com/vapor/config"
	"github.com/vapor/consensus"
	"github.com/vapor/database"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/env"
	"github.com/vapor/event"
	vaporLog "github.com/vapor/log"
	"github.com/vapor/net/websocket"
	"github.com/vapor/netsync"
	"github.com/vapor/proposal/blockproposer"
	"github.com/vapor/protocol"
	"github.com/vapor/protocol/bc/types"
	w "github.com/vapor/wallet"
)

const (
	webHost   = "http://127.0.0.1"
	logModule = "node"
)

// Node represent bytom node
type Node struct {
	cmn.BaseService

	config          *cfg.Config
	eventDispatcher *event.Dispatcher
	syncManager     *netsync.SyncManager

	wallet          *w.Wallet
	accessTokens    *accesstoken.CredentialStore
	notificationMgr *websocket.WSNotificationManager
	api             *api.API
	chain           *protocol.Chain
	cpuMiner        *blockproposer.BlockProposer
	miningEnable    bool
}

// NewNode create bytom node
func NewNode(config *cfg.Config) *Node {
	if err := lockDataDirectory(config); err != nil {
		cmn.Exit("Error: " + err.Error())
	}

	if err := cfg.LoadFederationFile(config.FederationFile(), config); err != nil {
		cmn.Exit(cmn.Fmt("Failed to load federated information:[%s]", err.Error()))
	}

	if err := vaporLog.InitLogFile(config); err != nil {
		log.WithField("err", err).Fatalln("InitLogFile failed")
	}

	log.WithFields(log.Fields{
		"module":             logModule,
		"pubkey":             config.PrivateKey().XPub(),
		"fed_xpubs":          config.Federation.Xpubs,
		"fed_quorum":         config.Federation.Quorum,
		"fed_controlprogram": hex.EncodeToString(cfg.FederationWScript(config.Federation)),
	}).Info()

	if err := consensus.InitActiveNetParams(config.ChainID); err != nil {
		log.Fatalf("Failed to init ActiveNetParams:[%s]", err.Error())
	}

	initCommonConfig(config)

	// Get store
	if config.DBBackend != "memdb" && config.DBBackend != "leveldb" {
		cmn.Exit(cmn.Fmt("Param db_backend [%v] is invalid, use leveldb or memdb", config.DBBackend))
	}
	coreDB := dbm.NewDB("core", config.DBBackend, config.DBDir())
	store := database.NewStore(coreDB)

	tokenDB := dbm.NewDB("accesstoken", config.DBBackend, config.DBDir())
	accessTokens := accesstoken.NewStore(tokenDB)

	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)
	chain, err := protocol.NewChain(store, txPool, dispatcher)
	if err != nil {
		cmn.Exit(cmn.Fmt("Failed to create chain structure: %v", err))
	}

	if err := checkConfig(chain, config); err != nil {
		panic(err)
	}

	var accounts *account.Manager
	var assets *asset.Registry
	var wallet *w.Wallet

	hsm, err := pseudohsm.New(config.KeysDir())
	if err != nil {
		cmn.Exit(cmn.Fmt("initialize HSM failed: %v", err))
	}

	if !config.Wallet.Disable {
		walletDB := dbm.NewDB("wallet", config.DBBackend, config.DBDir())
		walletStore := database.NewWalletStore(walletDB)
		accountStore := database.NewAccountStore(walletDB)
		accounts = account.NewManager(accountStore, chain)
		assets = asset.NewRegistry(walletDB, chain)
		wallet, err = w.NewWallet(walletStore, accounts, assets, hsm, chain, dispatcher, config.Wallet.TxIndex)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": err}).Error("init NewWallet")
		}

		// trigger rescan wallet
		if config.Wallet.Rescan {
			wallet.RescanBlocks()
		}
	}
	fastSyncDB := dbm.NewDB("fastsync", config.DBBackend, config.DBDir())
	syncManager, err := netsync.NewSyncManager(config, chain, txPool, dispatcher, fastSyncDB)
	if err != nil {
		cmn.Exit(cmn.Fmt("Failed to create sync manager: %v", err))
	}

	notificationMgr := websocket.NewWsNotificationManager(config.Websocket.MaxNumWebsockets, config.Websocket.MaxNumConcurrentReqs, chain, dispatcher)

	// run the profile server
	profileHost := config.ProfListenAddress
	if profileHost != "" {
		// Profiling vapord programs.see (https://blog.golang.org/profiling-go-programs)
		// go tool pprof http://profileHose/debug/pprof/heap
		go func() {
			if err = http.ListenAndServe(profileHost, nil); err != nil {
				cmn.Exit(cmn.Fmt("Failed to register tcp profileHost: %v", err))
			}
		}()
	}

	node := &Node{
		eventDispatcher: dispatcher,
		config:          config,
		syncManager:     syncManager,
		accessTokens:    accessTokens,
		wallet:          wallet,
		chain:           chain,
		miningEnable:    config.Mining,

		notificationMgr: notificationMgr,
	}

	node.cpuMiner = blockproposer.NewBlockProposer(chain, accounts, txPool, dispatcher)
	node.BaseService = *cmn.NewBaseService(nil, "Node", node)
	return node
}

// find whether config xpubs equal genesis block xpubs
func checkConfig(chain *protocol.Chain, config *cfg.Config) error {
	fedpegScript := cfg.FederationWScript(config.Federation)
	genesisBlock, err := chain.GetBlockByHeight(0)
	if err != nil {
		return err
	}
	typedInput := genesisBlock.Transactions[0].Inputs[0].TypedInput
	if v, ok := typedInput.(*types.CoinbaseInput); ok {
		if !reflect.DeepEqual(fedpegScript, v.Arbitrary) {
			return errors.New("config xpubs don't equal genesis block xpubs.")
		}
	}
	return nil
}

// Lock data directory after daemonization
func lockDataDirectory(config *cfg.Config) error {
	_, _, err := flock.New(filepath.Join(config.RootDir, "LOCK"))
	if err != nil {
		return errors.New("datadir already used by another process")
	}
	return nil
}

func initCommonConfig(config *cfg.Config) {
	cfg.CommonConfig = config
}

// Lanch web broser or not
func launchWebBrowser(port string) {
	webAddress := webHost + ":" + port
	log.Info("Launching System Browser with :", webAddress)
	if err := browser.Open(webAddress); err != nil {
		log.Error(err.Error())
		return
	}
}

func (n *Node) initAndstartAPIServer() {
	n.api = api.NewAPI(n.syncManager, n.wallet, n.cpuMiner, n.chain, n.config, n.accessTokens, n.eventDispatcher, n.notificationMgr)

	listenAddr := env.String("LISTEN", n.config.ApiAddress)
	env.Parse()
	n.api.StartServer(*listenAddr)
}

func (n *Node) OnStart() error {
	if n.miningEnable {
		if _, err := n.wallet.AccountMgr.GetMiningAddress(); err != nil {
			n.miningEnable = false
			log.Error(err)
		} else {
			n.cpuMiner.Start()
		}
	}
	if !n.config.VaultMode {
		if err := n.syncManager.Start(); err != nil {
			return err
		}
	}

	n.initAndstartAPIServer()
	if err := n.notificationMgr.Start(); err != nil {
		return err
	}

	if !n.config.Web.Closed {
		_, port, err := net.SplitHostPort(n.config.ApiAddress)
		if err != nil {
			log.Error("Invalid api address")
			return err
		}
		launchWebBrowser(port)
	}
	return nil
}

func (n *Node) OnStop() {
	n.notificationMgr.Shutdown()
	n.notificationMgr.WaitForShutdown()
	n.BaseService.OnStop()
	if n.miningEnable {
		n.cpuMiner.Stop()
	}
	if !n.config.VaultMode {
		n.syncManager.Stop()
	}
	n.eventDispatcher.Stop()
}

func (n *Node) RunForever() {
	// Sleep forever and then...
	cmn.TrapSignal(func() {
		n.Stop()
	})
}
