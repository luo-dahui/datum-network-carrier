// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lib/types/taskdata.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type TaskData struct {
	Identity   string `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity" xml:"identity"`
	NodeId     string `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeid" xml:"nodeid"`
	NodeName   string `protobuf:"bytes,3,opt,name=nodeName,proto3" json:"nodename" xml:"nodename"`
	DataId     string `protobuf:"bytes,4,opt,name=dataId,proto3" json:"dataid" xml:"dataid"`
	DataStatus string `protobuf:"bytes,5,opt,name=dataStatus,proto3" json:"datastatus" xml:"datastatus"`
	TaskId     string `protobuf:"bytes,6,opt,name=taskId,proto3" json:"taskid" xml:"taskid"`
	// success/failed/running/pending/waiting
	State         string      `protobuf:"bytes,7,opt,name=state,proto3" json:"state" xml:"state"`
	Reason        string      `protobuf:"bytes,8,opt,name=reason,proto3" json:"reason" xml:"reason"`
	EventCount    uint32      `protobuf:"varint,9,opt,name=eventCount,proto3" json:"eventcount" xml:"eventcount"`
	Desc          string      `protobuf:"bytes,10,opt,name=desc,proto3" json:"desc" xml:"desc"`
	PartnerList   []Partner   `protobuf:"bytes,11,rep,name=partnerList,proto3" json:"partnerlist" xml:"partnerlist"`
	EventDataList []EventData `protobuf:"bytes,12,rep,name=eventDataList,proto3" json:"eventdatalist" xml:"eventdatalist"`
}

func (m *TaskData) Reset()         { *m = TaskData{} }
func (m *TaskData) String() string { return proto.CompactTextString(m) }
func (*TaskData) ProtoMessage()    {}
func (*TaskData) Descriptor() ([]byte, []int) {
	return fileDescriptor_2293d9334aae6da1, []int{0}
}
func (m *TaskData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TaskData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TaskData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TaskData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskData.Merge(m, src)
}
func (m *TaskData) XXX_Size() int {
	return m.ProtoSize()
}
func (m *TaskData) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskData.DiscardUnknown(m)
}

var xxx_messageInfo_TaskData proto.InternalMessageInfo

type Partner struct {
	Alias    string `protobuf:"bytes,1,opt,name=alias,proto3" json:"alias" xml:"alias"`
	Identity string `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity" xml:"identity"`
	NodeId   string `protobuf:"bytes,3,opt,name=nodeId,proto3" json:"nodeid" xml:"nodeid"`
	NodeName string `protobuf:"bytes,4,opt,name=nodeName,proto3" json:"nodename" xml:"nodename"`
}

func (m *Partner) Reset()         { *m = Partner{} }
func (m *Partner) String() string { return proto.CompactTextString(m) }
func (*Partner) ProtoMessage()    {}
func (*Partner) Descriptor() ([]byte, []int) {
	return fileDescriptor_2293d9334aae6da1, []int{1}
}
func (m *Partner) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Partner) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Partner.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Partner) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Partner.Merge(m, src)
}
func (m *Partner) XXX_Size() int {
	return m.ProtoSize()
}
func (m *Partner) XXX_DiscardUnknown() {
	xxx_messageInfo_Partner.DiscardUnknown(m)
}

var xxx_messageInfo_Partner proto.InternalMessageInfo

type EventData struct {
	TaskId       string `protobuf:"bytes,1,opt,name=taskId,proto3" json:"taskid" xml:"taskid"`
	EventType    string `protobuf:"bytes,2,opt,name=eventType,proto3" json:"eventtype" xml:"eventtype"`
	EventAt      uint64 `protobuf:"varint,3,opt,name=eventAt,proto3" json:"eventat" xml:"eventat"`
	EventContent string `protobuf:"bytes,4,opt,name=eventContent,proto3" json:"eventcontent" xml:"eventcontent"`
	Identity     string `protobuf:"bytes,5,opt,name=identity,proto3" json:"identity" xml:"identity"`
}

func (m *EventData) Reset()         { *m = EventData{} }
func (m *EventData) String() string { return proto.CompactTextString(m) }
func (*EventData) ProtoMessage()    {}
func (*EventData) Descriptor() ([]byte, []int) {
	return fileDescriptor_2293d9334aae6da1, []int{2}
}
func (m *EventData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventData.Merge(m, src)
}
func (m *EventData) XXX_Size() int {
	return m.ProtoSize()
}
func (m *EventData) XXX_DiscardUnknown() {
	xxx_messageInfo_EventData.DiscardUnknown(m)
}

var xxx_messageInfo_EventData proto.InternalMessageInfo

func init() {
	proto.RegisterType((*TaskData)(nil), "types.TaskData")
	proto.RegisterType((*Partner)(nil), "types.Partner")
	proto.RegisterType((*EventData)(nil), "types.EventData")
}

func init() { proto.RegisterFile("lib/types/taskdata.proto", fileDescriptor_2293d9334aae6da1) }

var fileDescriptor_2293d9334aae6da1 = []byte{
	// 661 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0x3d, 0x6f, 0xdb, 0x3c,
	0x10, 0xc7, 0xad, 0xc4, 0x4e, 0x62, 0xda, 0xc9, 0x93, 0x47, 0x0f, 0xf0, 0x80, 0x08, 0x0a, 0xd1,
	0x60, 0x81, 0xc2, 0xe8, 0x8b, 0x05, 0xa4, 0x40, 0x87, 0x0c, 0x05, 0xea, 0xf4, 0x05, 0x29, 0xda,
	0xa2, 0x60, 0x33, 0x65, 0xa3, 0x2d, 0x36, 0x15, 0x62, 0x4b, 0x86, 0x48, 0xb7, 0xcd, 0xb7, 0xe8,
	0x47, 0xe8, 0x37, 0xe9, 0x9a, 0x31, 0x63, 0x27, 0x02, 0x89, 0xd1, 0x45, 0xa3, 0xa6, 0x8e, 0x05,
	0x8f, 0xb4, 0x24, 0x6f, 0x4d, 0x36, 0xdd, 0x4f, 0xf7, 0xbf, 0xbf, 0x8e, 0x77, 0x14, 0xc2, 0x93,
	0x78, 0x14, 0xaa, 0xf3, 0x99, 0x90, 0xa1, 0xe2, 0xf2, 0x2c, 0xe2, 0x8a, 0x0f, 0x66, 0x59, 0xaa,
	0x52, 0xbf, 0x05, 0x74, 0xef, 0x6e, 0x26, 0x66, 0xa9, 0x0c, 0x81, 0x8d, 0xe6, 0x1f, 0xc3, 0xd3,
	0xf4, 0x34, 0x85, 0x00, 0x9e, 0x6c, 0x2e, 0xfd, 0xdd, 0x42, 0x5b, 0xc7, 0x5c, 0x9e, 0x3d, 0xe7,
	0x8a, 0xfb, 0x07, 0x68, 0x2b, 0x8e, 0x44, 0xa2, 0x62, 0x75, 0x8e, 0xbd, 0x9e, 0xd7, 0x6f, 0x0f,
	0x83, 0x5c, 0x93, 0x92, 0x15, 0x9a, 0xec, 0x7c, 0x9d, 0x4e, 0x0e, 0xe8, 0x12, 0x50, 0x56, 0xbe,
	0xf3, 0xf7, 0xd1, 0x46, 0x92, 0x46, 0xe2, 0x28, 0xc2, 0x6b, 0xa0, 0xdc, 0xcb, 0x35, 0x01, 0x12,
	0x47, 0x85, 0x26, 0x5d, 0xd0, 0xd9, 0x90, 0x32, 0x97, 0x69, 0xfc, 0xcc, 0xd3, 0x3b, 0x3e, 0x15,
	0x78, 0xbd, 0xf2, 0x33, 0x2c, 0xe1, 0x53, 0x51, 0xfa, 0x2d, 0x01, 0x65, 0x65, 0xbe, 0xf1, 0x33,
	0x2d, 0x1f, 0x45, 0xb8, 0x59, 0xf9, 0x19, 0x52, 0xf3, 0xb3, 0x21, 0x65, 0x2e, 0xd3, 0x1f, 0x22,
	0x64, 0x9e, 0x3e, 0x28, 0xae, 0xe6, 0x12, 0xb7, 0x40, 0x47, 0x73, 0x4d, 0x80, 0x4a, 0xa0, 0x85,
	0x26, 0xbb, 0xa5, 0xd6, 0x22, 0xca, 0x6a, 0x2a, 0xe3, 0x6b, 0x8e, 0xfb, 0x28, 0xc2, 0x1b, 0x95,
	0xaf, 0x21, 0x35, 0x5f, 0x1b, 0x52, 0xe6, 0x32, 0xfd, 0x01, 0x6a, 0x99, 0x52, 0x02, 0x6f, 0x82,
	0x04, 0xe7, 0x9a, 0x58, 0x50, 0x68, 0xd2, 0x01, 0x05, 0x44, 0x94, 0x59, 0x6a, 0x3c, 0x32, 0xc1,
	0x65, 0x9a, 0xe0, 0xad, 0xca, 0xc3, 0x92, 0xd2, 0xc3, 0x86, 0x94, 0x39, 0x6e, 0x7a, 0x13, 0x9f,
	0x45, 0xa2, 0x0e, 0xd3, 0x79, 0xa2, 0x70, 0xbb, 0xe7, 0xf5, 0xb7, 0x6d, 0x6f, 0x40, 0xc7, 0x86,
	0x96, 0xbd, 0x55, 0x88, 0xb2, 0x9a, 0xca, 0xbf, 0x8f, 0x9a, 0x91, 0x90, 0x63, 0x8c, 0xc0, 0xf5,
	0xff, 0x5c, 0x13, 0x88, 0x0b, 0x4d, 0x90, 0x3d, 0x13, 0x21, 0xc7, 0x94, 0x01, 0xf3, 0x4f, 0x50,
	0x67, 0xc6, 0x33, 0x95, 0x88, 0xec, 0x4d, 0x2c, 0x15, 0xee, 0xf4, 0xd6, 0xfb, 0x9d, 0xfd, 0x9d,
	0x01, 0xac, 0xde, 0xe0, 0xbd, 0x7d, 0x33, 0xec, 0x5f, 0x68, 0xd2, 0xc8, 0x35, 0x59, 0xa6, 0x4e,
	0x62, 0x69, 0xbe, 0xe2, 0x5f, 0xa8, 0x56, 0x63, 0x94, 0xd5, 0x8b, 0xf9, 0x11, 0xda, 0x86, 0xaf,
	0x32, 0x4b, 0x09, 0xd5, 0xbb, 0x50, 0x7d, 0xd7, 0x55, 0x7f, 0xb1, 0x7c, 0x37, 0x7c, 0xe8, 0xea,
	0xdb, 0x74, 0x33, 0x25, 0xe7, 0xf0, 0x5f, 0xd5, 0xe7, 0x92, 0x52, 0xb6, 0x5a, 0x94, 0xfe, 0xf2,
	0xd0, 0xa6, 0xfb, 0x50, 0x33, 0x21, 0x3e, 0x89, 0xb9, 0x74, 0x6b, 0x0f, 0x13, 0x02, 0x50, 0x4e,
	0x08, 0x22, 0xca, 0x2c, 0x5d, 0xb9, 0x29, 0x6b, 0xb7, 0xbe, 0x29, 0xeb, 0xb7, 0xba, 0x29, 0xcd,
	0x9b, 0xdd, 0x14, 0xfa, 0x63, 0x0d, 0xb5, 0xcb, 0x23, 0xab, 0xed, 0xaf, 0xf7, 0xd7, 0xfb, 0xfb,
	0x14, 0xb5, 0xe1, 0xe8, 0x8e, 0xcf, 0x67, 0xc2, 0xb5, 0xdb, 0xcb, 0x35, 0xb1, 0xd0, 0xcc, 0xa4,
	0xd0, 0xe4, 0x9f, 0xea, 0xc4, 0x0d, 0xa1, 0xac, 0x92, 0xf8, 0x4f, 0xd0, 0x26, 0x04, 0xcf, 0x14,
	0xb4, 0xdc, 0x1c, 0xde, 0xc9, 0x35, 0xb1, 0x88, 0x9b, 0x69, 0x6d, 0x57, 0x5a, 0xae, 0x28, 0x5b,
	0x26, 0xfb, 0xaf, 0x51, 0xd7, 0x6d, 0x67, 0xa2, 0x44, 0xa2, 0x5c, 0xe7, 0xf7, 0x72, 0x4d, 0xba,
	0x6e, 0x85, 0x81, 0x17, 0x9a, 0xf8, 0xf5, 0xbd, 0x06, 0x48, 0xd9, 0x8a, 0x76, 0x65, 0x62, 0xad,
	0x9b, 0x4d, 0x6c, 0xf8, 0xf6, 0xe2, 0x2a, 0x68, 0x5c, 0x5e, 0x05, 0x8d, 0x8b, 0xeb, 0xc0, 0xbb,
	0xbc, 0x0e, 0xbc, 0x6f, 0x8b, 0xa0, 0xf1, 0x7d, 0x11, 0x78, 0x97, 0x8b, 0xa0, 0xf1, 0x73, 0x11,
	0x34, 0x4e, 0x1e, 0x9c, 0xc6, 0xea, 0xd3, 0x7c, 0x34, 0x18, 0xa7, 0xd3, 0x90, 0xa5, 0x52, 0x28,
	0xc5, 0x5f, 0x4e, 0xd2, 0x2f, 0xe1, 0x21, 0xcf, 0xb2, 0x58, 0x64, 0x8f, 0x5e, 0xa5, 0x61, 0xf9,
	0xbb, 0x1e, 0x6d, 0xc0, 0xaf, 0xf7, 0xf1, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xeb, 0x32, 0x12,
	0x28, 0xc2, 0x05, 0x00, 0x00,
}

func (m *TaskData) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TaskData) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TaskData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EventDataList) > 0 {
		for iNdEx := len(m.EventDataList) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EventDataList[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTaskdata(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x62
		}
	}
	if len(m.PartnerList) > 0 {
		for iNdEx := len(m.PartnerList) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PartnerList[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTaskdata(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x5a
		}
	}
	if len(m.Desc) > 0 {
		i -= len(m.Desc)
		copy(dAtA[i:], m.Desc)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Desc)))
		i--
		dAtA[i] = 0x52
	}
	if m.EventCount != 0 {
		i = encodeVarintTaskdata(dAtA, i, uint64(m.EventCount))
		i--
		dAtA[i] = 0x48
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.State) > 0 {
		i -= len(m.State)
		copy(dAtA[i:], m.State)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.State)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.TaskId) > 0 {
		i -= len(m.TaskId)
		copy(dAtA[i:], m.TaskId)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.TaskId)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.DataStatus) > 0 {
		i -= len(m.DataStatus)
		copy(dAtA[i:], m.DataStatus)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.DataStatus)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.DataId) > 0 {
		i -= len(m.DataId)
		copy(dAtA[i:], m.DataId)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.DataId)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.NodeName) > 0 {
		i -= len(m.NodeName)
		copy(dAtA[i:], m.NodeName)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.NodeName)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.NodeId) > 0 {
		i -= len(m.NodeId)
		copy(dAtA[i:], m.NodeId)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.NodeId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Identity) > 0 {
		i -= len(m.Identity)
		copy(dAtA[i:], m.Identity)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Identity)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Partner) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Partner) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Partner) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.NodeName) > 0 {
		i -= len(m.NodeName)
		copy(dAtA[i:], m.NodeName)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.NodeName)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.NodeId) > 0 {
		i -= len(m.NodeId)
		copy(dAtA[i:], m.NodeId)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.NodeId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Identity) > 0 {
		i -= len(m.Identity)
		copy(dAtA[i:], m.Identity)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Identity)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Alias) > 0 {
		i -= len(m.Alias)
		copy(dAtA[i:], m.Alias)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Alias)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventData) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventData) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Identity) > 0 {
		i -= len(m.Identity)
		copy(dAtA[i:], m.Identity)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.Identity)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.EventContent) > 0 {
		i -= len(m.EventContent)
		copy(dAtA[i:], m.EventContent)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.EventContent)))
		i--
		dAtA[i] = 0x22
	}
	if m.EventAt != 0 {
		i = encodeVarintTaskdata(dAtA, i, uint64(m.EventAt))
		i--
		dAtA[i] = 0x18
	}
	if len(m.EventType) > 0 {
		i -= len(m.EventType)
		copy(dAtA[i:], m.EventType)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.EventType)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.TaskId) > 0 {
		i -= len(m.TaskId)
		copy(dAtA[i:], m.TaskId)
		i = encodeVarintTaskdata(dAtA, i, uint64(len(m.TaskId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintTaskdata(dAtA []byte, offset int, v uint64) int {
	offset -= sovTaskdata(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *TaskData) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Identity)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.NodeId)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.NodeName)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.DataId)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.DataStatus)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.TaskId)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.State)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	if m.EventCount != 0 {
		n += 1 + sovTaskdata(uint64(m.EventCount))
	}
	l = len(m.Desc)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	if len(m.PartnerList) > 0 {
		for _, e := range m.PartnerList {
			l = e.ProtoSize()
			n += 1 + l + sovTaskdata(uint64(l))
		}
	}
	if len(m.EventDataList) > 0 {
		for _, e := range m.EventDataList {
			l = e.ProtoSize()
			n += 1 + l + sovTaskdata(uint64(l))
		}
	}
	return n
}

func (m *Partner) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Alias)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.Identity)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.NodeId)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.NodeName)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	return n
}

func (m *EventData) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.TaskId)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.EventType)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	if m.EventAt != 0 {
		n += 1 + sovTaskdata(uint64(m.EventAt))
	}
	l = len(m.EventContent)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	l = len(m.Identity)
	if l > 0 {
		n += 1 + l + sovTaskdata(uint64(l))
	}
	return n
}

func sovTaskdata(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTaskdata(x uint64) (n int) {
	return sovTaskdata(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TaskData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskdata
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TaskData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TaskData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Identity = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodeName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataStatus", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataStatus = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TaskId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TaskId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.State = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventCount", wireType)
			}
			m.EventCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EventCount |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Desc", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Desc = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PartnerList", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PartnerList = append(m.PartnerList, Partner{})
			if err := m.PartnerList[len(m.PartnerList)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventDataList", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EventDataList = append(m.EventDataList, EventData{})
			if err := m.EventDataList[len(m.EventDataList)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaskdata(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskdata
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Partner) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskdata
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Partner: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Partner: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Alias", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Alias = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Identity = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodeName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaskdata(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskdata
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskdata
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TaskId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TaskId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EventType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventAt", wireType)
			}
			m.EventAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EventAt |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventContent", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EventContent = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identity", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskdata
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskdata
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Identity = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaskdata(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskdata
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTaskdata(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTaskdata
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTaskdata
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTaskdata
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTaskdata
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTaskdata
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTaskdata        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTaskdata          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTaskdata = fmt.Errorf("proto: unexpected end of group")
)
