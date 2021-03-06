package shard_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
	tmhs := mock.NewTopicMessageHandlerStub()

	tmhs.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}

		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	tmhs.RegisterMessageProcessorCalled = func(topic string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}

		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return tmhs
}

func createDataPools() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}

	return pools
}

func createStore() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

func createTriesHolder() state.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(triesFactory.UserAccountTrie), &mock.TrieStub{})
	triesHolder.Put([]byte(triesFactory.PeerAccountTrie), &mock.TrieStub{})
	return triesHolder
}

//------- NewResolversContainerFactory

func TestNewResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		nil,
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewResolversContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		nil,
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTxStorage, err)
}

func TestNewResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		nil,
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewResolversContainerFactory_NilMarshalizerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		nil,
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		1,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		nil,
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewResolversContainerFactory_NilSliceSplitterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		nil,
		createTriesHolder(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewResolversContainerFactory_NilTrieDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		nil,
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		1,
	)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
}

//------- Create

func TestResolversContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.TransactionTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.ShardBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.MiniBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.PeerChBodyTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.TransactionTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.ShardBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.MiniBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.PeerChBodyTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.AccountTrieNodesTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	rcf, _ := shard.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	rcf, _ := shard.NewResolversContainerFactory(
		shardCoordinator,
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolder(),
		0,
	)

	container, _ := rcf.Create()

	numResolverSCRs := noOfShards + 1
	numResolverTxs := noOfShards + 1
	numResolverRewardTxs := noOfShards + 1
	numResolverHeaders := 1
	numResolverMiniBlocks := noOfShards + 1
	numResolverPeerChanges := 1
	numResolverMetaBlockHeaders := 1
	numResolverTrieNodes := 2
	totalResolvers := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverPeerChanges +
		numResolverMetaBlockHeaders + numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes

	assert.Equal(t, totalResolvers, container.Len())
}
