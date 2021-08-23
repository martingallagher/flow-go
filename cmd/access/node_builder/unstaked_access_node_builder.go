package node_builder

import (
	"context"

	"github.com/onflow/flow-go/cmd"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/local"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/network/p2p"
)

type UnstakedAccessNodeBuilder struct {
	*FlowAccessNodeBuilder
}

func NewUnstakedAccessNodeBuilder(anb *FlowAccessNodeBuilder) *UnstakedAccessNodeBuilder {
	return &UnstakedAccessNodeBuilder{
		FlowAccessNodeBuilder: anb,
	}
}

func (fnb *UnstakedAccessNodeBuilder) InitNodeInfo() {
	fnb.NodeID = flow.ZeroID                   // TODO: extract node id from networking key
	fnb.NodeConfig.NetworkKey = fnb.NetworkKey // use the networking that has been passed in
	fnb.NodeConfig.StakingKey = nil            // no staking key for the unstaked node
}

func (fnb *UnstakedAccessNodeBuilder) InitIDProviders() {
	fnb.Module("id providers", func(builder cmd.NodeBuilder, node *cmd.NodeConfig) error {
		fnb.IDTranslator = p2p.NewUnstakedNetworkIDTranslator()

		idCache, err := p2p.NewProtocolStateIDCache(node.Logger, node.State, fnb.ProtocolEvents)
		if err != nil {
			return err
		}

		fnb.IdentityProvider = idCache

		return nil
	})
}

func (builder *UnstakedAccessNodeBuilder) Initialize() cmd.NodeBuilder {

	ctx, cancel := context.WithCancel(context.Background())
	builder.Cancel = cancel

	builder.validateParams()

	builder.InitIDProviders()

	builder.deriveBootstrapPeerIdentities()

	builder.enqueueUnstakedNetworkInit(ctx)

	builder.enqueueConnectWithStakedAN()

	builder.EnqueueMetricsServerInit()

	builder.RegisterBadgerMetrics()

	builder.EnqxueueTracer()

	builder.PreInit(builder.initUnstakedLocal())

	return builder
}

func (builder *UnstakedAccessNodeBuilder) validateParams() {
	if len(builder.bootstrapNodeAddresses) != len(builder.bootstrapNodePublicKeys) {
		builder.Logger.Fatal().Msg("number of bootstrap node addresses and public keys should match")
	}
}

// deriveBootstrapPeerIdentities derives the Flow Identity of the bootstreap peers from the parameters.
// These are the identity of the staked and unstaked AN also acting as the DHT bootstrap server
func (builder *UnstakedAccessNodeBuilder) deriveBootstrapPeerIdentities() {
	ids, err := BootstrapIdentities(builder.bootstrapNodeAddresses, builder.bootstrapNodePublicKeys)
	builder.MustNot(err)
	builder.bootstrapIdentites = ids
}

// initUnstakedLocal initializes the unstaked node ID, network key and network address
// Currently, it reads a node-info.priv.json like any other node.
// TODO: read the node ID from the special bootstrap files
func (builder *UnstakedAccessNodeBuilder) initUnstakedLocal() func(builder cmd.NodeBuilder, node *cmd.NodeConfig) {
	return func(_ cmd.NodeBuilder, node *cmd.NodeConfig) {
		// for an unstaked node, set the identity here explicitly since it will not be found in the protocol state
		self := &flow.Identity{
			NodeID:        node.NodeID,
			NetworkPubKey: node.NetworkKey.PublicKey(),
			StakingPubKey: nil,             // no staking key needed for the unstaked node
			Role:          flow.RoleAccess, // unstaked node can only run as an access node
			Address:       builder.BindAddr,
		}

		me, err := local.New(self, nil)
		builder.MustNot(err).Msg("could not initialize local")
		node.Me = me
	}
}

// Build enqueues the sync engine and the follower engine for the unstaked access node.
// Currently, the unstaked AN only runs the follower engine.
func (anb *UnstakedAccessNodeBuilder) Build() AccessNodeBuilder {
	anb.
		Module("sync engine participants provider", func(builder cmd.NodeBuilder, node *cmd.NodeConfig) error {
			// use the default identifier provider
			anb.SyncEngineParticipantsProvider = node.Middleware.IdentifierProvider()
			return nil
		})
	anb.FlowAccessNodeBuilder.BuildConsensusFollower()
	return anb
}

// enqueueUnstakedNetworkInit enqueues the unstaked network component initialized for the unstaked node
func (builder *UnstakedAccessNodeBuilder) enqueueUnstakedNetworkInit(ctx context.Context) {

	builder.Component("unstaked network", func(_ cmd.NodeBuilder, node *cmd.NodeConfig) (module.ReadyDoneAware, error) {

		// NodeID for the unstaked node on the unstaked network
		unstakedNodeID := node.NodeID

		// Networking key
		unstakedNetworkKey := node.NetworkKey

		// Network Metrics
		// for now we use the empty metrics NoopCollector till we have defined the new unstaked network metrics
		unstakedNetworkMetrics := metrics.NewNoopCollector()

		libP2PFactory, err := builder.initLibP2PFactory(ctx, unstakedNodeID, unstakedNetworkKey)
		builder.MustNot(err)

		msgValidators := unstakedNetworkMsgValidators(unstakedNodeID)

		middleware := builder.initMiddleware(unstakedNodeID, unstakedNetworkMetrics, libP2PFactory, msgValidators...)

		// topology is nil since its automatically managed by libp2p
		network, err := builder.initNetwork(builder.Me, unstakedNetworkMetrics, middleware, nil)
		builder.MustNot(err)

		builder.Network = network
		builder.Middleware = middleware

		// for an unstaked node, the staked network and middleware is set to the same as the unstaked network and middlware
		builder.Network = network
		builder.Middleware = middleware

		builder.Logger.Info().Msgf("network will run on address: %s", builder.BindAddr)

		return builder.Network, err
	})
}

// enqueueConnectWithStakedAN enqueues the upstream connector component which connects the libp2p host of the unstaked
// AN with the staked AN.
// Currently, there is an issue with LibP2P stopping advertisements of subscribed topics if no peers are connected
// (https://github.com/libp2p/go-libp2p-pubsub/issues/442). This means that an unstaked AN could end up not being
// discovered by other unstaked ANs if it subscribes to a topic before connecting to the staked AN. Hence, the need
// of an explicit connect to the staked AN before the node attempts to subscribe to topics.
func (builder *UnstakedAccessNodeBuilder) enqueueConnectWithStakedAN() {
	builder.Component("unstaked network", func(_ cmd.NodeBuilder, _ *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		return newUpstreamConnector(builder.bootstrapIdentites, builder.LibP2PNode, builder.Logger), nil
	})
}
