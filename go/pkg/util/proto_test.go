package util

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func getMethodDescriptor(testModule any, methodName string) protoreflect.MethodDescriptor {

	md := getMethodDescriptorOp(methodName)

	if md == nil {
		if mod, ok := testModule.(*testing.T); ok {
			mod.Fatalf("Could not find %s method", methodName)
		}
		if mod, ok := testModule.(*testing.B); ok {
			mod.Fatalf("Could not find %s method", methodName)
		}
	}

	return md
}

func getMethodDescriptorOp(methodName string) protoreflect.MethodDescriptor {
	var md protoreflect.MethodDescriptor

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		for s := range fd.Services().Len() {
			service := fd.Services().Get(s)

			for i := range service.Methods().Len() {
				method := service.Methods().Get(i)

				if string(method.Name()) == methodName {
					md = method
					return false
				}
			}
		}
		return true
	})

	return md
}

func getServiceType(testModule any, inputDescriptor protoreflect.MethodDescriptor) protoreflect.MessageType {
	st := getServiceTypeOp(inputDescriptor)

	if st == nil {
		if mod, ok := testModule.(*testing.T); ok {
			mod.Fatalf("Could not find %s input descriptor", inputDescriptor.Input().FullName())
		}
		if mod, ok := testModule.(*testing.B); ok {
			mod.Fatalf("Could not find %s input descriptor", inputDescriptor.Input().FullName())
		}
	}

	return st
}

func getServiceTypeOp(descriptor protoreflect.MethodDescriptor) protoreflect.MessageType {
	var serviceType protoreflect.MessageType

	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {

		if mt.Descriptor().FullName() == descriptor.Input().FullName() {
			serviceType = mt
			return false
		}
		return true
	})

	return serviceType
}

func TestParseHandlerOptions(t *testing.T) {
	type args struct {
		md protoreflect.MethodDescriptor
	}
	tests := []struct {
		name     string
		md       protoreflect.MethodDescriptor
		validate func(*HandlerOptions) bool
	}{
		{
			name: "cache=STORE",
			md:   getMethodDescriptor(t, "PostPrompt"),
			validate: func(got *HandlerOptions) bool {
				return got.CacheType == int64(types.CacheType_STORE)
			},
		},
		{
			name: "throttle=1",
			md:   getMethodDescriptor(t, "PostFileContents"),
			validate: func(got *HandlerOptions) bool {
				return got.Throttle == 1
			},
		},
		{
			name: "multipart_request=true",
			md:   getMethodDescriptor(t, "PostFileContents"),
			validate: func(got *HandlerOptions) bool {
				return got.MultipartRequest == true
			},
		},
		{
			name: "cache=SKIP",
			md:   getMethodDescriptor(t, "GetFileContents"),
			validate: func(got *HandlerOptions) bool {
				return got.CacheType == int64(types.CacheType_SKIP)
			},
		},
		{
			name: "multipart_response=true",
			md:   getMethodDescriptor(t, "GetFileContents"),
			validate: func(got *HandlerOptions) bool {
				return got.MultipartResponse == true
			},
		},
		{
			name: "cache_duration=180",
			md:   getMethodDescriptor(t, "GetBookingTranscripts"),
			validate: func(got *HandlerOptions) bool {
				return got.CacheDuration == 180
			},
		},
		{
			name: "site_role=APP_GROUP_ROLES",
			md:   getMethodDescriptor(t, "PostGroupRole"),
			validate: func(got *HandlerOptions) bool {
				return got.SiteRole == int64(types.SiteRoles_APP_GROUP_ROLES)
			},
		},
		{
			name: "use_tx=true",
			md:   getMethodDescriptor(t, "PostGroupRole"),
			validate: func(got *HandlerOptions) bool {
				return got.UseTx == true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHandlerOptions(tt.md)
			if !tt.validate(got) {
				t.Errorf("ParseHandlerOptions() failed validation for %s", tt.name)
			}
		})
	}
}

func BenchmarkParseHandlerOptions(b *testing.B) {
	md := getMethodDescriptor(b, "PostPrompt")
	reset(b)

	for b.Loop() {
		_ = ParseHandlerOptions(md)
	}
}

func makeParseTestReq(pathOrQuery string) *http.Request {
	newReq, err := http.NewRequest("GET", pathOrQuery, nil)
	if err != nil {
		panic(err)
	}
	return newReq
}

func TestParseProtoQueryParams(t *testing.T) {
	tests := []struct {
		name   string
		method protoreflect.MethodDescriptor
		req    *http.Request
		want   proto.Message
	}{
		{
			name:   "serializes query paramaters into pb struct",
			method: getMethodDescriptor(t, "CheckGroupName"),
			req:    makeParseTestReq("/blah?name=test"),
			want: &types.CheckGroupNameRequest{
				Name: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := getServiceType(t, tt.method).New().Interface()
			ParseProtoQueryParams(pb, tt.req)
			if !proto.Equal(pb, tt.want) {
				t.Errorf("ParseProtoQueryParams() = %v, want %v", pb, tt.want)
			}
		})
	}
}

func BenchmarkParseProtoQueryParamsComplex(b *testing.B) {
	md := getMethodDescriptor(b, "PostQuote")
	pb := getServiceType(b, md).New().Interface()
	req := makeParseTestReq("/blah?scheduleBracketSlotId=12351235&serviceTierId=64236432&slotDate=04-12-2025")
	reset(b)

	for b.Loop() {
		ParseProtoQueryParams(pb, req)
	}
}

func BenchmarkParseProtoQueryParams(b *testing.B) {
	md := getMethodDescriptor(b, "GetQuoteById")
	pb := getServiceType(b, md).New().Interface()
	req := makeParseTestReq("/blah?id=0283743241")
	reset(b)

	for b.Loop() {
		ParseProtoQueryParams(pb, req)
	}
}

func TestParseProtoPathParams(t *testing.T) {
	tests := []struct {
		name   string
		method protoreflect.MethodDescriptor
		req    *http.Request
		want   proto.Message
	}{
		{
			name:   "serializes path parameters into pb struct",
			method: getMethodDescriptor(t, "GetGroupScheduleByDate"),
			req:    makeParseTestReq("/api/v1/group/schedules/group-schedule-id/date/date-value"),
			want: &types.GetGroupScheduleByDateRequest{
				GroupScheduleId: "group-schedule-id",
				Date:            "date-value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := getServiceType(t, tt.method).New().Interface()
			ParseProtoPathParams(
				pb,
				strings.Split(ParseHandlerOptions(tt.method).ServiceMethodURL, "/"),
				tt.req,
			)

			if !proto.Equal(pb, tt.want) {
				t.Errorf("ParseProtoPathParams() = %v, want %v", pb, tt.want)
			}
		})
	}
}

func BenchmarkParseProtoPathParams(b *testing.B) {
	md := getMethodDescriptor(b, "GetGroupScheduleByDate")
	pb := getServiceType(b, md).New().Interface()
	req := makeParseTestReq("/api/v1/group/schedules/group-schedule-id/date/date-value")
	methodParams := strings.Split(ParseHandlerOptions(md).ServiceMethodURL, "/")
	reset(b)
	for b.Loop() {
		ParseProtoPathParams(pb, methodParams, req)
	}
}

func Test_parseTag(t *testing.T) {
	pb := getServiceType(t, getMethodDescriptor(t, "GetGroupScheduleByDate")).New().Interface()
	field, ok := reflect.ValueOf(pb).Elem().Type().FieldByName("GroupScheduleId")
	if !ok {
		t.Fatal("no field GroupScheduleId")
	}
	type args struct {
		field     reflect.StructField
		fieldName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "parses existing field", args: args{field, "json"}, want: "groupScheduleId"},
		{name: "handles non-existing field", args: args{field, "other"}, want: "groupScheduleId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTag(tt.args.field, tt.args.fieldName); got != tt.want {
				t.Errorf("parseTag(%v, %v) = %v, want %v", tt.args.field, tt.args.fieldName, got, tt.want)
			}
		})
	}
}

func Benchmark_parseTag(b *testing.B) {
	pb := getServiceType(b, getMethodDescriptor(b, "GetGroupScheduleByDate")).New().Interface()
	field, ok := reflect.ValueOf(pb).Elem().Type().FieldByName("GroupScheduleId")
	if !ok {
		b.Fatal("no field GroupScheduleId")
	}

	reset(b)

	for b.Loop() {
		_ = parseTag(field, "json")
	}
}

func Benchmark_parseTagParsed(b *testing.B) {
	pb := getServiceType(b, getMethodDescriptor(b, "GetGroupScheduleByDate")).New().Interface()
	field, ok := reflect.ValueOf(pb).Elem().Type().FieldByName("GroupScheduleId")
	if !ok {
		b.Fatal("no field GroupScheduleId")
	}

	reset(b)

	for b.Loop() {
		_ = parseTag(field, "other")
	}
}

func Test_setProtoFieldValue(t *testing.T) {
	type args struct {
		msg      proto.Message
		jsonName string
		value    string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setProtoFieldValue(tt.args.msg, tt.args.jsonName, tt.args.value)
		})
	}
}
