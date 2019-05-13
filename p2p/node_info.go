package p2p

import (
	"fmt"
	"net"
	"strconv"

	"github.com/tendermint/go-crypto"

	cfg "github.com/vapor/config"
	"github.com/vapor/consensus"
	"github.com/vapor/version"
)

const maxNodeInfoSize = 10240 // 10Kb

//NodeInfo peer node info
type NodeInfo struct {
	PubKey  crypto.PubKeyEd25519 `json:"pub_key"`
	Moniker string               `json:"moniker"`
	Network string               `json:"network"`
	//MagicNumber used to isolate subnets with same network id
	MagicNumber uint64 `json:"magic_number"`
	RemoteAddr  string `json:"remote_addr"`
	ListenAddr  string `json:"listen_addr"`
	Version     string `json:"version"` // major.minor.revision
	// other application specific data
	//field 0: node service flags. field 1: node alias.
	Other []string `json:"other"`
}

func NewNodeInfo(config *cfg.Config, pubkey crypto.PubKeyEd25519, listenAddr string) *NodeInfo {
	other := []string{strconv.FormatUint(uint64(consensus.DefaultServices), 10)}
	if config.NodeAlias != "" {
		other = append(other, config.NodeAlias)
	}
	return &NodeInfo{
		PubKey:      pubkey,
		Moniker:     config.Moniker,
		Network:     config.ChainID,
		MagicNumber: config.P2P.MagicNumber,
		ListenAddr:  listenAddr,
		Version:     version.Version,
		Other:       other,
	}
}

// CompatibleWith checks if two NodeInfo are compatible with eachother.
// CONTRACT: two nodes are compatible if the major version matches and network match
func (info *NodeInfo) CompatibleWith(other *NodeInfo) error {
	if info.Network != other.Network {
		return fmt.Errorf("Peer is on a different network. Peer network: %v, node network: %v", other.Network, info.Network)
	}

	if info.MagicNumber != other.MagicNumber {
		return fmt.Errorf("Network magic number dismatch. Peer network magic number: %v, node network magic number: %v", other.MagicNumber, info.MagicNumber)
	}

	compatible, err := version.CompatibleWith(other.Version)
	if err != nil {
		return err
	}
	if !compatible {
		return fmt.Errorf("Peer is on a different major version. Peer version: %v, node version: %v", other.Version, info.Version)
	}

	return nil
}

func (info *NodeInfo) getPubkey() crypto.PubKeyEd25519 {
	return info.PubKey
}

//ListenHost peer listener ip address
func (info *NodeInfo) listenHost() string {
	host, _, _ := net.SplitHostPort(info.ListenAddr)
	return host
}

//RemoteAddrHost peer external ip address
func (info *NodeInfo) remoteAddrHost() string {
	host, _, _ := net.SplitHostPort(info.RemoteAddr)
	return host
}

//GetNetwork get node info network field
func (info *NodeInfo) GetNetwork() string {
	return info.Network
}

//String representation
func (info NodeInfo) String() string {
	return fmt.Sprintf("NodeInfo{pk: %v, moniker: %v, network: %v [listen %v], version: %v (%v)}", info.PubKey, info.Moniker, info.Network, info.ListenAddr, info.Version, info.Other)
}
