package twopc

import (
	"bytes"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common"
	"github.com/RosettaFlow/Carrier-Go/common/bytesutil"
	"github.com/RosettaFlow/Carrier-Go/common/rlputil"
	"github.com/RosettaFlow/Carrier-Go/common/timeutils"
	"github.com/RosettaFlow/Carrier-Go/common/traceutil"
	ctypes "github.com/RosettaFlow/Carrier-Go/consensus/twopc/types"
	ev "github.com/RosettaFlow/Carrier-Go/core/evengine"
	"github.com/RosettaFlow/Carrier-Go/core/resource"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	twopcpb "github.com/RosettaFlow/Carrier-Go/lib/netmsg/consensus/twopc"
	libtypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"github.com/RosettaFlow/Carrier-Go/p2p"
	"github.com/RosettaFlow/Carrier-Go/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"strings"
	"time"
)

const (
	//defaultCleanExpireProposalInterval  = 30 * time.Millisecond
	defaultRefreshProposalStateInternal = 300 * time.Millisecond
)

type Twopc struct {
	config                   *Config
	p2p                      p2p.P2P
	state                    *state
	resourceMng              *resource.Manager
	needReplayScheduleTaskCh chan *types.NeedReplayScheduleTask // send remote task to `Scheduler` to replay
	needExecuteTaskCh        chan *types.NeedExecuteTask        // send has was consensus remote tasks to taskManager
	asyncCallCh              chan func()
	quit                     chan struct{}
	taskConsResultCh         chan *types.TaskConsResult
	wal                      *walDB
	Errs                     []error
}

func New(
	conf *Config,
	resourceMng *resource.Manager,
	p2p p2p.P2P,
	needReplayScheduleTaskCh chan *types.NeedReplayScheduleTask,
	needExecuteTaskCh chan *types.NeedExecuteTask,
	taskConsResultCh chan *types.TaskConsResult,
) *Twopc {
	newWalDB := newWal(conf)
	return &Twopc{
		config:                   conf,
		p2p:                      p2p,
		state:                    newState(newWalDB),
		resourceMng:              resourceMng,
		needReplayScheduleTaskCh: needReplayScheduleTaskCh,
		needExecuteTaskCh:        needExecuteTaskCh,
		asyncCallCh:              make(chan func(), conf.PeerMsgQueueSize),
		quit:                     make(chan struct{}),
		taskConsResultCh:         taskConsResultCh,
		wal:                      newWalDB,
		Errs:                     make([]error, 0),
	}
}

func (t *Twopc) Start() error {
	t.recoverCache()
	go t.loop()
	log.Info("Started 2pc consensus engine ...")
	return nil
}
func (t *Twopc) Stop() error {
	close(t.quit)
	return nil
}
func (t *Twopc) loop() {
	refreshProposalStateTicker := time.NewTicker(defaultRefreshProposalStateInternal) // 300 ms
	for {
		select {

		case fn := <-t.asyncCallCh:
			fn()

		case <-refreshProposalStateTicker.C:

			t.refreshProposalState()

		case <-t.quit:
			log.Info("Stopped 2pc consensus engine ...")
			return
		}
	}
}

func (t *Twopc) OnConsensusMsg(pid peer.ID, msg types.ConsensusMsg) error {

	switch msg := msg.(type) {
	case *types.PrepareMsgWrap:
		return t.onPrepareMsg(pid, msg, types.RemoteNetworkMsg)
	case *types.PrepareVoteWrap:
		return t.onPrepareVote(pid, msg, types.RemoteNetworkMsg)
	case *types.ConfirmMsgWrap:
		return t.onConfirmMsg(pid, msg, types.RemoteNetworkMsg)
	case *types.ConfirmVoteWrap:
		return t.onConfirmVote(pid, msg, types.RemoteNetworkMsg)
	case *types.CommitMsgWrap:
		return t.onCommitMsg(pid, msg, types.RemoteNetworkMsg)
	case *types.TerminateConsensusMsgWrap: // Must be  local msg
		return t.onTerminateTaskConsensus(pid, msg)
	default:
		return fmt.Errorf("Unknown the 2pc msg type")

	}
}

func (t *Twopc) OnError() error {
	if len(t.Errs) == 0 {
		return nil
	}
	errStrs := make([]string, len(t.Errs))
	for _, err := range t.Errs {
		errStrs = append(errStrs, err.Error())
	}
	// reset Errs
	t.Errs = make([]error, 0)
	return fmt.Errorf("%s", strings.Join(errStrs, "\n"))
}

func (t *Twopc) OnPrepare(task *types.NeedConsensusTask) error {

	return nil
}
func (t *Twopc) OnHandle(nonConsTask *types.NeedConsensusTask) error {

	task := nonConsTask.GetTask()
	if t.state.HasProposalTaskWithPartyId(task.GetTaskId(), task.GetTaskSender().GetPartyId()) {
		log.Errorf("Failed to check org proposalTask whether have been not exist on OnHandle, but it's alreay exist, taskId: {%s}, partyId: {%s}",
			task.GetTaskId(), task.GetTaskSender().GetPartyId())
		t.stopTaskConsensus(ctypes.ErrPrososalTaskIsProcessed.Error(), common.Hash{}, task.GetTaskId(),
			apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, task.GetTaskSender(), task.GetTaskSender(),
			types.TaskConsensusInterrupt)
		return ctypes.ErrPrososalTaskIsProcessed
	}

	// Store task execute status `cons` before consensus when send task prepareMsg to remote peers
	if err := t.resourceMng.GetDB().StoreLocalTaskExecuteStatusValConsByPartyId(task.GetTaskId(), task.GetTaskSender().GetPartyId()); nil != err {
		log.WithError(err).Errorf("Failed to store local task about `cons` status on OnHandle,  taskId: {%s}, partyId: {%s}",
			task.GetTaskId(), task.GetTaskSender().GetPartyId())
		t.stopTaskConsensus("store task executeStatus about `cons` failed", common.Hash{}, task.GetTaskId(),
			apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, task.GetTaskSender(), task.GetTaskSender(),
			types.TaskConsensusInterrupt)
		return err
	}

	var buf bytes.Buffer
	buf.Write(t.config.Option.NodeID.Bytes())
	buf.Write([]byte(task.GetTaskId()))
	buf.Write([]byte(task.GetTaskData().GetTaskName()))
	buf.Write(bytesutil.Uint64ToBytes(task.GetTaskData().GetCreateAt()))
	buf.Write(bytesutil.Uint64ToBytes(nonConsTask.GetElectionAt()))
	proposalId := rlputil.RlpHash(buf.Bytes())

	log.Infof("Generate proposal, proposalId: {%s}, taskId: {%s}, partyId: {%s}", proposalId.String(), task.GetTaskId(), task.GetTaskSender().GetPartyId())

	// Store some local cache
	t.storeOrgProposalState(
		proposalId,
		task.GetTaskId(),
		task.GetTaskSender(),
		ctypes.NewOrgProposalState(task.GetTaskId(), apicommonpb.TaskRole_TaskRole_Sender, task.GetTaskSender(), nonConsTask.GetElectionAt()),
	)
	proposalTask := types.NewProposalTask(proposalId, task.GetTaskId(), nonConsTask.GetElectionAt())
	t.state.StoreProposalTaskWithPartyId(task.GetTaskSender().GetPartyId(), proposalTask)
	t.wal.StoreProposalTask(task.GetTaskSender().GetPartyId(), proposalTask)
	// Start handle task ...
	go func() {

		if err := t.sendPrepareMsg(proposalId, nonConsTask, nonConsTask.GetElectionAt()); nil != err {
			log.Errorf("Failed to call `sendPrepareMsg`, consensus epoch finished, proposalId: {%s}, taskId: {%s}, partyId: {%s}, err: \n%s",
				proposalId.String(), task.GetTaskId(), task.GetTaskSender().GetPartyId(), err)
			// Send consensus result to Scheduler
			t.stopTaskConsensus("send prepareMsg failed", proposalId, task.GetTaskId(),
				apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, task.GetTaskSender(), task.GetTaskSender(), types.TaskConsensusInterrupt)
			// clean some invalid data
			t.removeOrgProposalStateAndTask(proposalId, task.GetTaskSender().GetPartyId())
		}
	}()
	return nil
}

// Handle the prepareMsg from the task pulisher peer (on Subscriber)
func (t *Twopc) onPrepareMsg(pid peer.ID, prepareMsg *types.PrepareMsgWrap, nmls types.NetworkMsgLocationSymbol) error {

	msg, err := fetchPrepareMsg(prepareMsg)
	if nil != err {
		return err
	}
	log.WithField("traceId", traceutil.GenerateTraceID(prepareMsg.PrepareMsg)).Debugf("Received prepareMsg, consensusSymbol: {%s}, remote pid: {%s}, prepareMsg: %s", nmls.String(), pid, msg.String())

	// the prepareMsg is too late.
	now := timeutils.UnixMsecUint64()
	if (now - msg.GetCreateAt()) >= uint64(ctypes.PrepareMsgVotingDuration.Milliseconds()) {
		log.Errorf("received the prepareMsg is too late on onPrepareMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, now: {%d}, msgCreateAt: {%d}, duration: {%d}, valid duration: {%d}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(),
			now, msg.GetCreateAt(), now-msg.GetCreateAt(), ctypes.PrepareMsgVotingDuration.Milliseconds())
		return ctypes.ErrProposalIllegal
	}

	if t.state.HasOrgProposalWithPartyId(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been not exist on onPrepareMsg, but it's alreay exist, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalAlreadyProcessed
	}

	identity, err := t.resourceMng.GetDB().QueryIdentity()
	if nil != err {
		log.WithError(err).Errorf("Failed to call `QueryIdentity()` on onPrepareMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("query local identity failed, %s", err)
	}

	sender := fetchOrgByPartyRole(msg.GetMsgOption().GetSenderPartyId(), msg.GetMsgOption().GetSenderRole(), msg.GetTask())
	receiver := fetchOrgByPartyRole(msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetReceiverRole(), msg.GetTask())
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check msg.MsgOption sender and receiver on onPrepareMsg, some one is empty, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	// verify the receiver is myself ?
	if identity.GetIdentityId() != receiver.GetIdentityId() {
		log.Errorf("Failed to verify receiver identityId of prepareMsg, receiver is not me, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, my identityId: {%s}, receiver identityId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), identity.GetIdentityId(), receiver.GetIdentityId())
		return ctypes.ErrConsensusMsgInvalid
	}

	org := &apicommonpb.TaskOrganization{
		PartyId:    msg.GetMsgOption().GetReceiverPartyId(),
		NodeName:   identity.GetNodeName(),
		NodeId:     identity.GetNodeId(),
		IdentityId: identity.GetIdentityId(),
	}

	// If you have already voted then we will not vote again.
	// Cause the local message will only call the local function once,
	// and the remote message needs to prevent receiving the repeated forwarded consensus message.
	if nmls == types.RemoteNetworkMsg && t.state.HasPrepareVoting(msg.GetMsgOption().GetProposalId(), org) {
		log.Errorf("Failed to check remote peer prepare vote wether exist on onPrepareMsg, it's exist alreay, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrPrepareVotehadVoted
	}

	// Store task execute status `cons` before consensus when received a remote task prepareMsg
	if err := t.resourceMng.GetDB().StoreLocalTaskExecuteStatusValConsByPartyId(msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverPartyId()); nil != err {
		log.WithError(err).Errorf("Failed to store local task about `cons` status on onPrepareMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return err
	}

	// Store some local cache
	t.storeOrgProposalState(
		msg.GetMsgOption().GetProposalId(),
		msg.GetTask().GetTaskId(),
		sender,
		ctypes.NewOrgProposalState(msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole(), receiver, msg.GetCreateAt()),
	)

	proposalTask := types.NewProposalTask(msg.GetMsgOption().GetProposalId(), msg.GetTask().GetTaskId(), msg.GetCreateAt())
	t.state.StoreProposalTaskWithPartyId(msg.GetMsgOption().GetReceiverPartyId(), proposalTask)
	t.wal.StoreProposalTask(msg.GetMsgOption().GetReceiverPartyId(), proposalTask)

	// Send task to Scheduler to replay sched.
	needReplayScheduleTask := types.NewNeedReplayScheduleTask(msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetTask(), msg.GetNonce(), msg.GetWeights(), msg.GetCreateAt())
	t.sendNeedReplayScheduleTask(needReplayScheduleTask)
	replayTaskResult := needReplayScheduleTask.ReceiveResult()

	log.Debugf("Received the reschedule task result from `schedule.ReplaySchedule()`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, the result: %s",
		msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), replayTaskResult.String())

	var (
		vote       *twopcpb.PrepareVote
		content    string
		voteOption types.VoteOption
		resource   *types.PrepareVoteResource
	)

	if nil != replayTaskResult.GetErr() {
		voteOption = types.NO
		resource = &types.PrepareVoteResource{}
		content = fmt.Sprintf("prepare voting `NO` for proposal '%s', as %s", msg.GetMsgOption().GetProposalId().TerminalString(), replayTaskResult.GetErr())

		log.WithError(replayTaskResult.GetErr()).Warnf("Failed to replay schedule task on onPrepareMsg, replay result has err, will vote `NO`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
	} else {
		voteOption = types.YES
		resource = types.NewPrepareVoteResource(
			replayTaskResult.GetResource().GetId(),
			replayTaskResult.GetResource().GetIp(),
			replayTaskResult.GetResource().GetPort(),
			replayTaskResult.GetResource().GetPartyId(),
		)
		content = fmt.Sprintf("prepare voting `YES` for proposal '%s'", msg.GetMsgOption().GetProposalId().TerminalString())

		log.Infof("Succeed to replay schedule task on onPrepareMsg, will vote `YES`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
	}
	vote = makePrepareVote(
		msg.GetMsgOption().GetProposalId(),
		msg.GetMsgOption().GetReceiverRole(),
		msg.GetMsgOption().GetSenderRole(),
		msg.GetMsgOption().GetReceiverPartyId(),
		msg.GetMsgOption().GetSenderPartyId(),
		receiver,
		voteOption,
		resource,
		timeutils.UnixMsecUint64(),
	)

	// store event about prepare vote
	t.resourceMng.GetDB().StoreTaskEvent(&libtypes.TaskEvent{
		Type:       ev.TaskConsensusPrepareEpoch.GetType(),
		TaskId:     proposalTask.GetTaskId(),
		IdentityId: receiver.GetIdentityId(),
		PartyId:    receiver.GetPartyId(),
		Content:    content,
		CreateAt:   timeutils.UnixMsecUint64(),
	})

	// Store current peer own vote for checking whether to vote already
	if nmls == types.RemoteNetworkMsg {
		t.state.StorePrepareVote(types.FetchPrepareVote(vote))
	}
	go func() {
		if err := t.sendPrepareVote(pid, receiver, sender, vote); nil != err {
			log.Errorf("failed to call `sendPrepareVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, receiver role:{%s}, receiver partyId:{%s}, receiver peerId: {%s}, err: \n%s",
				msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetSenderRole().String(), msg.GetMsgOption().GetSenderPartyId(), pid, err)

			// release local resource and clean some data  (on task partner)
			t.stopTaskConsensus("send prepareVote failed", msg.GetMsgOption().GetProposalId(), msg.GetTask().GetTaskId(),
				msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
			t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		} else {
			log.WithField("traceId", traceutil.GenerateTraceID(vote)).Debugf("Succceed to call `sendPrepareVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, receiver role:{%s}, receiver partyId:{%s}, receiver peerId: {%s}",
				msg.GetMsgOption().GetProposalId().String(), msg.GetTask().GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(),
				msg.GetMsgOption().GetSenderRole().String(), msg.GetMsgOption().GetSenderPartyId(), pid)
		}
	}()
	return nil
}

// (on Publisher)
func (t *Twopc) onPrepareVote(pid peer.ID, prepareVote *types.PrepareVoteWrap, nmls types.NetworkMsgLocationSymbol) error {

	vote := fetchPrepareVote(prepareVote)

	log.WithField("traceId", traceutil.GenerateTraceID(prepareVote.PrepareVote)).Debugf("Received prepareVote, consensusSymbol: {%s}, remote pid: {%s}, prepareVote: %s", nmls.String(), pid, vote.String())

	if t.state.HasNotOrgProposalWithPartyId(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been exist on onPrepareVote, but it's not exist, proposalId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s onPrepareVote", ctypes.ErrProposalNotFound)
	}
	orgProposalState := t.mustGetOrgProposalState(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())

	// The vote in the consensus prepare epoch can be processed only if the current state is the prepare state
	if orgProposalState.IsNotPreparePeriod() {
		log.Errorf("Failed to check org proposalState priod on onPrepareVote, it's not prepare epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalPrepareVoteTimeout
	}

	// find the task of proposal on proposalTask
	proposalTask, ok := t.state.GetProposalTaskWithPartyId(orgProposalState.GetTaskId(), vote.GetMsgOption().GetReceiverPartyId())
	if !ok {
		log.Errorf("%s on onPrepareVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			ctypes.ErrProposalTaskNotFound, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the prepare vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalTaskNotFound, proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetOwner().GetIdentityId(), vote.GetMsgOption().GetReceiverPartyId())
	}

	task, err := t.resourceMng.GetDB().QueryLocalTask(proposalTask.GetTaskId())
	if nil != err {
		log.WithError(err).Errorf("Failed to query local task on onPrepareVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("Not found local task, on the prepare vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
	}

	sender := fetchOrgByPartyRole(vote.GetMsgOption().GetSenderPartyId(), vote.GetMsgOption().GetSenderRole(), task)
	receiver := fetchOrgByPartyRole(vote.GetMsgOption().GetReceiverPartyId(), vote.GetMsgOption().GetReceiverRole(), task)
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check vote.MsgOption sender and receiver of prepareVote on onPrepareVote, some one is empty, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	identity, err := t.resourceMng.GetDB().QueryIdentity()
	if nil != err {
		log.WithError(err).Errorf("Failed to call `QueryIdentity()` on onPrepareVote, some one is empty, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())

		// Send consensus result to interrupt consensus epoch and clean some data (on task sender)
		t.stopTaskConsensus(fmt.Sprintf("query local identity failed %s, when received prepareVote", err), vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("query local identity failed, %s", err)
	}
	// verify the receiver is myself ?
	if identity.GetIdentityId() != receiver.GetIdentityId() {
		log.Errorf("Failed to verify receiver identityId of prepareVote, receiver is not me, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	// Voter <the vote sender> voted repeatedly
	if t.state.HasPrepareVoting(vote.GetMsgOption().GetProposalId(), sender) {
		log.Errorf("%s on onPrepareVote, they are not same, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, vote sender role: {%s}, vote sender partyId: {%s}",
			ctypes.ErrPrepareVoteRepeatedly, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(),
			vote.GetMsgOption().GetSenderRole().String(), vote.GetMsgOption().GetSenderPartyId())
		return fmt.Errorf("%s, on the prepare vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrPrepareVoteRepeatedly, proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetOwner().GetIdentityId(), vote.GetMsgOption().GetReceiverPartyId())
	}

	identityValid, err := t.verifyPrepareVoteRoleIsTaskPartner(sender.GetIdentityId(), sender.GetPartyId(), vote.GetMsgOption().GetSenderRole(), task)
	if nil != err {
		log.WithError(err).Errorf("Failed to call `verifyPrepareVoteRoleIsTaskPartner()` verify prepare vote role on onPrepareVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return err
	}
	if !identityValid {
		log.Errorf("The prepare vote role is not include task partners on onPrepareVote, they are not same, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the prepare vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalPrepareVoteOwnerInvalid, proposalTask.GetTaskId(), vote.GetMsgOption().GetSenderRole().String(), sender.GetIdentityId(), sender.GetPartyId())
	}

	// verify resource of `YES` vote
	if vote.VoteOption == types.YES && vote.PeerInfoEmpty() {
		log.Errorf("%s on onPrepareVote, they are not same, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			ctypes.ErrProposalPrepareVoteResourceInvalid, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the prepare vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalPrepareVoteResourceInvalid, proposalTask.GetTaskId(), vote.GetMsgOption().GetSenderRole().String(), sender.GetIdentityId(), sender.GetPartyId())
	}

	// Store vote
	t.state.StorePrepareVote(vote)

	totalNeedVoteCount := uint32(len(task.GetTaskData().GetDataSuppliers()) +
		len(task.GetTaskData().GetPowerSuppliers()) +
		len(task.GetTaskData().GetReceivers()))

	yesVoteCount := t.state.GetTaskPrepareYesVoteCount(vote.GetMsgOption().GetProposalId())
	totalVotedCount := t.state.GetTaskPrepareTotalVoteCount(vote.GetMsgOption().GetProposalId())

	if totalNeedVoteCount == totalVotedCount {

		now := timeutils.UnixMsecUint64()

		// send confirm msg by option `start` to other remote peers,
		// (announce other peer to continue consensus epoch to confirm epoch)
		// and change proposal state from prepare epoch to confirm epoch
		if totalNeedVoteCount == yesVoteCount {

			// change state from prepare epoch to confirm epoch
			t.state.ChangeToConfirm(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId(), now)

			// store confirm peers resource info
			peers := t.makeConfirmTaskPeerDesc(vote.GetMsgOption().GetProposalId())
			t.storeConfirmTaskPeerInfo(vote.GetMsgOption().GetProposalId(), peers)

			go func() {

				log.Infof("PrepareVoting succeed on consensus prepare epoch, the `YES` vote count has enough, will send `START` confirm msg, the `YES` vote count: {%d}, need total count: {%d}, with proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
					yesVoteCount, totalNeedVoteCount, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
				if err := t.sendConfirmMsg(vote.GetMsgOption().GetProposalId(), task, peers, types.TwopcMsgStart, now); nil != err {
					log.Errorf("Failed to call `sendConfirmMsg` with `start` consensus prepare epoch on `onPrepareVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, err: \n%s",
						vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(), err)
					// Send consensus result to interrupt consensus epoch and clean some data (on task sender)
					t.stopTaskConsensus("send confirmMsg failed", vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
						apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
					t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())
				}
			}()

		} else {

			// send confirm msg by option `stop` to other remote peers,
			// (announce other peer to interrupt consensus epoch)
			// and remove local cache (task/proposal state/prepare vote) about proposal and task
			go func() {

				log.Infof("PrepareVoting failed on consensus prepare epoch, the `YES` vote count is no enough, will send `STOP` confirm msg, the `YES` vote count: {%d}, need total count: {%d}, with proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
					yesVoteCount, totalNeedVoteCount, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())

				if err := t.sendConfirmMsg(vote.GetMsgOption().GetProposalId(), task, t.makeEmptyConfirmTaskPeerDesc(), types.TwopcMsgStop, now); nil != err {
					log.Errorf("Failed to call `sendConfirmMsg` with `stop` consensus prepare epoch on `onPrepareVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, err: \n%s",
						vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(), err)
				}
				// Send consensus result to interrupt consensus epoch and clean some data (on task sender)
				t.stopTaskConsensus("the prepareMsg voting result was not passed", vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
					apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
				t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())
			}()
		}
	}
	return nil
}

// (on Subscriber)
func (t *Twopc) onConfirmMsg(pid peer.ID, confirmMsg *types.ConfirmMsgWrap, nmls types.NetworkMsgLocationSymbol) error {

	msg := fetchConfirmMsg(confirmMsg)

	log.WithField("traceId", traceutil.GenerateTraceID(confirmMsg.ConfirmMsg)).Debugf("Received remote confirmMsg, consensusSymbol: {%s}, remote pid: {%s}, confirmMsg: %s", nmls.String(), pid, msg.String())

	if t.state.HasNotOrgProposalWithPartyId(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been exist on onConfirmMsg, but it's not exist, proposalId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s onConfirmMsg", ctypes.ErrProposalNotFound)
	}

	orgProposalState := t.mustGetOrgProposalState(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())

	// The vote in the consensus prepare epoch or confirm epoch can be processed just if the current state is the prepare state or confirm state.
	if orgProposalState.IsCommitPeriod() {
		log.Errorf("Failed to check org proposalState priod on onConfirmMsg, it's commit epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalConfirmMsgTimeout
	}

	// find the task of proposal on proposalTask
	proposalTask, ok := t.state.GetProposalTaskWithPartyId(orgProposalState.GetTaskId(), msg.GetMsgOption().GetReceiverPartyId())
	if !ok {
		log.Errorf("%s on onConfirmMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			ctypes.ErrProposalTaskNotFound, msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the confirm msg [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalTaskNotFound, msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
	}

	task, err := t.resourceMng.GetDB().QueryLocalTask(proposalTask.GetTaskId())
	if nil != err {
		log.WithError(err).Errorf("Failed to query local task on onConfirmMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("Not found local task, on the confirm msg [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetOwner().GetIdentityId(), msg.GetMsgOption().GetReceiverPartyId())
	}

	sender := fetchOrgByPartyRole(msg.GetMsgOption().GetSenderPartyId(), msg.GetMsgOption().GetSenderRole(), task)
	receiver := fetchOrgByPartyRole(msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetReceiverRole(), task)
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check msg.MsgOption sender and receiver of confirmMsg on onConfirmMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	identity, err := t.resourceMng.GetDB().QueryIdentity()
	if nil != err {
		log.WithError(err).Errorf("Failed to call `QueryIdentity()` on onConfirmMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		// release local resource and clean some data  (on task partner)
		t.stopTaskConsensus(fmt.Sprintf("query local identity failed %s, when received confirmMsg", err), msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("query local identity failed, %s", err)
	}

	// verify the receiver is myself ?
	if identity.GetIdentityId() != receiver.GetIdentityId() {
		log.Errorf("Failed to verify receiver identityId of confirmMsg, receiver is not me, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	org := &apicommonpb.TaskOrganization{
		PartyId:    msg.GetMsgOption().GetReceiverPartyId(),
		NodeName:   identity.GetNodeName(),
		NodeId:     identity.GetNodeId(),
		IdentityId: identity.GetIdentityId(),
	}

	// If you have already voted then we will not vote again.
	// Cause the local message will only call the local function once,
	// and the remote message needs to prevent receiving the repeated forwarded consensus message.
	if nmls == types.RemoteNetworkMsg && t.state.HasConfirmVoting(msg.GetMsgOption().GetProposalId(), org) {
		log.Errorf("Failed to check remote peer confirm vote wether voting on onConfirmMsg, it's voting alreay, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, confirmMsgOption: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetConfirmOption().String())
		return ctypes.ErrConfirmVotehadVoted
	}

	// check msg confirm option value is `start` or `stop` ?
	if msg.GetConfirmOption() == types.TwopcMsgStop || msg.GetConfirmOption() == types.TwopcMsgUnknown {
		log.Warnf("verify confirmMsgOption is not `Start` of confirmMsg on onConfirmMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, confirmMsgOption: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetConfirmOption().String())
		// release local resource and clean some data  (on task partner)
		t.stopTaskConsensus(fmt.Sprintf("check confirm option is %s when received confirmMsg", msg.GetConfirmOption().String()), msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	var (
		vote       *twopcpb.ConfirmVote
		content    string
		voteOption types.VoteOption
	)

	// verify peers resources
	if msg.PeersEmpty() {
		voteOption = types.NO
		content = fmt.Sprintf("confirm voting `NO` for proposal '%s', as received empty peers on confirm msg", msg.GetMsgOption().GetProposalId().TerminalString())

		log.Warnf("Failed to verify peers resources of confirmMsg on onConfirmMsg, the peerDesc reources is empty, will vote `NO`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, confirmMsgOption: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetConfirmOption().String())

	} else {
		// store confirm peers resource info
		t.storeConfirmTaskPeerInfo(msg.GetMsgOption().GetProposalId(), msg.GetPeers())
		voteOption = types.YES
		content = fmt.Sprintf("confirm voting `YES` for proposal '%s'", msg.GetMsgOption().GetProposalId().TerminalString())

		log.Infof("Succeed to verify peers resources of confirmMsg on onConfirmMsg, will vote `YES`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, confirmMsgOption: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetConfirmOption().String())

	}
	vote = makeConfirmVote(
		proposalTask.GetProposalId(),
		msg.GetMsgOption().GetReceiverRole(),
		msg.GetMsgOption().GetSenderRole(),
		msg.GetMsgOption().GetReceiverPartyId(),
		msg.GetMsgOption().GetSenderPartyId(),
		receiver,
		voteOption,
		timeutils.UnixMsecUint64(),
	)

	// store event about confirm vote
	t.resourceMng.GetDB().StoreTaskEvent(&libtypes.TaskEvent{
		Type:       ev.TaskConsensusConfirmEpoch.GetType(),
		TaskId:     proposalTask.GetTaskId(),
		IdentityId: receiver.GetIdentityId(),
		PartyId:    receiver.GetPartyId(),
		Content:    content,
		CreateAt:   timeutils.UnixMsecUint64(),
	})

	// Store current peer own vote for checking whether to vote already
	if nmls == types.RemoteNetworkMsg {
		t.state.StoreConfirmVote(types.FetchConfirmVote(vote))
	}

	// change state from prepare epoch to confirm epoch
	t.state.ChangeToConfirm(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetCreateAt())

	// In any case, as long as voting 'NO', Need to clean the local cache
	if voteOption == types.NO {
		// release local resource and clean some data  (on task partner)
		t.stopTaskConsensus("send `NO` confirmVote when received empty peers confirmMsg", msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
	}

	go func() {

		if err := t.sendConfirmVote(pid, receiver, sender, vote); nil != err {
			log.Errorf("failed to call `sendConfirmVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, receiver role:{%s}, receiver partyId:{%s}, receiver peerId: {%s}, \n%s",
				msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetSenderRole().String(), msg.GetMsgOption().GetSenderPartyId(), pid, err)
			// release local resource and clean some data  (on task partner)
			t.stopTaskConsensus("send confirmVote failed", msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
				msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
			t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		} else {
			log.WithField("traceId", traceutil.GenerateTraceID(vote)).Debugf("Succceed to call `sendConfirmVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, receiver role:{%s}, receiver partyId:{%s}, receiver peerId: {%s}",
				msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetSenderRole().String(), msg.GetMsgOption().GetSenderPartyId(), pid)
		}
	}()

	return nil
}

// (on Publisher)
func (t *Twopc) onConfirmVote(pid peer.ID, confirmVote *types.ConfirmVoteWrap, nmls types.NetworkMsgLocationSymbol) error {

	vote := fetchConfirmVote(confirmVote)

	log.WithField("traceId", traceutil.GenerateTraceID(confirmVote.ConfirmVote)).Debugf("Received confirmVote, consensusSymbol: {%s}, remote pid: {%s}, comfirmVote: %s", nmls.String(), pid, vote.String())

	if t.state.HasNotOrgProposalWithPartyId(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been exist on onConfirmVote, but it's not exist, proposalId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s onConfirmVote", ctypes.ErrProposalNotFound)
	}
	orgProposalState := t.mustGetOrgProposalState(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())

	// The vote in the consensus confirm epoch can be processed only if the current state is the confirm state
	if orgProposalState.IsPreparePeriod() {
		log.Errorf("Failed to check org proposalState priod on onConfirmVote, it's not confirm epoch and is prepare epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalConfirmVoteFuture
	}
	if orgProposalState.IsCommitPeriod() {
		log.Errorf("Failed to check org proposalState priod on onConfirmVote, it's not confirm epoch and is commit epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalPrepareVoteTimeout
	}

	// find the task of proposal on proposalTask
	proposalTask, ok := t.state.GetProposalTaskWithPartyId(orgProposalState.GetTaskId(), vote.GetMsgOption().GetReceiverPartyId())
	if !ok {
		log.Errorf("%s on onConfirmVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			ctypes.ErrProposalTaskNotFound, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the confirm vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalTaskNotFound, proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(),
			vote.GetMsgOption().GetOwner().GetIdentityId(), vote.GetMsgOption().GetReceiverPartyId())
	}

	task, err := t.resourceMng.GetDB().QueryLocalTask(proposalTask.GetTaskId())
	if nil != err {
		log.WithError(err).Errorf("Failed to query local task on onConfirmVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("Not found local task, on the confirm vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(),
			vote.GetMsgOption().GetOwner().GetIdentityId(), vote.GetMsgOption().GetReceiverPartyId())
	}

	sender := fetchOrgByPartyRole(vote.MsgOption.SenderPartyId, vote.MsgOption.SenderRole, task)
	receiver := fetchOrgByPartyRole(vote.MsgOption.ReceiverPartyId, vote.MsgOption.ReceiverRole, task)
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check vote.MsgOption sender and receiver of confirmVote on onConfirmVote, some one is empty, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	identity, err := t.resourceMng.GetDB().QueryIdentity()
	if nil != err {
		log.WithError(err).Errorf("Failed to call `QueryIdentity()` on onConfirmVote, some one is empty, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())

		// Send consensus result to interrupt consensus epoch and clean some data (on task sender)
		t.stopTaskConsensus(fmt.Sprintf("query local identity failed %s, when received confirmVote", err), vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("query local identity failed, %s", err)
	}
	// verify the receiver is myself ?
	if identity.GetIdentityId() != receiver.GetIdentityId() {
		log.Errorf("Failed to verify receiver identityId of confirmVote, receiver is not me, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	// Voter <the vote sender> voted repeatedly
	if t.state.HasConfirmVoting(vote.GetMsgOption().GetProposalId(), sender) {
		log.Errorf("%s on onConfirmVote, they are not same, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, vote sender role: {%s}, vote sender partyId: {%s}",
			ctypes.ErrConfirmVoteRepeatedly, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(),
			vote.GetMsgOption().GetSenderRole().String(), vote.GetMsgOption().GetSenderPartyId())
		return ctypes.ErrConfirmVoteRepeatedly
	}

	identityValid, err := t.verifyConfirmVoteRoleIsTaskPartner(sender.GetIdentityId(), sender.GetPartyId(), vote.GetMsgOption().GetSenderRole(), task)
	if nil != err {
		log.WithError(err).Errorf("Failed to call `verifyConfirmVoteRoleIsTaskPartner()` verify confirm vote role on onConfirmVote, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return err
	}
	if !identityValid {
		log.Errorf("The confirm vote role is not include task partners on onConfirmVote, they are not same, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the confirm vote [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalConfirmVoteVoteOwnerInvalid, proposalTask.GetTaskId(), vote.GetMsgOption().GetSenderRole().String(), sender.GetIdentityId(), sender.GetPartyId())
	}

	// Store vote
	t.state.StoreConfirmVote(vote)

	totalNeedVoteCount := uint32(len(task.GetTaskData().GetDataSuppliers()) +
		len(task.GetTaskData().GetPowerSuppliers()) +
		len(task.GetTaskData().GetReceivers()))

	yesVoteCount := t.state.GetTaskConfirmYesVoteCount(vote.GetMsgOption().GetProposalId())
	totalVotedCount := t.state.GetTaskConfirmTotalVoteCount(vote.GetMsgOption().GetProposalId())

	if totalNeedVoteCount == totalVotedCount {

		now := timeutils.UnixMsecUint64()

		// send commit msg by option `start` to other remote peers,
		// (announce other peer to continue consensus epoch to commit epoch)
		// and change proposal state from confirm epoch to commit epoch
		if totalNeedVoteCount == yesVoteCount {

			// change state from confirm epoch to commit epoch
			t.state.ChangeToCommit(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId(), now)

			go func() {

				log.Debugf("ConfirmVoting succeed on consensus confirm epoch, the `YES` vote count has enough, will send `Start` commit msg, the `YES` vote count: {%d}, need total count: {%d}, with proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
					yesVoteCount, totalNeedVoteCount, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())

				if err := t.sendCommitMsg(vote.GetMsgOption().GetProposalId(), task, types.TwopcMsgStart, now); nil != err {
					log.Errorf("Failed to call `sendCommitMsg` with `start` consensus confirm epoch on `onConfirmVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, err: \n%s",
						vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(), err)
					// Send consensus result (on task sender)
					t.stopTaskConsensus("send commitMsg failed", vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
						apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
				} else {
					// Send consensus result (on task sender)
					t.replyTaskConsensusResult(types.NewTaskConsResult(proposalTask.GetTaskId(), types.TaskConsensusFinished, nil))
				}
				// Finally, whether the commitmsg is sent successfully or not, the local cache needs to be cleared
				t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())

			}()

		} else {

			// send commit msg by option `stop` to other remote peers,
			// (announce other peer to interrupt consensus epoch)
			// and remove local cache (task/proposal state/prepare vote/confirm vote/peerDesc) about proposal and task
			go func() {

				log.Debugf("ConfirmVoting failed on consensus confirm epoch, the `YES` vote count is no enough, will send `Stop` commit msg, the `YES` vote count: {%d}, need total count: {%d}, with proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
					yesVoteCount, totalNeedVoteCount, vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId())

				if err := t.sendCommitMsg(vote.GetMsgOption().GetProposalId(), task, types.TwopcMsgStop, now); nil != err {
					log.Errorf("Failed to call `sendCommitMsg` with `stop` consensus confirm epoch on `onConfirmVote`, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, err: \n%s",
						vote.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), vote.GetMsgOption().GetReceiverRole().String(), vote.GetMsgOption().GetReceiverPartyId(), err)
				}
				// Send consensus result to interrupt consensus epoch and clean some data (on task sender)
				t.stopTaskConsensus("the cofirmMsg voting result was not passed", vote.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
					apicommonpb.TaskRole_TaskRole_Sender, apicommonpb.TaskRole_TaskRole_Sender, receiver, receiver, types.TaskConsensusInterrupt)
				t.removeOrgProposalStateAndTask(vote.GetMsgOption().GetProposalId(), vote.GetMsgOption().GetReceiverPartyId())
			}()
		}
	}
	return nil
}

// (on Subscriber)
func (t *Twopc) onCommitMsg(pid peer.ID, cimmitMsg *types.CommitMsgWrap, nmls types.NetworkMsgLocationSymbol) error {

	msg := fetchCommitMsg(cimmitMsg)

	log.WithField("traceId", traceutil.GenerateTraceID(cimmitMsg.CommitMsg)).Debugf("Received commitMsg, consensusSymbol: {%s}, remote pid: {%s}, commitMsg: %s", nmls.String(), pid, msg.String())

	if t.state.HasNotOrgProposalWithPartyId(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been exist on onCommitMsg, but it's not exist, proposalId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s onCommitMsg", ctypes.ErrProposalNotFound)
	}

	orgProposalState := t.mustGetOrgProposalState(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())

	// The vote in the consensus confirm epoch or commit epoch can be processed just if the current state is the confirm state or commit state
	if orgProposalState.IsPreparePeriod() {
		log.Errorf("Failed to check org proposalState priod on onCommitMsg, it's not commit epoch and is prepare epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalCommitMsgFuture
	}
	if orgProposalState.IsFinishedPeriod() {
		log.Errorf("Failed to check org proposalState priod on onCommitMsg, it's not commit epoch and is finished epoch now, proposalId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrProposalCommitMsgTimeout
	}

	// find the task of proposal on proposalTask
	proposalTask, ok := t.state.GetProposalTaskWithPartyId(orgProposalState.GetTaskId(), msg.GetMsgOption().GetReceiverPartyId())
	if !ok {
		log.Errorf("%s on onCommitMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			ctypes.ErrProposalTaskNotFound, msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("%s, on the commit msg [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			ctypes.ErrProposalTaskNotFound, proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(),
			msg.GetMsgOption().GetOwner().GetIdentityId(), msg.GetMsgOption().GetReceiverPartyId())
	}

	task, err := t.resourceMng.GetDB().QueryLocalTask(proposalTask.GetTaskId())
	if nil != err {
		log.WithError(err).Errorf("Failed to query local task on onCommitMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("Not found local task, on the commit msg [taskId: %s, taskRole: %s, identity: %s, partyId: %s]",
			proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(),
			msg.GetMsgOption().GetOwner().GetIdentityId(), msg.GetMsgOption().GetReceiverPartyId())
	}

	sender := fetchOrgByPartyRole(msg.GetMsgOption().GetSenderPartyId(), msg.GetMsgOption().GetSenderRole(), task)
	receiver := fetchOrgByPartyRole(msg.GetMsgOption().GetReceiverPartyId(), msg.GetMsgOption().GetReceiverRole(), task)
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check msg.MsgOption sender and receiver of commitMsg on onCommitMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	identity, err := t.resourceMng.GetDB().QueryIdentity()
	if nil != err {
		log.WithError(err).Errorf("Failed to call `QueryIdentity()` on onCommitMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		// release local resource and clean some data  (on task partner)
		t.stopTaskConsensus(fmt.Sprintf("query local identity failed %s, when received commitMsg", err), msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		return fmt.Errorf("query local identity failed, %s", err)
	}

	// verify the receiver is myself ?
	if identity.GetIdentityId() != receiver.GetIdentityId() {
		log.Errorf("Failed to verify receiver identityId of commitMsg, receiver is not me, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	// check msg commit option value is `start` or `stop` ?
	if msg.GetCommitOption() == types.TwopcMsgStop || msg.GetCommitOption() == types.TwopcMsgUnknown {
		log.Warnf("verify commitMsgOption is not `Start` of commitMsg on onCommitMsg, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}, confirmMsgOption: {%s}",
			msg.GetMsgOption().GetProposalId().String(), proposalTask.GetTaskId(), msg.GetMsgOption().GetReceiverRole().String(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetCommitOption().String())
		// release local resource and clean some data  (on task partner)
		t.stopTaskConsensus(fmt.Sprintf("check commit option is %s when received commitMsg", msg.GetCommitOption().String()), msg.GetMsgOption().GetProposalId(), proposalTask.GetTaskId(),
			msg.GetMsgOption().GetReceiverRole(), msg.GetMsgOption().GetSenderRole(), receiver, sender, types.TaskConsensusInterrupt)
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	// change state from confirm epoch to commit epoch
	t.state.ChangeToCommit(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId(), msg.GetCreateAt())

	go func() {

		// store succeed consensus event for partyId
		t.resourceMng.GetDB().StoreTaskEvent(&libtypes.TaskEvent{
			Type:       ev.TaskSucceedConsensus.GetType(),
			TaskId:     proposalTask.GetTaskId(),
			IdentityId: receiver.GetIdentityId(),
			PartyId:    receiver.GetPartyId(),
			Content:    fmt.Sprintf("succeed consensus."),
			CreateAt:   timeutils.UnixMsecUint64(),
		})
		// If receiving `CommitMsg` is successful,
		// we will forward `schedTask` to `taskManager` to send it to `Fighter` to execute the task.
		t.driveTask(pid, msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverRole(), receiver, msg.GetMsgOption().GetSenderRole(), sender, proposalTask.GetTaskId())
		t.removeOrgProposalStateAndTask(msg.GetMsgOption().GetProposalId(), msg.GetMsgOption().GetReceiverPartyId())
	}()

	// Finally, it is left 'taskmanager' to call 'releaselocalresourcewithtask()' to release local resources after handle `driveTask()`.
	// No more processing here.
	return nil
}

func (t *Twopc) onTerminateTaskConsensus(pid peer.ID, msg *types.TerminateConsensusMsgWrap) error {

	msgOption := types.FetchMsgOption(msg.GetMsgOption())
	log.Infof("Start terminate task consensus, taskId: {%s}, partyId: {%s}", msg.GetTaskId(), msgOption.GetReceiverPartyId())

	// find the task of proposal on proposalTask
	proposalTask, ok := t.state.GetProposalTaskWithPartyId(msg.GetTaskId(), msgOption.GetReceiverPartyId())
	if !ok {
		log.Errorf("%s on onTerminateTaskConsensus, taskId: {%s}, partyId: {%s}", ctypes.ErrProposalTaskNotFound, msg.GetTaskId(), msgOption.GetReceiverPartyId())
		return fmt.Errorf("%s, on the interrupt consensus [taskId: %s, partyId: %s]",
			ctypes.ErrProposalTaskNotFound, msg.GetTaskId(), msgOption.GetReceiverPartyId())
	}

	task, err := t.resourceMng.GetDB().QueryLocalTask(proposalTask.GetTaskId())
	if nil != err {
		log.WithError(err).Errorf("Failed to query local task on onTerminateTaskConsensus, taskId: {%s}, partyId: {%s}", msg.GetTaskId(), msgOption.GetReceiverPartyId())
		return fmt.Errorf("Not found local task, on the interrupt consensus [taskId: %s, partyId: %s]",
			msg.GetTaskId(), msgOption.GetReceiverPartyId())
	}

	if t.state.HasNotOrgProposalWithPartyId(proposalTask.GetProposalId(), msgOption.GetReceiverPartyId()) {
		log.Errorf("Failed to check org proposalState whether have been exist on onTerminateTaskConsensus, but it's not exist, proposalId: {%s}, taskId: {%s}, partyId: {%s}",
			proposalTask.GetProposalId().String(), msg.GetTaskId(), msgOption.GetReceiverPartyId())
		return fmt.Errorf("%s, on the interrupt consensus", ctypes.ErrProposalNotFound)
	}

	sender := fetchOrgByPartyRole(msgOption.GetSenderPartyId(), msgOption.GetSenderRole(), task)
	receiver := fetchOrgByPartyRole(msgOption.GetReceiverPartyId(), msgOption.GetReceiverRole(), task)
	if nil == sender || nil == receiver {
		log.Errorf("Failed to check msg.MsgOption sender and receiver of interruptMsg on onTerminateTaskConsensus, proposalId: {%s}, taskId: {%s}, role: {%s}, partyId: {%s}",
			proposalTask.GetProposalId().String(), msg.GetTaskId(), msgOption.GetReceiverRole(), msgOption.GetReceiverPartyId())
		return ctypes.ErrConsensusMsgInvalid
	}

	orgProposalState := t.mustGetOrgProposalState(proposalTask.GetProposalId(), msgOption.GetReceiverPartyId())
	switch orgProposalState.CurrPeriodNum() {
	case ctypes.PeriodPrepare:
		// remove `proposal state` and `task cache` AND inerrupt consensus with sender OR release local locked resource with partner
		t.stopTaskConsensus("interrupt consensus with terminate task while prepare epoch", proposalTask.GetProposalId(), msg.GetTaskId(),
			msgOption.GetReceiverRole(), msgOption.GetSenderRole(), receiver, sender, types.TaskTerminate)
		t.removeOrgProposalStateAndTask(proposalTask.GetProposalId(), proposalTask.GetTaskId())
	case ctypes.PeriodConfirm:
		// remove `proposal state` and `task cache` AND inerrupt consensus with sender OR release local locked resource with partner
		t.stopTaskConsensus("interrupt consensus with terminate task while confirm epoch", proposalTask.GetProposalId(), msg.GetTaskId(),
			msgOption.GetReceiverRole(), msgOption.GetSenderRole(), receiver, sender, types.TaskTerminate)
		t.removeOrgProposalStateAndTask(proposalTask.GetProposalId(), proposalTask.GetTaskId())
	case ctypes.PeriodCommit, ctypes.PeriodFinished:
		// need send terminate msg with task manager
		// so do nothing here
	default:
		log.Errorf("unknown org proposalState priod on onTerminateTaskConsensus, proposalId: {%s}, taskId: {%s}, partyId: {%s}, peroid: {%s}",
			proposalTask.GetProposalId().String(), msg.GetTaskId(), msgOption.GetReceiverPartyId(), orgProposalState.GetPeriodStr())
		return fmt.Errorf("unknown org proposalState priod, on the interrupt consensus")
	}
	return nil
}
