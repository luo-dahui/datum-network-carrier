package types

import (
	"encoding/json"
	"github.com/RosettaFlow/Carrier-Go/common"
	"github.com/RosettaFlow/Carrier-Go/common/bytesutil"
	"github.com/RosettaFlow/Carrier-Go/common/rlputil"
	"time"

	"github.com/RosettaFlow/Carrier-Go/consensus/twopc/utils"
	"github.com/RosettaFlow/Carrier-Go/types"
	"sync/atomic"
)

type taskOption struct {
	Role                  types.TaskRole           `json:"role"` // The role information of the current recipient of the task
	TaskId                string                   `json:"taskId"`
	TaskName              string                   `json:"taskName"`
	Owner                 *types.NodeAlias         `json:"owner"`
	AlgoSupplier          *types.NodeAlias         `json:"algoSupplier"`
	DataSupplier          []*dataSupplierOption    `json:"dataSupplier"`
	PowerSupplier         []*powerSupplierOption   `json:"powerSupplier"`
	Receivers             []*receiverOption        `json:"receivers"`
	OperationCost         *types.TaskOperationCost `json:"operationCost"`
	CalculateContractCode string                   `json:"calculateContractCode"`
	DataSplitContractCode string                   `json:"dataSplitContractCode"`
	CreateAt              uint64                   `json:"createat"`
}

func (t *taskOption) Hash() common.Hash {
	return rlputil.RlpHash(t)
}

type dataSupplierOption struct {
	MemberInfo      *types.NodeAlias `json:"memberInfo"`
	MetaDataId      string           `json:"metaDataId"`
	ColumnIndexList []uint64         `json:"columnIndexList"`
}
type powerSupplierOption struct {
	MemberInfo *types.NodeAlias `json:"memberInfo"`
}
type receiverOption struct {
	MemberInfo *types.NodeAlias   `json:"memberInfo"`
	Providers  []*types.NodeAlias `json:"providers"`
}

type taskPeerInfo struct {
	// Used to connect when running task, internal network resuorce of org.
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type PrepareMsg struct {
	ProposalID  common.Hash   `json:"proposalId"`
	TaskOption  *taskOption   `json:"taskOption"`
	CreateAt    uint64        `json:"createAt"`
	Sign        types.MsgSign `json:"sign"`
	messageHash atomic.Value  `rlp:"-"`
}

func (msg *PrepareMsg) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}
func (msg *PrepareMsg) MsgHash() common.Hash {
	if mhash := msg.messageHash.Load(); mhash != nil {
		return mhash.(common.Hash)
	}
	v := utils.BuildHash(PrepareProposalMsg, utils.MergeBytes(msg.ProposalID.Bytes(),
		msg.TaskOption.Hash().Bytes(), msg.Sign.Bytes(), bytesutil.Uint64ToBytes(msg.CreateAt)))
	msg.messageHash.Store(v)
	return v
}

type PrepareVote struct {
	ProposalID  common.Hash      `json:"proposalId"`
	Role        types.TaskRole   `json:"role"` // The role information of the current recipient of the task
	Owner       *types.NodeAlias `json:"owner"`
	VoteOption  types.VoteOption `json:"voteOption"`
	PeerInfo    *taskPeerInfo    `json:"peerInfo"`
	CreateAt    uint64           `json:"createAt"`
	Sign        types.MsgSign    `json:"sign"`
	messageHash atomic.Value     `rlp:"-"`
}

func (msg *PrepareVote) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

func (msg *PrepareVote) MsgHash() common.Hash {
	if mhash := msg.messageHash.Load(); mhash != nil {
		return mhash.(common.Hash)
	}
	v := utils.BuildHash(PrepareVoteMsg, utils.MergeBytes(msg.ProposalID.Bytes(), /*msg.VoteNodeID.Bytes(), */ // TODO 编码 NodeAlias
		msg.VoteOption.Bytes(), msg.Sign.Bytes(), bytesutil.Uint64ToBytes(msg.CreateAt)))
	msg.messageHash.Store(v)
	return v
}

type ProposalStatePeriod uint32

const (
	PeriodUnknown ProposalStatePeriod = 0
	PeriodPrepare ProposalStatePeriod = 1
	PeriodConfirm ProposalStatePeriod = 2
	PeriodCommit  ProposalStatePeriod = 3

	PrepareMsgVotingTimeout = 3 * time.Second
	ConfirmMsgVotingTimeout = 1 * time.Second
)

type ConfirmEpoch uint64

func (c ConfirmEpoch) Uint64() uint64 { return uint64(c) }

const (
	ConfirmEpochUnknown ConfirmEpoch = 0
	ConfirmEpochFirst   ConfirmEpoch = 1
	ConfirmEpochSecond  ConfirmEpoch = 2

	SendTaskDir ProposalTaskDir = 0x00
	RecvTaskDir ProposalTaskDir = 0x01
)

var (
	// during 60s, if the proposal haven't been done, kill it
	ProposalDeadlineDuration = uint64(60 * (time.Second.Nanoseconds()))
)

type ProposalTaskDir uint8
type ProposalState struct {
	ProposalId         common.Hash
	TaskDir            ProposalTaskDir
	TaskId             string
	PeriodNum          ProposalStatePeriod
	PrePeriodStartTime uint64
	PeriodStartTime    uint64 // the timestemp
	ConfirmEpoch       ConfirmEpoch
	// Clear `ProposalState` ,
	// when the current time is greater than the `DeadlineDuration` createAt of proposalState
	DeadlineDuration uint64
	CreateAt         uint64
}

var EmptyProposalState = new(ProposalState)

func NewProposalState(proposalId common.Hash, taskId string, TaskDir ProposalTaskDir, startTime uint64) *ProposalState {
	return &ProposalState{
		ProposalId:       proposalId,
		TaskId:           taskId,
		TaskDir:          TaskDir,
		PeriodNum:        PeriodPrepare,
		PeriodStartTime:  startTime,
		ConfirmEpoch:     ConfirmEpochUnknown,
		DeadlineDuration: ProposalDeadlineDuration,
		CreateAt:         uint64(time.Now().UnixNano()),
	}
}

func (pstate *ProposalState) GetProposalId() common.Hash         { return pstate.ProposalId }
func (pstate *ProposalState) CurrPeriodNum() ProposalStatePeriod { return pstate.PeriodNum }

//func (pstate *ProposalState) CurrPeriodDuration() uint64 {
//	return pstate.PeriodStartTime - pstate.PeriodEndTime
//}
func (pstate *ProposalState) IsPreparePeriod() bool    { return pstate.PeriodNum == PeriodPrepare }
func (pstate *ProposalState) IsConfirmPeriod() bool    { return pstate.PeriodNum == PeriodConfirm }
func (pstate *ProposalState) IsCommitPeriod() bool     { return pstate.PeriodNum == PeriodCommit }
func (pstate *ProposalState) IsNotPreparePeriod() bool { return !pstate.IsPreparePeriod() }
func (pstate *ProposalState) IsNotConfirmPeriod() bool { return !pstate.IsConfirmPeriod() }
func (pstate *ProposalState) IsNotCommitPeriod() bool  { return !pstate.IsCommitPeriod() }
func (pstate *ProposalState) IsFirstConfirmEpoch() bool {
	if pstate.ConfirmEpoch == ConfirmEpochFirst {
		return true
	}
	return false
}
func (pstate *ProposalState) IsSecondConfirmEpoch() bool {
	if pstate.ConfirmEpoch == ConfirmEpochSecond {
		return true
	}
	return false
}
func (pstate *ProposalState) IsEmpty() bool {
	if pstate == EmptyProposalState {
		return true
	}
	return false
}
func (pstate *ProposalState) GetConfirmEpoch() ConfirmEpoch { return pstate.ConfirmEpoch }
func (pstate *ProposalState) IsPrepareTimeout() bool {
	now := uint64(time.Now().UnixNano())
	duration := uint64(PrepareMsgVotingTimeout.Nanoseconds())
	// Due to the time boundary problem, the value `==`
	if pstate.PeriodNum == PeriodPrepare && (pstate.PeriodStartTime-now) >= duration {
		return true
	}
	return false
}
func (pstate *ProposalState) IsFirstConfirmEpochTimeout() bool {

	now := uint64(time.Now().UnixNano())
	duration := uint64(ConfirmMsgVotingTimeout.Nanoseconds())

	if pstate.IsFirstConfirmEpoch() && (pstate.PeriodStartTime-now) >= duration {
		return true
	}
	return false
}
func (pstate *ProposalState) IsSecondConfirmEpochTimeout() bool {

	now := uint64(time.Now().UnixNano())
	duration := uint64(ConfirmMsgVotingTimeout.Nanoseconds())

	if pstate.IsSecondConfirmEpoch() && (pstate.PeriodStartTime-now) >= duration {
		return true
	}
	return false
}

func (pstate *ProposalState) ChangeToConfirm(startTime uint64) {
	pstate.PrePeriodStartTime = pstate.PeriodStartTime
	pstate.PeriodStartTime = startTime
	pstate.PeriodNum = PeriodConfirm
	pstate.ConfirmEpoch = ConfirmEpochFirst
}
func (pstate *ProposalState) ChangeToConfirmSecondEpoch(startTime uint64) {
	pstate.PrePeriodStartTime = pstate.PeriodStartTime
	pstate.PeriodStartTime = startTime
	pstate.ConfirmEpoch = ConfirmEpochSecond
}
func (pstate *ProposalState) ChangeToCommit(startTime uint64) {
	pstate.PrePeriodStartTime = pstate.PeriodStartTime
	pstate.PeriodStartTime = startTime
	pstate.PeriodNum = PeriodCommit
	pstate.ConfirmEpoch = ConfirmEpochUnknown
}
