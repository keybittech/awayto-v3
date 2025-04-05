package util

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func getMethodDescriptor(testModule interface{}, methodName string) protoreflect.MethodDescriptor {

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

		for s := 0; s < fd.Services().Len(); s++ {
			service := fd.Services().Get(s)

			for i := 0; i < service.Methods().Len(); i++ {
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

func getServiceType(testModule interface{}, inputDescriptor protoreflect.MethodDescriptor) protoreflect.MessageType {
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
				return got.CacheType == types.CacheType_STORE
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
				return got.CacheType == types.CacheType_SKIP
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
				return got.SiteRole == types.SiteRoles_APP_GROUP_ROLES.String()
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
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ParseHandlerOptions(md)
	}
}

func TestParseProtoQueryParams(t *testing.T) {
	tests := []struct {
		name        string
		method      protoreflect.MethodDescriptor
		queryParams url.Values
		want        proto.Message
	}{
		{
			name:   "serializes query paramaters into pb struct",
			method: getMethodDescriptor(t, "GetUserProfileDetailsBySub"),
			queryParams: url.Values{
				"sub": []string{"test"},
			},
			want: &types.GetUserProfileDetailsBySubRequest{
				Sub: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := getServiceType(t, tt.method).New().Interface()
			ParseProtoQueryParams(reflect.ValueOf(pb).Elem(), tt.queryParams)
			if !proto.Equal(pb, tt.want) {
				t.Errorf("ParseProtoQueryParams() = %v, want %v", pb, tt.want)
			}
		})
	}
}

func BenchmarkParseProtoQueryParamsComplex(b *testing.B) {
	md := getMethodDescriptor(b, "GetUserProfileDetailsBySub")
	pb := getServiceType(b, md).New().Interface()
	// More complex query with multiple parameters
	queryParams := url.Values{
		"sub":     []string{"test-subject"},
		"user_id": []string{"123456"},
		"role":    []string{"admin"},
		"active":  []string{"true"},
	}
	val := reflect.ValueOf(pb).Elem()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseProtoQueryParams(val, queryParams)
	}
}

func BenchmarkParseProtoQueryParams(b *testing.B) {
	md := getMethodDescriptor(b, "GetUserProfileDetailsBySub")
	pb := getServiceType(b, md).New().Interface()
	queryParams := url.Values{
		"sub": []string{"test"},
	}
	val := reflect.ValueOf(pb).Elem()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseProtoQueryParams(val, queryParams)
	}
}

func TestParseProtoPathParams(t *testing.T) {
	tests := []struct {
		name   string
		method protoreflect.MethodDescriptor
		url    string
		want   proto.Message
	}{
		{
			name:   "serializes path parameters into pb struct",
			method: getMethodDescriptor(t, "GetGroupScheduleByDate"),
			url:    "/api/v1/group/schedules/group-schedule-id/date/date-value",
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
				reflect.ValueOf(pb).Elem(),
				strings.Split(ParseHandlerOptions(tt.method).ServiceMethodURL, "/"),
				strings.Split(strings.TrimPrefix(tt.url, "/api"), "/"),
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
	url := "/api/v1/group/schedules/group-schedule-id/date/date-value"
	options := ParseHandlerOptions(md)
	val := reflect.ValueOf(pb).Elem()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseProtoPathParams(val, strings.Split(options.ServiceMethodURL, "/"), strings.Split(strings.TrimLeft(url, "/api"), "/"))
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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = parseTag(field, "json")
	}
}

func Benchmark_parseTagParsed(b *testing.B) {
	pb := getServiceType(b, getMethodDescriptor(b, "GetGroupScheduleByDate")).New().Interface()
	field, ok := reflect.ValueOf(pb).Elem().Type().FieldByName("GroupScheduleId")
	if !ok {
		b.Fatal("no field GroupScheduleId")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = parseTag(field, "other")
	}
}
