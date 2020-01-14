package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewDataTrieFactory struct {
	StorageConfig    config.StorageConfig
	SyncFolder       string
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}

type dataTrieFactory struct {
	shardCoordinator sharding.Coordinator
	trieStorage      data.StorageManager
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
}

func NewDataTrieFactory(args ArgsNewDataTrieFactory) (*dataTrieFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, sharding.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, data.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, sharding.ErrNilHasher
	}

	dbConfig := storageFactory.GetDBFromConfig(args.StorageConfig.DB)
	dbConfig.FilePath = path.Join(args.SyncFolder, "syncTries")
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(args.StorageConfig.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(args.StorageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
	if err != nil {
		return nil, err
	}

	d := &dataTrieFactory{
		shardCoordinator: args.ShardCoordinator,
		trieStorage:      trieStorage,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
	}

	return d, nil
}

func (d *dataTrieFactory) Create() (update.DataTriesContainer, error) {
	container := state.NewDataTriesHolder()

	for i := uint32(0); i < d.shardCoordinator.NumberOfShards(); i++ {
		err := d.createAndAddOneTrie(i, factory.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := d.createAndAddOneTrie(sharding.MetachainShardId, factory.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = d.createAndAddOneTrie(sharding.MetachainShardId, factory.ValidatorAccount, container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (d *dataTrieFactory) createAndAddOneTrie(shId uint32, accType factory.Type, container update.DataTriesContainer) error {
	dataTrie, err := trie.NewTrie(d.trieStorage, d.marshalizer, d.hasher)
	if err != nil {
		return err
	}

	trieId := update.CreateTrieIdentifier(shId, accType)
	container.Put([]byte(trieId), dataTrie)

	return nil
}

func (d *dataTrieFactory) IsInterfaceNil() bool {
	return d == nil
}