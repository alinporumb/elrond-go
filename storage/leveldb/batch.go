package leveldb

import (
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

type batch struct {
	batch *leveldb.Batch
}

// NewBatch creates a batch
func NewBatch() *batch {
	return &batch{
		batch: &leveldb.Batch{},
	}
}

// Put inserts one entry - key, value pair - into the batch
func (b *batch) Put(key []byte, val []byte) error {
	b.batch.Put(key, val)
	return nil
}

// Delete deletes the entry for the provided key from the batch
func (b *batch) Delete(key []byte) error {
	startTime := time.Now()
	b.batch.Delete(key)
	elapsedTime := time.Since(startTime)
	log.Trace("elapsed time to remove hash from levelDB batch",
		"time [s]", elapsedTime,
	)

	return nil
}

// Reset clears the contents of the batch
func (b *batch) Reset() {
	b.batch.Reset()
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *batch) IsInterfaceNil() bool {
	return b == nil
}
