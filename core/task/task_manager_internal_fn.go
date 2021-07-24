package task

import (
	"encoding/json"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common/timeutils"
	ev "github.com/RosettaFlow/Carrier-Go/core/evengine"
	pb "github.com/RosettaFlow/Carrier-Go/lib/consensus/twopc"
	"github.com/RosettaFlow/Carrier-Go/lib/fighter/common"
	libTypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"github.com/RosettaFlow/Carrier-Go/types"
	"strconv"
)



func (m *Manager) driveTaskForExecute (task *types.DoneScheduleTaskChWrap) error {

	//switch task.SelfTaskRole {
	//case types.TaskOnwer:
	//	//dataNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_DATANODE)
	//	//if nil != err {
	//	//	return err
	//	//}
	//	//ip := string(task.Task.Resources.OwnerPeerInfo.Ip)
	//	//port := string(task.Task.Resources.OwnerPeerInfo.Port)
	//	//
	//	//var dataNodeId string
	//	//for _, dataNode := range dataNodeList {
	//	//	if ip == dataNode.ExternalIp && port == dataNode.ExternalPort {
	//	//		dataNodeId = dataNode.Id
	//	//		break
	//	//	}
	//	//}
	//	return m.executeTaskOnDataNode(task)
	//
	//case types.DataSupplier:
	//	//dataNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_DATANODE)
	//	//if nil != err {
	//	//	return err
	//	//}
	//	//
	//	//tmp := make(map[string]struct{}, len(task.Task.Resources.DataSupplierPeerInfoList))
	//	//for _, dataNode := range task.Task.Resources.DataSupplierPeerInfoList {
	//	//	tmp[string(dataNode.Ip) + "_" + string(dataNode.Port)] = struct{}{}
	//	//}
	//
	//	// task 中的 dataNode 可能是 单个组织的 多个 dataNode,
	//	// 逐个下发 task
	//	for _, dataNode := range dataNodeList {
	//		if _, ok := tmp[dataNode.ExternalIp + "_" + dataNode.ExternalPort]; ok {
	//			if err := m.executeTaskOnDataNode(task); nil != err {
	//				log.Errorf("Failed to execute task on dataNode: %s, %s", dataNode.Id, err)
	//				return err
	//			}
	//		}
	//	}
	//
	//case types.PowerSupplier:
	//	jobNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_JOBNODE)
	//	if nil != err {
	//		return err
	//	}
	//
	//	tmp := make(map[string]struct{}, len(task.Task.Resources.PowerSupplierPeerInfoList))
	//	for _, jobNode := range task.Task.Resources.PowerSupplierPeerInfoList {
	//		tmp[string(jobNode.Ip) + "_" + string(jobNode.Port)] = struct{}{}
	//	}
	//
	//	// task 中的 jobNode 可能是 单个组织的 多个 jobNode,
	//	// 逐个下发 task
	//	for _, jobNode := range jobNodeList {
	//		if _, ok := tmp[jobNode.ExternalIp + "_" + jobNode.ExternalPort]; ok {
	//			if err := m.executeTaskOnJobNode(jobNode.Id, task); nil != err {
	//				log.Errorf("Failed to execute task on jobNode: %s, %s", jobNode.Id, err)
	//				return err
	//			}
	//		}
	//	}
	//
	//case types.ResultSupplier:
	//	dataNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_DATANODE)
	//	if nil != err {
	//		return err
	//	}
	//
	//	tmp := make(map[string]struct{}, len(task.Task.Resources.ResultReceiverPeerInfoList))
	//	for _, receiveNode := range task.Task.Resources.ResultReceiverPeerInfoList {
	//		tmp[string(receiveNode.Ip) + "_" + string(receiveNode.Port)] = struct{}{}
	//	}
	//
	//	// task 中的 dataNode 可能是 单个组织的 多个 dataNode,
	//	// 逐个下发 task
	//	for _, dataNode := range dataNodeList {
	//		if _, ok := tmp[dataNode.ExternalIp + "_" + dataNode.ExternalPort]; ok {
	//			if err := m.executeTaskOnDataNode(task); nil != err {
	//				log.Errorf("Failed to execute task on receiveNode: %s, %s", dataNode.Id, err)
	//				return err
	//			}
	//		}
	//	}
	//}

	switch task.SelfTaskRole {
	case types.TaskOnwer, types.DataSupplier, types.ResultSupplier :
		return m.executeTaskOnDataNode(task)
	case types.PowerSupplier:
		return m.executeTaskOnJobNode(task)
	default:
		return fmt.Errorf("Faided to driveTaskForExecute(), Unknown task role, taskId: {%s}, taskRole: {%s}", task.Task.SchedTask.TaskId(), task.SelfTaskRole.String())
	}
	return nil
}

func (m *Manager) executeTaskOnDataNode(task *types.DoneScheduleTaskChWrap) error {

	dataNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_DATANODE)
	if nil != err {
		return err
	}

	var find bool
	var ip string
	var port string
	if string(task.Task.Resources.OwnerPeerInfo.PartyId) == task.SelfIdentity.PartyId {
			ip = string(task.Task.Resources.OwnerPeerInfo.Ip)
			port = string(task.Task.Resources.OwnerPeerInfo.Port)
			find = true
	}
	if !find {
		for _, resource := range task.Task.Resources.DataSupplierPeerInfoList {
			if string(resource.PartyId) == task.SelfIdentity.PartyId {
				ip = string(resource.Ip)
				port = string(resource.Port)
				find = true
				break
			}
		}
	}

	if !find {
		for _, resource := range task.Task.Resources.ResultReceiverPeerInfoList {
			if string(resource.PartyId) == task.SelfIdentity.PartyId {
				ip = string(resource.Ip)
				port = string(resource.Port)
				find = true
				break
			}
		}
	}

	if !find {
		return fmt.Errorf("Failed to call executeTaskOnDataNode(), not find the self resource, taskId: {%s}, self.IdentityId: {%s}, self.partyId: {%s}",
			task.Task.SchedTask.TaskId(), task.SelfIdentity.Identity, task.SelfIdentity.PartyId)
	}


	var dataNodeId string
	for _, dataNode := range dataNodeList {
		if dataNode.ExternalIp == ip && dataNode.ExternalPort == port {
			dataNodeId = dataNode.Id
			break
		}
	}

	// clinet *grpclient.DataNodeClient,
	client, isconn := m.resourceClientSet.QueryDataNodeClient(dataNodeId)
	if !isconn {
		if err := client.Reconnect(); nil != err {
			log.Errorf("Failed to connect internal data node, taskId: {%s}, dataNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
				task.Task.SchedTask.TaskId(), dataNodeId, ip, port, err)
			return err
		}
	}
	req, err := m.makeTaskReadyGoReq(task)
	if nil != err {
		log.Errorf("Falied to make taskReadyGoReq, taskId: {%s}, dataNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
			task.Task.SchedTask.TaskId(), dataNodeId, ip, port, err)
		return err
	}

	resp, err := client.HandleTaskReadyGo(req)
	if nil != err {
		log.Errorf("Falied to publish schedTask to `data-Fighter` node to executing, taskId: {%s}, dataNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
			task.Task.SchedTask.TaskId(), dataNodeId, ip, port, err)
		return err
	}
	if !resp.Ok {
		log.Errorf("Falied to publish schedTask to `data-Fighter` node to executing, taskId: {%s}, dataNodeId: {%s}, ip: {%s}, port: {%s}",
			task.Task.SchedTask.TaskId(), dataNodeId, ip, port)
		return nil
	}

	task.Task.SchedTask.TaskData().StartAt = uint64(timeutils.UnixMsec())
	m.addRunningTaskCache(task)

	log.Infof("Success to publish schedTask to `data-Fighter` node to executing,  taskId: {%s}, dataNodeId: {%s}, ip: {%s}, port: {%s}",
		task.Task.SchedTask.TaskId(), dataNodeId, ip, port)
	return nil
}

func (m *Manager) executeTaskOnJobNode(task *types.DoneScheduleTaskChWrap) error {

	jobNodeList, err := m.dataCenter.GetRegisterNodeList(types.PREFIX_TYPE_JOBNODE)
	if nil != err {
		return err
	}

	var find bool
	var ip string
	var port string
	for _, resource := range task.Task.Resources.PowerSupplierPeerInfoList {
		if string(resource.PartyId) == task.SelfIdentity.PartyId {
			ip = string(resource.Ip)
			port = string(resource.Port)
			find = true
			break
		}
	}

	if !find {
		return fmt.Errorf("Failed to call executeTaskOnJobNode(), not find the self resource, taskId: {%s}, self.IdentityId: {%s}, self.partyId: {%s}",
			task.Task.SchedTask.TaskId(), task.SelfIdentity.Identity, task.SelfIdentity.PartyId)
	}


	var jobNodeId string
	for _, jobNode := range jobNodeList {
		if jobNode.ExternalIp == ip && jobNode.ExternalPort == port {
			jobNodeId = jobNode.Id
			break
		}
	}

	// clinet *grpclient.DataNodeClient,
	client, isconn := m.resourceClientSet.QueryJobNodeClient(jobNodeId)
	if !isconn {
		if err := client.Reconnect(); nil != err {
			log.Errorf("Failed to connect internal job node, taskId: {%s}, jobNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
				task.Task.SchedTask.TaskId(), jobNodeId, ip, port, err)
			return err
		}
	}
	req, err := m.makeTaskReadyGoReq(task)
	if nil != err {
		log.Errorf("Falied to make taskReadyGoReq, taskId: {%s}, jobNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
			task.Task.SchedTask.TaskId(), jobNodeId, ip, port, err)
		return err
	}

	resp, err := client.HandleTaskReadyGo(req)
	if nil != err {
		log.Errorf("Falied to publish schedTask to `job-Fighter` node to executing, taskId: {%s}, jobNodeId: {%s}, ip: {%s}, port: {%s}, err: {%}",
			task.Task.SchedTask.TaskId(), jobNodeId, ip, port, err)
		return err
	}
	if !resp.Ok {
		log.Errorf("Falied to publish schedTask to `job-Fighter` node to executing, taskId: {%s}, jobNodeId: {%s}, ip: {%s}, port: {%s}",
			task.Task.SchedTask.TaskId(), jobNodeId, ip, port)
		return nil
	}

	task.Task.SchedTask.TaskData().StartAt = uint64(timeutils.UnixMsec())
	m.addRunningTaskCache(task)

	log.Infof("Success to publish schedTask to `job-Fighter` node to executing, taskId: {%s}, jobNodeId: {%s}, ip: {%s}, port: {%s}",
		task.Task.SchedTask.TaskId(), jobNodeId, ip, port)
	return nil
}


func (m *Manager) pulishFinishedTaskToDataCenter(taskId, taskState string) {
	taskWrap, ok := m.queryRunningTaskCacheOk(taskId)
	if !ok {
		return
	}

	eventList, err := m.dataCenter.GetTaskEventList(taskWrap.Task.SchedTask.TaskId())
	if nil != err {
		log.Error("Failed to Query all task event list for sending datacenter", "taskId", taskWrap.Task.SchedTask.TaskId)
		return
	}
	if err := m.dataCenter.InsertTask(m.convertScheduleTaskToTask(taskWrap.Task.SchedTask, eventList, taskState)); nil != err {
		log.Error("Failed to save task to datacenter", "taskId", taskWrap.Task.SchedTask.TaskId)
		return
	}

	// 发送到 dataCenter 成功后 ...
	close(taskWrap.ResultCh)
	// clean local task cache
	m.removeRunningTaskCache(taskId)
	// 解锁 本地 资源缓存
	m.resourceMng.UnLockLocalResourceWithTask(taskId)
	// 清掉 本地任务
	m.dataCenter.RemoveLocalTask(taskId)
	// 清掉 本地事件
	m.dataCenter.CleanTaskEventList(taskId)
}
func (m *Manager) sendTaskResultMsgToConsensus(taskId string) {

	taskWrap, ok := m.queryRunningTaskCacheOk(taskId)
	if !ok {
		log.Errorf( "Not found taskwrap, taskId: %s", taskId)
		return
	}

	taskResultMsg := m.makeTaskResult(taskWrap)
	if nil != taskResultMsg {
		taskWrap.ResultCh <- taskResultMsg
	}
	close(taskWrap.ResultCh)
	// clean local task cache
	m.removeRunningTaskCache(taskWrap.Task.SchedTask.TaskId())
}

func (m *Manager) sendTaskMsgsToScheduler(msgs types.TaskMsgs) {
	m.localTaskMsgCh <- msgs
}
func (m *Manager) sendTaskEvent(event *types.TaskEventInfo){
	m.eventCh <- event
}



func (m *Manager) storeErrTaskMsg(msg *types.TaskMsg, events []*libTypes.EventData, reason string) error {
	msg.Data.TaskData().EventDataList = events
	msg.Data.TaskData().EventCount = uint32(len(events))
	msg.Data.TaskData().Reason = reason
	msg.Data.TaskData().EndAt = uint64(timeutils.UnixMsec())
	return m.dataCenter.InsertTask(msg.Data)
}


func (m *Manager) convertScheduleTaskToTask(task *types.Task, eventList []*types.TaskEventInfo, state string)  *types.Task {
	task.TaskData().EventDataList = types.ConvertTaskEventArrToDataCenter(eventList)
	task.TaskData().EventCount = uint32(len(eventList))
	task.TaskData().EndAt = uint64(timeutils.UnixMsec())
	task.TaskData().State = state
	return task
}

func (m *Manager) makeTaskReadyGoReq(task *types.DoneScheduleTaskChWrap) (*common.TaskReadyGoReq, error) {

	ownerPort := string(task.Task.Resources.OwnerPeerInfo.Port)
	port, err := strconv.Atoi(ownerPort)
	if nil != err {
		return nil, err
	}

	var dataPartyArr []string
	var powerPartyArr []string
	var receiverPartyArr []string

	peerList :=  []*common.TaskReadyGoReq_Peer{
		&common.TaskReadyGoReq_Peer {
			Ip: string(task.Task.Resources.OwnerPeerInfo.Ip),
			Port: int32(port),
			PartyId:  string(task.Task.Resources.OwnerPeerInfo.PartyId),
		},
	}
	dataPartyArr = append(dataPartyArr, string(task.Task.Resources.OwnerPeerInfo.PartyId))

	for _, dataSupplier := range task.Task.Resources.DataSupplierPeerInfoList {
		portStr := string(dataSupplier.Port)
		port, err := strconv.Atoi(portStr)
		if nil != err {
			return nil, err
		}
		peerList = append(peerList, &common.TaskReadyGoReq_Peer {
			Ip: string(dataSupplier.Ip),
			Port: int32(port),
			PartyId:  string(dataSupplier.PartyId),
		})
		dataPartyArr = append(dataPartyArr, string(dataSupplier.PartyId))
	}

	for _, powerSupplier := range task.Task.Resources.PowerSupplierPeerInfoList {
		portStr := string(powerSupplier.Port)
		port, err := strconv.Atoi(portStr)
		if nil != err {
			return nil, err
		}
		peerList = append(peerList, &common.TaskReadyGoReq_Peer {
			Ip: string(powerSupplier.Ip),
			Port: int32(port),
			PartyId:  string(powerSupplier.PartyId),
		})

		powerPartyArr = append(powerPartyArr, string(powerSupplier.PartyId))
	}

	for _, receiver := range task.Task.Resources.ResultReceiverPeerInfoList {
		portStr := string(receiver.Port)
		port, err := strconv.Atoi(portStr)
		if nil != err {
			return nil, err
		}
		peerList = append(peerList, &common.TaskReadyGoReq_Peer {
			Ip: string(receiver.Ip),
			Port: int32(port),
			PartyId:  string(receiver.PartyId),
		})

		receiverPartyArr = append(receiverPartyArr, string(receiver.PartyId))
	}

	contractExtraParams, err := m.makeContractParams(task)
	if nil != err {
		return nil, err
	}

	return &common.TaskReadyGoReq{
		TaskId: task.Task.SchedTask.TaskId(),
		ContractId: task.Task.SchedTask.TaskData().CalculateContractCode,
		//DataId: "",
		PartyId: task.SelfIdentity.PartyId,
		//EnvId: "",
		Peers: peerList,
		ContractCfg: contractExtraParams,
		DataParty: dataPartyArr,
		ComputationParty: powerPartyArr,
		ResultParty: receiverPartyArr,
	}, nil
}

func  (m *Manager) makeContractParams (task *types.DoneScheduleTaskChWrap) (string, error) {



	partyId := task.SelfIdentity.PartyId


	var find bool
	var filePath string
	var columnNameList []string

	if task.SelfTaskRole  == types.TaskOnwer || task.SelfTaskRole == types.DataSupplier {
		for _, dataSupplier := range task.Task.SchedTask.TaskData().MetadataSupplier {
			if partyId == dataSupplier.Organization.PartyId {

				metaData, err := m.dataCenter.GetMetadataByDataId(dataSupplier.MetaId)
				if nil != err {
					return "", err
				}
				filePath = metaData.MetadataData().FilePath
				for _, col := range dataSupplier.ColumnList {
					columnNameList = append(columnNameList, col.Cname)
				}
				find = true
				break
			}
		}
	}

	if !find {
		return "", fmt.Errorf("can not find the dataSupplier, taskId: {%s}, self.IdentityId: {%s}, seld.PartyId: {%s}",
			task.Task.SchedTask.TaskId(), task.SelfIdentity.Identity, task.SelfIdentity.PartyId)
	}
	// 目前 默认只会用一列, 后面再拓展 ..
	req := &types.FighterTaskReadyGoReqContractCfg{
		PartyId: "p2",
		DataParty: struct {
			InputFile  string    `json:"input_file"`
			IdColumnName string    `json:"id_column_name"`
		}{
			InputFile: filePath,
			IdColumnName: columnNameList[0],  // 目前 默认只会用一列, 后面再拓展 ..
		},
	}

	var dynamicParameter map[string]interface{}
	if err := json.Unmarshal([]byte(task.Task.SchedTask.TaskData().ContractExtraParams), &dynamicParameter); nil != err {
		return "", fmt.Errorf("can not json Unmarshal the `ContractExtraParams` of task, taskId: {%s}, self.IdentityId: {%s}, seld.PartyId: {%s}",
			task.Task.SchedTask.TaskId(), task.SelfIdentity.Identity, task.SelfIdentity.PartyId)
	}
	req.DynamicParameter = dynamicParameter

	b, err := json.Marshal(req)
	if nil != err {
		return "", fmt.Errorf("can not json Marshal the `FighterTaskReadyGoReqContractCfg`, taskId: {%s}, self.IdentityId: {%s}, seld.PartyId: {%s}",
			task.Task.SchedTask.TaskId(), task.SelfIdentity.Identity, task.SelfIdentity.PartyId)
	}
	return  string(b), nil
}

func (m *Manager) addRunningTaskCache(task *types.DoneScheduleTaskChWrap) {
	m.runningTaskCacheLock.Lock()
	m.runningTaskCache[task.Task.SchedTask.TaskId()] = task
	m.runningTaskCacheLock.Unlock()
}

func (m *Manager) removeRunningTaskCache(taskId string) {
	m.runningTaskCacheLock.Lock()
	delete(m.runningTaskCache, taskId)
	m.runningTaskCacheLock.Unlock()
}

func (m *Manager) queryRunningTaskCacheOk(taskId string) (*types.DoneScheduleTaskChWrap, bool) {
	task, ok := m.runningTaskCache[taskId]
	return task, ok
}

func (m *Manager) queryRunningTaskCache(taskId string) *types.DoneScheduleTaskChWrap {
	task, _ := m.queryRunningTaskCacheOk(taskId)
	return task
}

func (m *Manager) makeTaskResult (taskWrap *types.DoneScheduleTaskChWrap)  *types.TaskResultMsgWrap {

	if taskWrap.Task.TaskDir  ==  types.SendTaskDir || types.TaskOnwer == taskWrap.SelfTaskRole {
		log.Errorf("send task OR task owner can not make TaskResult Msg")
		return nil
	}


	eventList, err := m.dataCenter.GetTaskEventList(taskWrap.Task.SchedTask.TaskId())
	if nil != err {
		log.Errorf("Failed to make TaskResultMsg with query task eventList, taskId {%s}, err {%s}", taskWrap.Task.SchedTask.TaskId(), err)
		return nil
	}
	return &types.TaskResultMsgWrap{
		TaskResultMsg: &pb.TaskResultMsg{
			ProposalId: taskWrap.ProposalId.Bytes(),
			TaskRole: taskWrap.SelfTaskRole.Bytes(),
			TaskId: []byte(taskWrap.Task.SchedTask.TaskId()),
			Owner: &pb.TaskOrganizationIdentityInfo{
				PartyId: []byte(taskWrap.SelfIdentity.PartyId),
				Name: []byte(taskWrap.SelfIdentity.NodeName),
				NodeId: []byte(taskWrap.SelfIdentity.NodeId),
				IdentityId: []byte(taskWrap.SelfIdentity.Identity),
			},
			TaskEventList: types.ConvertTaskEventArr(eventList),
			CreateAt: uint64(timeutils.UnixMsec()),
			Sign: nil,
		},
	}
}

func (m *Manager) handleEvent(event *types.TaskEventInfo) error {
	eventType := event.Type
	if len(eventType) != ev.EventTypeCharLen {
		return ev.IncEventType
	}
	// TODO need to validate the task that have been processing ? Maybe~
	if event.Type == ev.TaskExecuteSucceedEOF.Type || event.Type == ev.TaskExecuteFailedEOF.Type {
		if task, ok := m.queryRunningTaskCacheOk(event.TaskId); ok {

			// 先 缓存下 最终休止符 event
			m.dataCenter.StoreTaskEvent(event)

			if task.Task.TaskDir == types.RecvTaskDir {
				// 因为是 task 参与者, 所以需要构造 taskResult 发送给 task 发起者..  (里面有解锁本地资源 ...)
				m.sendTaskResultMsgToConsensus(event.TaskId)

			} else {
				//  如果是 自己的task, 认为任务终止 ... 发送到 dataCenter (里面有解锁本地资源 ...)
				if event.Type == ev.TaskExecuteSucceedEOF.Type {
					m.pulishFinishedTaskToDataCenter(event.TaskId, types.TaskStateSuccess.String())
				} else {
					m.pulishFinishedTaskToDataCenter(event.TaskId, types.TaskStateFailed.String())
				}

			}
		}
		return nil
	} else {

		// 不是休止符 event, 任务还在继续, 保存 event
		return m.dataCenter.StoreTaskEvent(event)
	}
}
func (m *Manager) handleDoneScheduleTask(taskId string) {

	log.Debugf("Start handle DoneScheduleTask, taskId: {%s}", taskId)

	task, ok := m.queryRunningTaskCacheOk(taskId)
	if !ok {
		log.Debugf("Failed to start handle DoneScheduleTask, not found local task cache, taskId: {%s}", taskId)
		return
	}

	switch task.SelfTaskRole {
	case types.TaskOnwer:
		switch task.Task.TaskState {
		case types.TaskStateFailed, types.TaskStateSuccess:

			// 发起方直接 往 dataCenter 发送数据 (里面有解锁 本地资源 ...)
			m.pulishFinishedTaskToDataCenter(taskId, task.Task.TaskState.String())

		case types.TaskStateRunning:

			if err := m.driveTaskForExecute(task); nil != err {
				log.Errorf("Failed to execute task on taskOnwer node, taskId: %s, %s", task.Task.SchedTask.TaskId, err)
				event := m.eventEngine.GenerateEvent(ev.TaskFailed.Type,
					task.Task.SchedTask.TaskId(), task.Task.SchedTask.TaskData().Identity, fmt.Sprintf("failed to execute task"))
				// 因为是 自己的任务, 所以直接将 task  和 event list  发给 dataCenter  (里面有解锁 本地资源 ...)
				m.dataCenter.StoreTaskEvent(event)
				m.pulishFinishedTaskToDataCenter(taskId, types.TaskStateFailed.String())  //
			}
			// TODO 而执行最终[成功]的 根据 Fighter 上报的 event 在 handleEvent() 里面处理
		default:
			log.Error("Failed to handle unknown task", "taskId", task.Task.SchedTask.TaskId)
		}
	//case types.DataSupplier:
	//case types.PowerSupplier:
	//case types.ResultSupplier:
	default:
		switch task.Task.TaskState {
		case types.TaskStateFailed, types.TaskStateSuccess:
			// 因为是 task 参与者, 所以需要构造 taskResult 发送给 task 发起者..  (里面有解锁 本地资源 ...)
			m.sendTaskResultMsgToConsensus(taskId)
		case types.TaskStateRunning:

			if err := m.driveTaskForExecute(task); nil != err {
				log.Errorf("Failed to execute task on taskOnwer node, taskId: %s, %s", task.Task.SchedTask.TaskId, err)
				identityId, _ := m.dataCenter.GetIdentityId()
				event := m.eventEngine.GenerateEvent(ev.TaskFailed.Type,
					task.Task.SchedTask.TaskId(), identityId, fmt.Sprintf("failed to execute task"))

				// 因为是 task 参与者, 所以需要构造 taskResult 发送给 task 发起者.. (里面有解锁 本地资源 ...)
				m.dataCenter.StoreTaskEvent(event)
				m.sendTaskResultMsgToConsensus(taskId)
			}
		default:
			log.Error("Failed to handle unknown task", "taskId", task.Task.SchedTask.TaskId)
		}
	}
}