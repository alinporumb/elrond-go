package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TODO: increase code coverage with unit tests

type smartContractResults struct {
	*basePreProcess
	chRcvAllScrs                 chan bool
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	scrForBlock                  txsForBlock
	scrPool                      dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	scrProcessor                 process.SmartContractResultProcessor
	accounts                     state.AccountsAdapter
}

// NewSmartContractResultPreprocessor creates a new smartContractResult preprocessor object
func NewSmartContractResultPreprocessor(
	scrDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	scrProcessor process.SmartContractResultProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte),
	gasHandler process.GasHandler,
	economicsFee process.FeeHandler,
) (*smartContractResults, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(scrDataPool) {
		return nil, process.ErrNilUTxDataPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilUTxStorage
	}
	if check.IfNil(scrProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestSmartContractResult == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	bpp := &basePreProcess{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
		gasHandler:       gasHandler,
		economicsFee:     economicsFee,
	}

	scr := &smartContractResults{
		basePreProcess:               bpp,
		storage:                      store,
		scrPool:                      scrDataPool,
		onRequestSmartContractResult: onRequestSmartContractResult,
		scrProcessor:                 scrProcessor,
		accounts:                     accounts,
	}

	scr.chRcvAllScrs = make(chan bool)
	scr.scrPool.RegisterHandler(scr.receivedSmartContractResult)
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)

	return scr, nil
}

// waitForScrHashes waits for a call whether all the requested smartContractResults appeared
func (scr *smartContractResults) waitForScrHashes(waitTime time.Duration) error {
	select {
	case <-scr.chRcvAllScrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested smartContractResults arrived and were saved into the pool
func (scr *smartContractResults) IsDataPrepared(requestedScrs int, haveTime func() time.Duration) error {
	if requestedScrs > 0 {
		log.Debug("requested missing scrs",
			"num scrs", requestedScrs)
		err := scr.waitForScrHashes(haveTime())
		scr.scrForBlock.mutTxsForBlock.Lock()
		missingScrs := scr.scrForBlock.missingTxs
		scr.scrForBlock.missingTxs = 0
		scr.scrForBlock.mutTxsForBlock.Unlock()
		log.Debug("received missing scrs",
			"num scrs", requestedScrs-missingScrs)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTxBlockFromPools removes smartContractResults and miniblocks from associated pools
func (scr *smartContractResults) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil || body.IsInterfaceNil() {
		return process.ErrNilTxBlockBody
	}

	err := scr.removeDataFromPools(body, miniBlockPool, scr.scrPool, block.SmartContractResultBlock)

	return err
}

// RestoreTxBlockIntoPools restores the smartContractResults and miniblocks to associated pools
func (scr *smartContractResults) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return 0, process.ErrNilMiniBlockPool
	}

	scrRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		scrBuff, err := scr.storage.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			log.Debug("unsigned tx from mini block was not found in UnsignedTransactionUnit",
				"sender shard ID", miniBlock.SenderShardID,
				"receiver shard ID", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
			)

			return scrRestored, err
		}

		for txHash, txBuff := range scrBuff {
			tx := smartContractResult.SmartContractResult{}
			err = scr.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return scrRestored, err
			}

			scr.scrPool.AddData([]byte(txHash), &tx, strCache)
		}

		miniBlockHash, err := core.CalculateHash(scr.marshalizer, scr.hasher, miniBlock)
		if err != nil {
			return scrRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock)

		scrRestored += len(miniBlock.TxHashes)
	}

	return scrRestored, nil
}

// ProcessBlockTransactions processes all the smartContractResult from the block.Body, updates the state
func (scr *smartContractResults) ProcessBlockTransactions(
	body block.Body,
	haveTime func() bool,
) error {

	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		// smart contract results are needed to be processed only at destination and only if they are cross shard
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}
		if miniBlock.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			scr.scrForBlock.mutTxsForBlock.RLock()
			txInfoFromMap := scr.scrForBlock.txHashAndInfo[string(txHash)]
			scr.scrForBlock.mutTxsForBlock.RUnlock()
			if txInfoFromMap == nil || txInfoFromMap.tx == nil {
				log.Debug("missing transaction in ProcessBlockTransactions ", "type", block.SmartContractResultBlock, "txHash", txHash)
				return process.ErrMissingTransaction
			}

			currScr, ok := txInfoFromMap.tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err := scr.processSmartContractResult(
				txHash,
				currScr,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveTxBlockToStorage saves smartContractResults from body into storage
func (scr *smartContractResults) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}
		if miniBlock.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		err := scr.saveTxsToStorage(miniBlock.TxHashes, &scr.scrForBlock, scr.storage, dataRetriever.UnsignedTransactionUnit)
		if err != nil {
			return err
		}
	}

	return nil
}

// receivedSmartContractResult is a call back function which is called when a new smartContractResult
// is added in the smartContractResult pool
func (scr *smartContractResults) receivedSmartContractResult(txHash []byte) {
	receivedAllMissing := scr.baseReceivedTransaction(txHash, &scr.scrForBlock, scr.scrPool, block.SmartContractResultBlock)

	if receivedAllMissing {
		scr.chRcvAllScrs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created smartContractResults at this round
func (scr *smartContractResults) CreateBlockStarted() {
	_ = process.EmptyChannel(scr.chRcvAllScrs)

	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.missingTxs = 0
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)
	scr.scrForBlock.mutTxsForBlock.Unlock()
}

// RequestBlockTransactions request for smartContractResults if missing from a block.Body
func (scr *smartContractResults) RequestBlockTransactions(body block.Body) int {
	requestedSCResults := 0
	missingSCResultsForShards := scr.computeMissingAndExistingSCResultsForShards(body)

	scr.scrForBlock.mutTxsForBlock.Lock()
	for senderShardID, mbsTxHashes := range missingSCResultsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			scr.setMissingSCResultsForShard(senderShardID, mbTxHashes)
		}
	}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	for senderShardID, mbsTxHashes := range missingSCResultsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			requestedSCResults += len(mbTxHashes.txHashes)
			scr.onRequestSmartContractResult(senderShardID, mbTxHashes.txHashes)
		}
	}

	return requestedSCResults
}

func (scr *smartContractResults) setMissingSCResultsForShard(senderShardID uint32, mbTxHashes *txsHashesInfo) {
	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: mbTxHashes.receiverShardID}
	for _, txHash := range mbTxHashes.txHashes {
		scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfoToSet}
	}
}

// computeMissingAndExistingSCResultsForShards calculates what smartContractResults are available and what are missing from block.Body
func (scr *smartContractResults) computeMissingAndExistingSCResultsForShards(body block.Body) map[uint32][]*txsHashesInfo {
	scrTxs := block.Body{}
	for _, mb := range body {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}
		if mb.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		scrTxs = append(scrTxs, mb)
	}

	missingTxsForShard := scr.computeExistingAndMissing(
		scrTxs,
		&scr.scrForBlock,
		scr.chRcvAllScrs,
		block.SmartContractResultBlock,
		scr.scrPool)

	return missingTxsForShard
}

// processAndRemoveBadSmartContractResults processed smartContractResults, if scr are with error it removes them from pool
func (scr *smartContractResults) processSmartContractResult(
	smartContractResultHash []byte,
	smartContractResult *smartContractResult.SmartContractResult,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := scr.scrProcessor.ProcessSmartContractResult(smartContractResult)
	if err != nil {
		return err
	}

	//TODO: These lines could be deleted as in this point these values are already set (this action only overwrites them)
	txShardInfoToSet := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.txHashAndInfo[string(smartContractResultHash)] = &txInfo{tx: smartContractResult, txShardInfo: txShardInfoToSet}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	return nil
}

// RequestTransactionsForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if miniBlock == nil {
		return 0
	}

	missingScrsForMiniBlock := scr.computeMissingScrsForMiniBlock(miniBlock)
	if len(missingScrsForMiniBlock) > 0 {
		scr.onRequestSmartContractResult(miniBlock.SenderShardID, missingScrsForMiniBlock)
	}

	return len(missingScrsForMiniBlock)
}

// computeMissingScrsForMiniBlock computes missing smartContractResults for a certain miniblock
func (scr *smartContractResults) computeMissingScrsForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	if miniBlock.Type != block.SmartContractResultBlock {
		return [][]byte{}
	}

	missingSmartContractResults := make([][]byte, 0, len(miniBlock.TxHashes))
	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			scr.scrPool,
			false)

		if tx == nil || tx.IsInterfaceNil() {
			missingSmartContractResults = append(missingSmartContractResults, txHash)
		}
	}

	return sliceUtil.TrimSliceSliceByte(missingSmartContractResults)
}

// getAllScrsFromMiniBlock gets all the smartContractResults from a miniblock into a new structure
func (scr *smartContractResults) getAllScrsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*smartContractResult.SmartContractResult, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := scr.scrPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilUTxDataPool
	}

	// verify if all smartContractResult exists
	scResSlice := make([]*smartContractResult.SmartContractResult, 0, len(mb.TxHashes))
	txHashes := make([][]byte, 0, len(mb.TxHashes))
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilSmartContractResult
		}

		tx, ok := tmp.(*smartContractResult.SmartContractResult)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		txHashes = append(txHashes, txHash)
		scResSlice = append(scResSlice, tx)
	}

	return smartContractResult.TrimSlicePtr(scResSlice), sliceUtil.TrimSliceSliceByte(txHashes), nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (scr *smartContractResults) CreateAndProcessMiniBlocks(
	_ uint32,
	_ uint32,
	_ func() bool,
) (block.MiniBlockSlice, error) {

	return nil, nil
}

// ProcessMiniBlock processes all the smartContractResults from a and saves the processed smartContractResults in local cache complete miniblock
func (scr *smartContractResults) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
) error {

	if miniBlock.Type != block.SmartContractResultBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	processedTxHashes := make([][]byte, 0)

	defer func() {
		if err != nil {
			scr.gasHandler.RemoveGasConsumed(processedTxHashes)
			scr.gasHandler.RemoveGasRefunded(processedTxHashes)
		}
	}()

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for index := range miniBlockScrs {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		err = scr.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockScrs[index],
			miniBlockTxHashes[index],
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard)

		if err != nil {
			return err
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	for index := range miniBlockScrs {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[index])
		if err != nil {
			return err
		}
	}

	txShardInfoToSet := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	scr.scrForBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockScrs[index], txShardInfo: txShardInfoToSet}
	}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	return nil
}

// CreateMarshalizedData marshalizes smartContractResults and creates and saves them into a new structure
func (scr *smartContractResults) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs, err := scr.createMarshalizedData(txHashes, &scr.scrForBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedTxs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	scr.scrForBlock.mutTxsForBlock.RLock()
	scrPool := make(map[string]data.TransactionHandler, len(scr.scrForBlock.txHashAndInfo))
	for txHash, txInfoFromMap := range scr.scrForBlock.txHashAndInfo {
		scrPool[txHash] = txInfoFromMap.tx
	}
	scr.scrForBlock.mutTxsForBlock.RUnlock()

	return scrPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (scr *smartContractResults) IsInterfaceNil() bool {
	return scr == nil
}
