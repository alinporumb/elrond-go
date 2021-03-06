package metachain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		nil,
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		nil,
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		nil,
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		nil,
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		nil,
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		nil,
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMiniBlocksCompacter(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		nil,
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)
	assert.Equal(t, process.ErrNilMiniBlocksCompacter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		nil,
		&mock.BlockTrackerMock{},
	)
	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		nil,
	)
	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)
	assert.False(t, ppcm.IsInterfaceNil())
}

func TestPreProcessorsContainerFactory_CreateErrTxPreproc(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{}
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		dataPool,
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewPoolsHolderMock(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
