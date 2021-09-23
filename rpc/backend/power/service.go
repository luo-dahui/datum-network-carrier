package power

import (
	"context"
	"errors"
	pb "github.com/RosettaFlow/Carrier-Go/lib/api"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	"github.com/RosettaFlow/Carrier-Go/rpc/backend"
	"github.com/RosettaFlow/Carrier-Go/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

func (svr *Server) GetGlobalPowerDetailList(ctx context.Context, req *emptypb.Empty) (*pb.GetGlobalPowerDetailListResponse, error) {
	powerList, err := svr.B.GetGlobalPowerDetailList()
	if nil != err {
		log.WithError(err).Error("RPC-API:GetGlobalPowerDetailList failed")
		return nil, ErrGetTotalPowerList
	}
	log.Debugf("RPC-API:GetGlobalPowerDetailList succeed, powerList: {%d}, json: %s", len(powerList), utilGetGlobalPowerDetailResponseArrString(powerList))
	return &pb.GetGlobalPowerDetailListResponse{
		Status: 0,
		Msg: backend.OK,
		PowerList: powerList,
	}, nil
}

func (svr *Server) GetLocalPowerDetailList(ctx context.Context, req *emptypb.Empty) (*pb.GetLocalPowerDetailListResponse, error) {
	powerList, err := svr.B.GetLocalPowerDetailList()
	if nil != err {
		log.WithError(err).Error("RPC-API:GetLocalPowerDetailList failed")
		return nil, ErrGetSinglePowerList
	}
	log.Debugf("RPC-API:GetLocalPowerDetailList succeed, powerList: {%d}, json: %s", len(powerList), utilGetLocalPowerDetailResponseArrString(powerList))
	return &pb.GetLocalPowerDetailListResponse{
		Status: 0,
		Msg: backend.OK,
		PowerList: powerList,
	}, nil
}
func utilGetGlobalPowerDetailResponseArrString(resp []*pb.GetGlobalPowerDetailResponse) string {
	arr := make([]string, len(resp))
	for i, u := range resp {
		arr[i] = u.String()
	}
	if len(arr) != 0 {
		return "[" +  strings.Join(arr, ",") + "]"
	}
	return "[]"
}
func utilGetLocalPowerDetailResponseArrString(resp []*pb.GetLocalPowerDetailResponse) string {
	arr := make([]string, len(resp))
	for i, u := range resp {
		arr[i] = u.String()
	}
	if len(arr) != 0 {
		return "[" +  strings.Join(arr, ",") + "]"
	}
	return "[]"
}

func (svr *Server) PublishPower(ctx context.Context, req *pb.PublishPowerRequest) (*pb.PublishPowerResponse, error) {
	if req == nil {
		return nil, errors.New("required owner")
	}

	_, err := svr.B.GetNodeIdentity()
	if nil != err {
		log.WithError(err).Errorf("RPC-API:PublishPower failed, query local identity failed, can not publish power")
		return nil, ErrSendPowerMsg
	}

	powerMsg := types.NewPowerMessageFromRequest(req)
	powerId := powerMsg.GenPowerId()


	err = svr.B.SendMsg(powerMsg)
	if nil != err {
		log.WithError(err).Errorf("RPC-API:PublishPower failed, jobNodeId: {%s}, powerId: {%s}", req.GetJobNodeId(), powerId)
		return nil, ErrSendPowerMsg
	}
	log.Debugf("RPC-API:PublishPower succeed, jobNodeId: {%s}, powerId: {%s}", req.GetJobNodeId(), powerId)
	return &pb.PublishPowerResponse{
		Status:  0,
		Msg:     backend.OK,
		PowerId: powerId,
	}, nil
}

func (svr *Server) RevokePower(ctx context.Context, req *pb.RevokePowerRequest) (*apicommonpb.SimpleResponse, error) {
	if req == nil {
		return nil, errors.New("required owner")
	}
	if req.PowerId == "" {
		return nil, errors.New("required powerId")
	}

	_, err := svr.B.GetNodeIdentity()
	if nil != err {
		log.WithError(err).Errorf("RPC-API:RevokePower failed, query local identity failed, can not revoke power")
		return nil, ErrSendPowerRevokeMsg
	}

	powerRevokeMsg := types.NewPowerRevokeMessageFromRequest(req)

	err = svr.B.SendMsg(powerRevokeMsg)
	if nil != err {
		log.WithError(err).Errorf("RPC-API:RevokePower failed, powerId: {%s}", req.PowerId)
		return nil, ErrSendPowerRevokeMsg
	}
	log.Debugf("RPC-API:RevokePower succeed, powerId: {%s}", req.PowerId)
	return &apicommonpb.SimpleResponse{
		Status: 0,
		Msg:    backend.OK,
	}, nil
}
