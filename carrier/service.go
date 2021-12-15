package carrier

import (
	"context"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/auth"
	"github.com/RosettaFlow/Carrier-Go/common/flags"
	"github.com/RosettaFlow/Carrier-Go/consensus/chaincons"
	"github.com/RosettaFlow/Carrier-Go/consensus/twopc"
	"github.com/RosettaFlow/Carrier-Go/core"
	"github.com/RosettaFlow/Carrier-Go/core/election"
	"github.com/RosettaFlow/Carrier-Go/core/evengine"
	"github.com/RosettaFlow/Carrier-Go/core/message"
	"github.com/RosettaFlow/Carrier-Go/core/resource"
	"github.com/RosettaFlow/Carrier-Go/core/schedule"
	"github.com/RosettaFlow/Carrier-Go/core/task"
	"github.com/RosettaFlow/Carrier-Go/db"
	"github.com/RosettaFlow/Carrier-Go/grpclient"
	"github.com/RosettaFlow/Carrier-Go/handler"
	pb "github.com/RosettaFlow/Carrier-Go/lib/api"
	"github.com/RosettaFlow/Carrier-Go/p2p"
	"github.com/RosettaFlow/Carrier-Go/service/discovery"
	"github.com/RosettaFlow/Carrier-Go/types"
	"github.com/urfave/cli/v2"
	"strconv"
	"strings"
	"sync"
)

type Service struct {
	isRunning      bool
	processingLock sync.RWMutex
	config         *Config
	carrierDB      core.CarrierDB
	ctx            context.Context
	cancel         context.CancelFunc
	mempool        *message.Mempool
	Engines        map[types.ConsensusEngineType]handler.Engine

	// DB interfaces
	dataDb     db.Database
	APIBackend *CarrierAPIBackend

	resourceManager *resource.Manager
	messageManager  *message.MessageHandler
	TaskManager     handler.TaskManager
	authManager     *auth.AuthorityManager
	scheduler       schedule.Scheduler
	consulManager   *discovery.ConnectConsul
	runError        error

	// internal resource node set (Fighter node grpc client set)
	resourceClientSet *grpclient.InternalResourceClientSet
}

// NewService creates a new CarrierServer object (including the
// initialisation of the common Carrier object)
func NewService(ctx context.Context, cliCtx *cli.Context, config *Config, mockIdentityIdsFile, consensusStateFile string) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	nodeIdStr := config.P2P.NodeId()
	// read config from p2p config.
	nodeId, _ := p2p.HexID(nodeIdStr)

	pool := message.NewMempool(&message.MempoolConfig{NodeId: nodeIdStr})
	eventEngine := evengine.NewEventEngine(config.CarrierDB)

	// TODO The size of these ch is currently written dead ...
	localTaskMsgCh, needReplayScheduleTaskCh, needExecuteTaskCh, taskConsResultCh :=
		make(chan types.TaskDataArray, 100),
		make(chan *types.NeedReplayScheduleTask, 600),
		make(chan *types.NeedExecuteTask, 600),
		make(chan *types.TaskConsResult, 600)

	resourceClientSet := grpclient.NewInternalResourceNodeSet()
	resourceMng := resource.NewResourceManager(config.CarrierDB, mockIdentityIdsFile)
	authManager := auth.NewAuthorityManager(config.CarrierDB)
	scheduler := schedule.NewSchedulerStarveFIFO(election.NewVrfElector(config.P2P.PirKey(), resourceClientSet, resourceMng), eventEngine, resourceMng, authManager)
	twopcEngine := twopc.New(
		&twopc.Config{
			Option: &twopc.OptionConfig{
				NodePriKey: config.P2P.PirKey(),
				NodeID:     nodeId,
			},
			PeerMsgQueueSize:    1024,
			ConsensusStateFile:  consensusStateFile,
			DefaultConsensusWal: config.DefaultConsensusWal,
			DatabaseCache:       config.DatabaseCache,
			DatabaseHandles:     config.DatabaseHandles,
		},
		resourceMng,
		config.P2P,
		needReplayScheduleTaskCh,
		needExecuteTaskCh,
		taskConsResultCh,
	)
	taskManager := task.NewTaskManager(
		config.P2P,
		scheduler,
		twopcEngine,
		eventEngine,
		resourceMng,
		authManager,
		resourceClientSet,
		localTaskMsgCh,
		needReplayScheduleTaskCh,
		needExecuteTaskCh,
		taskConsResultCh,
	)

	port, _ := strconv.Atoi(cliCtx.String(flags.RPCPort.Name))
	ip := cliCtx.String(flags.RPCHost.Name)

	serverIp := cliCtx.String("discovery-server-ip")
	if serverIp == "" {
		serverIp = "127.0.0.1"
	}
	serverPort := cliCtx.String("discovery-server-port")
	if serverPort == "" {
		serverPort = "8500"
	}
	healthCheckInterval := cliCtx.Int("discovery-health-check-interval")
	if healthCheckInterval <= 0 {
		healthCheckInterval = 3
	}
	healthCheckDeregister := cliCtx.Int("discovery-health-check-deregister")
	if healthCheckDeregister <= 0 {
		healthCheckDeregister = 1
	}
	serviceName := cliCtx.String("discovery-service-name")
	if serviceName == "" {
		serviceName = "carrier"
	}
	serviceId := cliCtx.String("discovery-service-id")
	if serviceId == "" {
		serviceId = "carrier"
	}
	serviceTags := cliCtx.String("discovery-service-tags")
	tagArray := make([]string, 0)
	if serviceTags == "" {
		tagArray = append(tagArray, "carrier")
	} else {
		serviceTags = strings.Replace(serviceTags, " ", "", -1)
		for _, value := range strings.Split(serviceTags, ",") {
			tagArray = append(tagArray, value)
		}
	}
	consul := discovery.New(&discovery.ConsulService{
		IP:         ip,
		Port:       port,
		Tags:       tagArray,
		Name:       serviceName,
		Id:         serviceId,
		Interval:   healthCheckInterval,
		Deregister: healthCheckDeregister,
	},
		serverIp,
		serverPort,
	)
	err := consul.RegisterDiscoveryService()
	if err != nil {
		log.WithError(err).Fatal("RegisterDiscoveryService fail!")
	}
	s := &Service{
		ctx:               ctx,
		cancel:            cancel,
		config:            config,
		carrierDB:         config.CarrierDB,
		mempool:           pool,
		resourceManager:   resourceMng,
		messageManager:    message.NewHandler(pool, config.CarrierDB, taskManager, authManager, resourceClientSet),
		TaskManager:       taskManager,
		authManager:       authManager,
		resourceClientSet: resourceClientSet,
		consulManager:     consul,
	}

	//s.APIBackend = &CarrierAPIBackend{carrier: s}
	s.APIBackend = NewCarrierAPIBackend(s)
	s.Engines = make(map[types.ConsensusEngineType]handler.Engine, 0)
	s.Engines[types.TwopcTyp] = twopcEngine
	s.Engines[types.ChainconsTyp] = chaincons.New()

	// load stored jobNode and dataNode
	jobNodeList, err := s.carrierDB.QueryRegisterNodeList(pb.PrefixTypeJobNode)
	if err == nil {
		for _, node := range jobNodeList {
			client, err := grpclient.NewJobNodeClient(ctx, fmt.Sprintf("%s:%s", node.InternalIp, node.InternalPort), node.Id)
			if err == nil {
				s.resourceClientSet.StoreJobNodeClient(node.Id, client)
			}
		}
	}
	dataNodeList, err := s.carrierDB.QueryRegisterNodeList(pb.PrefixTypeDataNode)
	if err == nil {
		for _, node := range dataNodeList {
			client, err := grpclient.NewDataNodeClient(ctx, fmt.Sprintf("%s:%s", node.InternalIp, node.InternalPort), node.Id)
			if err == nil {
				s.resourceClientSet.StoreDataNodeClient(node.Id, client)
			}
		}
	}
	return s, nil
}

func (s *Service) Start() error {

	if nil != s.authManager {
		if err := s.authManager.Start(); nil != err {
			log.WithError(err).Errorf("Failed to start the authManager, err: %v", err)
		}
	}

	for typ, engine := range s.Engines {
		if err := engine.Start(); nil != err {
			log.WithError(err).Errorf("Cound not start the consensus engine: %s, err: %v", typ.String(), err)
		}
	}
	if nil != s.resourceManager {
		if err := s.resourceManager.Start(); nil != err {
			log.WithError(err).Errorf("Failed to start the resourceManager, err: %v", err)
		}
	}
	if nil != s.messageManager {
		if err := s.messageManager.Start(); nil != err {
			log.WithError(err).Errorf("Failed to start the messageManager, err: %v", err)
		}
	}
	if nil != s.TaskManager {
		if err := s.TaskManager.Start(); nil != err {
			log.WithError(err).Errorf("Failed to start the TaskManager, err: %v", err)
		}
	}
	if nil != s.scheduler {
		if err := s.scheduler.Start(); nil != err {
			log.WithError(err).Errorf("Failed to start the schedule, err: %v", err)
		}
	}

	return nil
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
	s.carrierDB.Stop()

	for typ, engine := range s.Engines {
		if err := engine.Stop(); nil != err {
			log.WithError(err).Errorf("Cound not close the consensus engine: %s, err: %v", typ.String(), err)
		}
	}
	if nil != s.resourceManager {
		if err := s.resourceManager.Stop(); nil != err {
			log.WithError(err).Errorf("Failed to stop the resourceManager, err: %v", err)
		}
	}
	if nil != s.messageManager {
		if err := s.messageManager.Stop(); nil != err {
			log.WithError(err).Errorf("Failed to stop the messageManager, err: %v", err)
		}
	}
	if nil != s.TaskManager {
		if err := s.TaskManager.Stop(); nil != err {
			log.WithError(err).Errorf("Failed to stop the TaskManager, err: %v", err)
		}
	}
	if nil != s.scheduler {
		if err := s.scheduler.Stop(); nil != err {
			log.WithError(err).Errorf("Failed to stop the schedule, err: %v", err)
		}
	}

	if nil != s.authManager {
		if err := s.authManager.Stop(); nil != err {
			log.WithError(err).Errorf("Failed to stop the authManager, err: %v", err)
		}
	}

	return nil
}

// Status is service health checks. Return nil or error.
func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	if s.runError != nil {
		return s.runError
	}
	return nil
}
