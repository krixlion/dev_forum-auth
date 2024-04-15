// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v4.22.2
// source: ec.proto

package ecpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ECType int32

const (
	ECType_UNDEFINED ECType = 0
	ECType_P256      ECType = 1
	ECType_P384      ECType = 2
	ECType_P521      ECType = 3
)

// Enum value maps for ECType.
var (
	ECType_name = map[int32]string{
		0: "UNDEFINED",
		1: "P256",
		2: "P384",
		3: "P521",
	}
	ECType_value = map[string]int32{
		"UNDEFINED": 0,
		"P256":      1,
		"P384":      2,
		"P521":      3,
	}
)

func (x ECType) Enum() *ECType {
	p := new(ECType)
	*p = x
	return p
}

func (x ECType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ECType) Descriptor() protoreflect.EnumDescriptor {
	return file_ec_proto_enumTypes[0].Descriptor()
}

func (ECType) Type() protoreflect.EnumType {
	return &file_ec_proto_enumTypes[0]
}

func (x ECType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ECType.Descriptor instead.
func (ECType) EnumDescriptor() ([]byte, []int) {
	return file_ec_proto_rawDescGZIP(), []int{0}
}

type EC struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Crv ECType `protobuf:"varint,1,opt,name=crv,proto3,enum=auth.ECType" json:"crv,omitempty"`
	X   string `protobuf:"bytes,2,opt,name=x,proto3" json:"x,omitempty"`
	Y   string `protobuf:"bytes,3,opt,name=y,proto3" json:"y,omitempty"`
}

func (x *EC) Reset() {
	*x = EC{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ec_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EC) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EC) ProtoMessage() {}

func (x *EC) ProtoReflect() protoreflect.Message {
	mi := &file_ec_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EC.ProtoReflect.Descriptor instead.
func (*EC) Descriptor() ([]byte, []int) {
	return file_ec_proto_rawDescGZIP(), []int{0}
}

func (x *EC) GetCrv() ECType {
	if x != nil {
		return x.Crv
	}
	return ECType_UNDEFINED
}

func (x *EC) GetX() string {
	if x != nil {
		return x.X
	}
	return ""
}

func (x *EC) GetY() string {
	if x != nil {
		return x.Y
	}
	return ""
}

var File_ec_proto protoreflect.FileDescriptor

var file_ec_proto_rawDesc = []byte{
	0x0a, 0x08, 0x65, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x61, 0x75, 0x74, 0x68,
	0x22, 0x40, 0x0a, 0x02, 0x45, 0x43, 0x12, 0x1e, 0x0a, 0x03, 0x63, 0x72, 0x76, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x45, 0x43, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x03, 0x63, 0x72, 0x76, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x01, 0x79, 0x2a, 0x35, 0x0a, 0x06, 0x45, 0x43, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a, 0x09,
	0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x50,
	0x32, 0x35, 0x36, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x33, 0x38, 0x34, 0x10, 0x02, 0x12,
	0x08, 0x0a, 0x04, 0x50, 0x35, 0x32, 0x31, 0x10, 0x03, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x72, 0x69, 0x78, 0x6c, 0x69, 0x6f, 0x6e,
	0x2f, 0x64, 0x65, 0x76, 0x5f, 0x66, 0x6f, 0x72, 0x75, 0x6d, 0x2d, 0x61, 0x75, 0x74, 0x68, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x63, 0x3b, 0x65,
	0x63, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ec_proto_rawDescOnce sync.Once
	file_ec_proto_rawDescData = file_ec_proto_rawDesc
)

func file_ec_proto_rawDescGZIP() []byte {
	file_ec_proto_rawDescOnce.Do(func() {
		file_ec_proto_rawDescData = protoimpl.X.CompressGZIP(file_ec_proto_rawDescData)
	})
	return file_ec_proto_rawDescData
}

var file_ec_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ec_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ec_proto_goTypes = []interface{}{
	(ECType)(0), // 0: auth.ECType
	(*EC)(nil),  // 1: auth.EC
}
var file_ec_proto_depIdxs = []int32{
	0, // 0: auth.EC.crv:type_name -> auth.ECType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ec_proto_init() }
func file_ec_proto_init() {
	if File_ec_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ec_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EC); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ec_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ec_proto_goTypes,
		DependencyIndexes: file_ec_proto_depIdxs,
		EnumInfos:         file_ec_proto_enumTypes,
		MessageInfos:      file_ec_proto_msgTypes,
	}.Build()
	File_ec_proto = out.File
	file_ec_proto_rawDesc = nil
	file_ec_proto_goTypes = nil
	file_ec_proto_depIdxs = nil
}
