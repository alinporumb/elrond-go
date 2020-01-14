package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	accountFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

type resolversContainerFactory struct {
	shardCoordinator  sharding.Coordinator
	messenger         dataRetriever.TopicMessageHandler
	marshalizer       marshal.Marshalizer
	intRandomizer     dataRetriever.IntRandomizer
	dataTrieContainer update.DataTriesContainer
}

type ArgsNewResolversContainerFactory struct {
	ShardCoordinator  sharding.Coordinator
	Messenger         dataRetriever.TopicMessageHandler
	Marshalizer       marshal.Marshalizer
	DataTrieContainer update.DataTriesContainer
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(args ArgsNewResolversContainerFactory) (*resolversContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(args.Messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}

	return &resolversContainerFactory{
		shardCoordinator:  args.ShardCoordinator,
		messenger:         args.Messenger,
		marshalizer:       args.Marshalizer,
		intRandomizer:     &random.ConcurrentSafeIntRandomizer{},
		dataTrieContainer: args.DataTrieContainer,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, resolverSlice, err := rcf.generateTrieNodesResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver dataRetriever.Resolver,
	createChannel bool,
) (dataRetriever.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

func (rcf *resolversContainerFactory) generateTrieNodesResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)
	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		resolver, err := rcf.createTrieNodesResolver(factory.AccountTrieNodesTopic, i)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice = append(resolverSlice, resolver)
		trieId := update.CreateTrieIdentifier(i, accountFactory.UserAccount)
		keys = append(keys, trieId)
	}

	resolver, err := rcf.createTrieNodesResolver(factory.AccountTrieNodesTopic, sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	trieId := update.CreateTrieIdentifier(sharding.MetachainShardId, accountFactory.ValidatorAccount)
	keys = append(keys, trieId)

	resolver, err = rcf.createTrieNodesResolver(factory.ValidatorTrieNodesTopic, sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	trieId = update.CreateTrieIdentifier(sharding.MetachainShardId, accountFactory.ValidatorAccount)
	keys = append(keys, trieId)

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(baseTopic string, shId uint32) (dataRetriever.Resolver, error) {
	topic := baseTopic + rcf.shardCoordinator.CommunicationIdentifier(shId)
	excludePeersFromTopic := topic + rcf.shardCoordinator.CommunicationIdentifier(rcf.shardCoordinator.SelfId())

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, excludePeersFromTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		topic,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		shId,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewTrieNodeResolver(
		resolverSender,
		rcf.dataTrieContainer.Get([]byte(topic)),
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}