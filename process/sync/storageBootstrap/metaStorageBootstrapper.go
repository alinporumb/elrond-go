package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

type metaStorageBootstrapper struct {
	*storageBootstrapper
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler
}

// NewMetaStorageBootstrapper is method used to create a nes storage bootstrapper
func NewMetaStorageBootstrapper(arguments ArgsMetaStorageBootstrapper) (*metaStorageBootstrapper, error) {
	base := &storageBootstrapper{
		bootStorer:       arguments.BootStorer,
		forkDetector:     arguments.ForkDetector,
		blkExecutor:      arguments.BlockProcessor,
		blkc:             arguments.ChainHandler,
		marshalizer:      arguments.Marshalizer,
		store:            arguments.Store,
		shardCoordinator: arguments.ShardCoordinator,
		blockTracker:     arguments.BlockTracker,

		uint64Converter:     arguments.Uint64Converter,
		bootstrapRoundIndex: arguments.BootstrapRoundIndex,
	}

	boot := metaStorageBootstrapper{
		storageBootstrapper:      base,
		pendingMiniBlocksHandler: arguments.PendingMiniBlocksHandler,
	}

	base.bootstrapper = &boot
	base.headerNonceHashStore = boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)

	return &boot, nil
}

// LoadFromStorage will load all blocks from storage
func (msb *metaStorageBootstrapper) LoadFromStorage() error {
	return msb.loadBlocks()
}

// IsInterfaceNil returns true if there is no value under the interface
func (msb *metaStorageBootstrapper) IsInterfaceNil() bool {
	return msb == nil
}

func (msb *metaStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		header, err := process.GetShardHeaderFromStorage(crossNotarizedHeader.Hash, msb.marshalizer, msb.store)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", crossNotarizedHeader.ShardId,
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", crossNotarizedHeader.Hash)

		msb.blockTracker.AddCrossNotarizedHeader(crossNotarizedHeader.ShardId, header, crossNotarizedHeader.Hash)
		msb.blockTracker.AddTrackedHeader(header, crossNotarizedHeader.Hash)
	}

	return nil
}

func (msb *metaStorageBootstrapper) getHeader(hash []byte) (data.HeaderHandler, error) {
	return process.GetMetaHeaderFromStorage(hash, msb.marshalizer, msb.store)
}

func (msb *metaStorageBootstrapper) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return block.Body{}, nil
}

func (msb *metaStorageBootstrapper) cleanupNotarizedStorage(metaBlockHash []byte) {
	log.Debug("cleanup notarized storage")

	metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, msb.marshalizer, msb.store)
	if err != nil {
		log.Debug("meta block is not found in MetaBlockUnit storage",
			"hash", metaBlockHash)
		return
	}

	shardHeaderHashes := make([][]byte, len(metaBlock.ShardInfo))
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		shardHeaderHashes[i] = metaBlock.ShardInfo[i].HeaderHash
	}

	for _, shardHeaderHash := range shardHeaderHashes {
		var shardHeader *block.Header
		shardHeader, err = process.GetShardHeaderFromStorage(shardHeaderHash, msb.marshalizer, msb.store)
		if err != nil {
			log.Debug("shard header is not found in BlockHeaderUnit storage",
				"hash", shardHeaderHash)
			continue
		}

		log.Debug("removing shard header from ShardHdrNonceHashDataUnit storage",
			"shradId", shardHeader.GetShardID(),
			"nonce", shardHeader.GetNonce(),
			"hash", shardHeaderHash)

		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardHeader.GetShardID())
		storer := msb.store.GetStorer(hdrNonceHashDataUnit)
		nonceToByteSlice := msb.uint64Converter.ToByteSlice(shardHeader.GetNonce())
		err = storer.Remove(nonceToByteSlice)
		if err != nil {
			log.Debug("shard header was not removed from ShardHdrNonceHashDataUnit storage",
				"shradId", shardHeader.GetShardID(),
				"nonce", shardHeader.GetNonce(),
				"hash", shardHeaderHash,
				"error", err.Error())
		}
	}
}

func (msb *metaStorageBootstrapper) applySelfNotarizedHeaders(selfNotarizedHeadersHashes [][]byte) ([]data.HeaderHandler, error) {
	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	return selfNotarizedHeaders, nil
}

func (msb *metaStorageBootstrapper) applyNumPendingMiniBlocks(pendingMiniBlocks []bootstrapStorage.PendingMiniBlockInfo) {
	for _, pendingMiniBlockInfo := range pendingMiniBlocks {
		msb.pendingMiniBlocksHandler.SetNumPendingMiniBlocks(pendingMiniBlockInfo.ShardID, pendingMiniBlockInfo.NumPendingMiniBlocks)

		log.Debug("set pending miniblocks",
			"shard", pendingMiniBlockInfo.ShardID,
			"num", pendingMiniBlockInfo.NumPendingMiniBlocks)
	}
}
