package block

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

const broadcastDelay = 2 * time.Second

func TestInterceptedShardBlockHeaderVerifiedWithCorrectConsensusGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3
	singleSigner := &singlesig.BlsSingleSigner{}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, randomness)
	header = fillHeaderFields(nodesMap[0][0], header, singleSigner)
	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}
}

func TestInterceptedMetaBlockVerifiedWithCorrectConsensusGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Metachain node Generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(
		sharding.MetachainShardId,
		nodesMap,
		round,
		nonce,
		randomness,
	)

	nodesMap[sharding.MetachainShardId][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}

	// all nodes in shard do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}
}

func TestInterceptedShardBlockHeaderWithLeaderSignatureAndRandSeedChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	singleSigner := &singlesig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(kyber.NewSuitePairingBn256())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
		singleSigner,
		keyGen,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	nodeToSendFrom := nodesMap[0][0]

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, randomness)
	header.SetPrevRandSeed(randomness)
	header = fillHeaderFields(nodeToSendFrom, header, singleSigner)
	nodeToSendFrom.BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}
}

func TestInterceptedShardHeaderBlockWithWrongPreviousRandSeedShouldNotBeAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	singleSigner := &singlesig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(kyber.NewSuitePairingBn256())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
		singleSigner,
		keyGen,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	wrongRandomness := []byte("wrong randomness")
	round := uint64(2)
	nonce := uint64(2)
	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, wrongRandomness)

	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		_, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Error(t, err)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		_, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Error(t, err)
	}
}

func fillHeaderFields(proposer *integrationTests.TestProcessorNode, hdr data.HeaderHandler, signer crypto.SingleSigner) data.HeaderHandler {
	leaderSk := proposer.NodeKeys.Sk

	randSeed, _ := signer.Sign(leaderSk, hdr.GetPrevRandSeed())
	hdr.SetRandSeed(randSeed)

	hdrClone := hdr.Clone()
	hdrClone.SetLeaderSignature(nil)
	headerJsonBytes, _ := integrationTests.TestMarshalizer.Marshal(hdrClone)
	leaderSign, _ := signer.Sign(leaderSk, headerJsonBytes)
	hdr.SetLeaderSignature(leaderSign)

	return hdr
}
