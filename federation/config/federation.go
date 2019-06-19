package config

import (
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/crypto/ed25519/chainkd"
	"github.com/vapor/protocol/vm/vmutil"
)

func ParseFedProg(warders []Warder, quorum int) []byte {
	SortWarders(warders)

	xpubs := []chainkd.XPub{}
	for _, w := range warders {
		xpubs = append(xpubs, w.XPub)
	}

	fedpegScript, err := vmutil.P2SPMultiSigProgram(chainkd.XPubKeys(xpubs), quorum)
	if err != nil {
		log.Panicf("fail to generate federation scirpt for federation: %v", err)
	}

	return fedpegScript
}

type byPosition []Warder

func (w byPosition) Len() int           { return len(w) }
func (w byPosition) Swap(i, j int)      { w[i], w[j] = w[j], w[i] }
func (w byPosition) Less(i, j int) bool { return w[i].Position < w[j].Position }

func SortWarders(warders []Warder) []Warder {
	sort.Sort(byPosition(warders))
	return warders
}
