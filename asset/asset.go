package asset

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"

	"github.com/vapor/consensus"
	dbm "github.com/vapor/database/leveldb"
	chainjson "github.com/vapor/encoding/json"
	"github.com/vapor/errors"
	"github.com/vapor/protocol"
	"github.com/vapor/protocol/bc"
)

// DefaultNativeAsset native BTM asset
var DefaultNativeAsset *Asset

const (
	maxAssetCache = 1000
)

var (
	assetIndexKey  = []byte("AssetIndex")
	assetPrefix    = []byte("Asset:")
	aliasPrefix    = []byte("AssetAlias:")
	extAssetPrefix = []byte("EXA:")
)

func initNativeAsset() {
	alias := consensus.BTMAlias

	definitionBytes, _ := serializeAssetDef(consensus.BTMDefinitionMap)
	DefaultNativeAsset = &Asset{
		AssetID:           *consensus.BTMAssetID,
		Alias:             &alias,
		VMVersion:         1,
		DefinitionMap:     consensus.BTMDefinitionMap,
		RawDefinitionByte: definitionBytes,
	}
}

// AliasKey store asset alias prefix
func aliasKey(name string) []byte {
	return append(aliasPrefix, []byte(name)...)
}

// Key store asset prefix
func Key(id *bc.AssetID) []byte {
	return append(assetPrefix, id.Bytes()...)
}

// ExtAssetKey return store external assets key
func ExtAssetKey(id *bc.AssetID) []byte {
	return append(extAssetPrefix, id.Bytes()...)
}

// pre-define errors for supporting bytom errorFormatter
var (
	ErrDuplicateAlias = errors.New("duplicate asset alias")
	ErrDuplicateAsset = errors.New("duplicate asset id")
	ErrSerializing    = errors.New("serializing asset definition")
	ErrMarshalAsset   = errors.New("failed marshal asset")
	ErrFindAsset      = errors.New("fail to find asset")
	ErrInternalAsset  = errors.New("btm has been defined as the internal asset")
	ErrNullAlias      = errors.New("null asset alias")
)

//NewRegistry create new registry
func NewRegistry(db dbm.DB, chain *protocol.Chain) *Registry {
	initNativeAsset()
	return &Registry{
		db:         db,
		chain:      chain,
		cache:      lru.New(maxAssetCache),
		aliasCache: lru.New(maxAssetCache),
	}
}

// Registry tracks and stores all known assets on a blockchain.
type Registry struct {
	db    dbm.DB
	chain *protocol.Chain

	cacheMu    sync.Mutex
	cache      *lru.Cache
	aliasCache *lru.Cache

	assetIndexMu sync.Mutex
	assetMu      sync.Mutex
}

//Asset describe asset on bytom chain
type Asset struct {
	AssetID           bc.AssetID             `json:"id"`
	Alias             *string                `json:"alias"`
	VMVersion         uint64                 `json:"vm_version"`
	RawDefinitionByte chainjson.HexBytes     `json:"raw_definition_byte"`
	DefinitionMap     map[string]interface{} `json:"definition"`
}

// SaveExtAsset store external asset
func (reg *Registry) SaveExtAsset(a *Asset) error {
	reg.assetMu.Lock()
	defer reg.assetMu.Unlock()

	aliasKey := aliasKey(a.AssetID.String())
	if existed := reg.db.Get(aliasKey); existed != nil {
		return ErrDuplicateAlias
	}

	assetKey := ExtAssetKey(&a.AssetID)
	if existAsset := reg.db.Get(assetKey); existAsset != nil {
		return ErrDuplicateAsset
	}

	rawAsset, err := json.Marshal(a)
	if err != nil {
		return ErrMarshalAsset
	}

	storeBatch := reg.db.NewBatch()
	storeBatch.Set(aliasKey, []byte(a.AssetID.String()))
	storeBatch.Set(assetKey, rawAsset)
	storeBatch.Write()
	return nil
}

// SaveAsset store asset
func (reg *Registry) SaveAsset(a *Asset, alias string) error {
	reg.assetMu.Lock()
	defer reg.assetMu.Unlock()

	aliasKey := aliasKey(alias)
	if existed := reg.db.Get(aliasKey); existed != nil {
		return ErrDuplicateAlias
	}

	assetKey := Key(&a.AssetID)
	if existAsset := reg.db.Get(assetKey); existAsset != nil {
		return ErrDuplicateAsset
	}

	rawAsset, err := json.Marshal(a)
	if err != nil {
		return ErrMarshalAsset
	}

	storeBatch := reg.db.NewBatch()
	storeBatch.Set(aliasKey, []byte(a.AssetID.String()))
	storeBatch.Set(assetKey, rawAsset)
	storeBatch.Write()
	return nil
}

// FindByID retrieves an Asset record along with its signer, given an assetID.
func (reg *Registry) FindByID(ctx context.Context, id *bc.AssetID) (*Asset, error) {
	reg.cacheMu.Lock()
	cached, ok := reg.cache.Get(id.String())
	reg.cacheMu.Unlock()
	if ok {
		return cached.(*Asset), nil
	}

	bytes := reg.db.Get(Key(id))
	if bytes == nil {
		return nil, ErrFindAsset
	}

	asset := &Asset{}
	if err := json.Unmarshal(bytes, asset); err != nil {
		return nil, err
	}

	reg.cacheMu.Lock()
	reg.cache.Add(id.String(), asset)
	reg.cacheMu.Unlock()
	return asset, nil
}

// FindByAlias retrieves an Asset record along with its signer,
// given an asset alias.
func (reg *Registry) FindByAlias(alias string) (*Asset, error) {
	reg.cacheMu.Lock()
	cachedID, ok := reg.aliasCache.Get(alias)
	reg.cacheMu.Unlock()
	if ok {
		return reg.FindByID(nil, cachedID.(*bc.AssetID))
	}

	rawID := reg.db.Get(aliasKey(alias))
	if rawID == nil {
		return nil, errors.Wrapf(ErrFindAsset, "no such asset, alias: %s", alias)
	}

	assetID := &bc.AssetID{}
	if err := assetID.UnmarshalText(rawID); err != nil {
		return nil, err
	}

	reg.cacheMu.Lock()
	reg.aliasCache.Add(alias, assetID)
	reg.cacheMu.Unlock()
	return reg.FindByID(nil, assetID)
}

//GetAliasByID return asset alias string by AssetID string
func (reg *Registry) GetAliasByID(id string) string {
	//btm
	if id == consensus.BTMAssetID.String() {
		return consensus.BTMAlias
	}

	assetID := &bc.AssetID{}
	if err := assetID.UnmarshalText([]byte(id)); err != nil {
		return ""
	}

	asset, err := reg.FindByID(nil, assetID)
	if err != nil {
		return ""
	}

	return *asset.Alias
}

// GetAsset get asset by assetID
func (reg *Registry) GetAsset(id string) (*Asset, error) {
	var assetID bc.AssetID
	if err := assetID.UnmarshalText([]byte(id)); err != nil {
		return nil, err
	}

	if assetID.String() == DefaultNativeAsset.AssetID.String() {
		return DefaultNativeAsset, nil
	}

	asset := &Asset{}
	if interAsset := reg.db.Get(Key(&assetID)); interAsset != nil {
		if err := json.Unmarshal(interAsset, asset); err != nil {
			return nil, err
		}
		return asset, nil
	}

	if extAsset := reg.db.Get(ExtAssetKey(&assetID)); extAsset != nil {
		definitionMap := make(map[string]interface{})
		if err := json.Unmarshal(extAsset, &definitionMap); err != nil {
			return nil, err
		}
		alias := assetID.String()
		asset.Alias = &alias
		asset.AssetID = assetID
		asset.DefinitionMap = definitionMap
		return asset, nil
	}

	return nil, errors.WithDetailf(ErrFindAsset, "no such asset, assetID: %s", id)
}

// ListAssets returns the accounts in the db
func (reg *Registry) ListAssets(id string) ([]*Asset, error) {
	assets := []*Asset{DefaultNativeAsset}

	assetIDStr := strings.TrimSpace(id)
	if assetIDStr == DefaultNativeAsset.AssetID.String() {
		return assets, nil
	}

	if assetIDStr != "" {
		assetID := &bc.AssetID{}
		if err := assetID.UnmarshalText([]byte(assetIDStr)); err != nil {
			return nil, err
		}

		asset := &Asset{}
		interAsset := reg.db.Get(Key(assetID))
		if interAsset != nil {
			if err := json.Unmarshal(interAsset, asset); err != nil {
				return nil, err
			}
			return []*Asset{asset}, nil
		}

		return []*Asset{}, nil
	}

	assetIter := reg.db.IteratorPrefix(assetPrefix)
	defer assetIter.Release()

	for assetIter.Next() {
		asset := &Asset{}
		if err := json.Unmarshal(assetIter.Value(), asset); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

// serializeAssetDef produces a canonical byte representation of an asset
// definition. Currently, this is implemented using pretty-printed JSON.
// As is the standard for Go's map[string] serialization, object keys will
// appear in lexicographic order. Although this is mostly meant for machine
// consumption, the JSON is pretty-printed for easy reading.
func serializeAssetDef(def map[string]interface{}) ([]byte, error) {
	if def == nil {
		def = make(map[string]interface{}, 0)
	}
	return json.MarshalIndent(def, "", "  ")
}

//UpdateAssetAlias updates asset alias
func (reg *Registry) UpdateAssetAlias(id, newAlias string) error {
	oldAlias := reg.GetAliasByID(id)
	newAlias = strings.ToUpper(strings.TrimSpace(newAlias))

	if oldAlias == consensus.BTMAlias || newAlias == consensus.BTMAlias {
		return ErrInternalAsset
	}

	if oldAlias == "" || newAlias == "" {
		return ErrNullAlias
	}

	reg.assetMu.Lock()
	defer reg.assetMu.Unlock()

	if _, err := reg.FindByAlias(newAlias); err == nil {
		return ErrDuplicateAlias
	}

	findAsset, err := reg.FindByAlias(oldAlias)
	if err != nil {
		return err
	}

	storeBatch := reg.db.NewBatch()
	findAsset.Alias = &newAlias
	assetID := &findAsset.AssetID
	rawAsset, err := json.Marshal(findAsset)
	if err != nil {
		return err
	}

	storeBatch.Set(Key(assetID), rawAsset)
	storeBatch.Set(aliasKey(newAlias), []byte(assetID.String()))
	storeBatch.Delete(aliasKey(oldAlias))
	storeBatch.Write()

	reg.cacheMu.Lock()
	reg.aliasCache.Add(newAlias, assetID)
	reg.aliasCache.Remove(oldAlias)
	reg.cacheMu.Unlock()

	return nil
}
