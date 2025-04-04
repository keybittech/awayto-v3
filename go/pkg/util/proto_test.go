package util

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestUtilParseHandlerOptions(t *testing.T) {
	var postPrompt protoreflect.MethodDescriptor
	var postFileContents protoreflect.MethodDescriptor
	var getFileContents protoreflect.MethodDescriptor
	var getBookingTranscripts protoreflect.MethodDescriptor
	var postGroupRole protoreflect.MethodDescriptor

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)

			switch serviceMethod.Name() {
			case "PostPrompt":
				postPrompt = serviceMethod
			case "PostFileContents":
				postFileContents = serviceMethod
			case "GetFileContents":
				getFileContents = serviceMethod
			case "GetBookingTranscripts":
				getBookingTranscripts = serviceMethod
			case "PostGroupRole":
				postGroupRole = serviceMethod
			default:
			}
		}
		return true
	})

	if postPrompt == nil || postFileContents == nil || getFileContents == nil || getBookingTranscripts == nil || postGroupRole == nil {
		t.Fatalf(
			"got nil service method %t %t %t %t %t",
			postPrompt == nil,
			postFileContents == nil,
			getFileContents == nil,
			getBookingTranscripts == nil,
			postGroupRole == nil,
		)
	}

	type args struct {
		md protoreflect.MethodDescriptor
	}
	tests := []struct {
		name     string
		args     args
		validate func(*HandlerOptions) bool
	}{
		{
			name: "cache=STORE",
			args: args{postPrompt},
			validate: func(got *HandlerOptions) bool {
				return got.CacheType == types.CacheType_STORE
			},
		},
		{
			name: "throttle=1",
			args: args{postFileContents},
			validate: func(got *HandlerOptions) bool {
				return got.Throttle == 1
			},
		},
		{
			name: "multipart_request=true",
			args: args{postFileContents},
			validate: func(got *HandlerOptions) bool {
				return got.MultipartRequest == true
			},
		},
		{
			name: "cache=SKIP",
			args: args{getFileContents},
			validate: func(got *HandlerOptions) bool {
				return got.CacheType == types.CacheType_SKIP
			},
		},
		{
			name: "multipart_response=true",
			args: args{getFileContents},
			validate: func(got *HandlerOptions) bool {
				return got.MultipartResponse == true
			},
		},
		{
			name: "cache_duration=180",
			args: args{getBookingTranscripts},
			validate: func(got *HandlerOptions) bool {
				return got.CacheDuration == 180
			},
		},
		{
			name: "site_role=APP_GROUP_ROLES",
			args: args{postGroupRole},
			validate: func(got *HandlerOptions) bool {
				return got.SiteRole == types.SiteRoles_APP_GROUP_ROLES.String()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHandlerOptions(tt.args.md)
			if !tt.validate(got) {
				t.Errorf("ParseHandlerOptions() failed validation for %s", tt.name)
			}
		})
	}
}

func getMethodDescriptor(t *testing.T, methodName string) protoreflect.MethodDescriptor {
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

	if md == nil {
		t.Fatalf("Could not find %s method", methodName)
	}
	return md
}

func getServiceType(t *testing.T, descriptor protoreflect.MethodDescriptor) protoreflect.MessageType {
	var serviceType protoreflect.MessageType

	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {

		if mt.Descriptor().FullName() == descriptor.Input().FullName() {
			serviceType = mt
			return false
		}
		return true
	})

	if serviceType == nil {
		t.Fatal("Could not find input message type for GetUserProfileDetailsBySub")
	}
	return serviceType
}

func TestUtilParseProtoQueryParams(t *testing.T) {
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

func TestUtilParseProtoPathParams(t *testing.T) {
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
				ParseHandlerOptions(tt.method).ServiceMethodURL,
				tt.url,
			)

			if !proto.Equal(pb, tt.want) {
				t.Errorf("ParseProtoPathParams() = %v, want %v", pb, tt.want)
			}
		})
	}
}
