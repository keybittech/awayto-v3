package util

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type HandlerOptions struct {
	NoLogFields       []protoreflect.Name
	SiteRoleName      string
	Pattern           string
	ServiceMethodURL  string
	CacheDuration     int64
	CacheType         int64
	SiteRole          int64
	Throttle          int64
	HasQueryParams    bool
	HasPathParams     bool
	MultipartResponse bool
	MultipartRequest  bool
	UseTx             bool
}

func ParseHandlerOptions(md protoreflect.MethodDescriptor) *HandlerOptions {
	parsedOptions := &HandlerOptions{}
	fieldsLen, err := Itoi32(md.Input().Fields().Len())
	if err != nil {
		log.Fatal(err)
	}

	if md.Input().Fields().Len() > 0 {
		for i := int32(0); i < fieldsLen; i++ {
			field := md.Input().Fields().ByNumber(protowire.Number(i + 1))

			if proto.HasExtension(field.Options(), types.E_Nolog) {
				parsedOptions.NoLogFields = append(parsedOptions.NoLogFields, field.Name())
			}
		}
	}

	inputOpts := md.Options().(*descriptor.MethodOptions)

	var serviceMethodMethod, serviceMethodUrl string

	// get the URL of the handler
	httpRule := &annotations.HttpRule{}

	if proto.HasExtension(inputOpts, annotations.E_Http) {
		ext := proto.GetExtension(inputOpts, annotations.E_Http)
		httpRule, _ = ext.(*annotations.HttpRule)
	}

	switch {
	case httpRule.GetPost() != "":
		serviceMethodMethod = "POST"
		serviceMethodUrl = httpRule.GetPost()
	case httpRule.GetGet() != "":
		serviceMethodMethod = "GET"
		serviceMethodUrl = httpRule.GetGet()
	case httpRule.GetPut() != "":
		serviceMethodMethod = "PUT"
		serviceMethodUrl = httpRule.GetPut()
	case httpRule.GetDelete() != "":
		serviceMethodMethod = "DELETE"
		serviceMethodUrl = httpRule.GetDelete()
	case httpRule.GetPatch() != "":
		serviceMethodMethod = "PATCH"
		serviceMethodUrl = httpRule.GetPatch()
	default:
	}

	parsedOptions.ServiceMethodURL = serviceMethodUrl

	// attach /api to /v1 /v2, etc -- resulting in /api/v1/ which is the standard API_PATH
	parsedOptions.Pattern = fmt.Sprintf("%s /api%s", serviceMethodMethod, serviceMethodUrl)

	if strings.Contains(serviceMethodUrl, "?") {
		parsedOptions.HasQueryParams = true
	}

	if strings.Contains(serviceMethodUrl, "{") {
		parsedOptions.HasPathParams = true
	}

	if proto.HasExtension(inputOpts, types.E_SiteRole) {
		siteRoles := proto.GetExtension(inputOpts, types.E_SiteRole).(types.SiteRoles)
		roles := strings.Split(fmt.Sprint(siteRoles), ",")
		roleBits := StringsToBitmask(roles)
		parsedOptions.SiteRole = roleBits
		if roleBits > math.MinInt32 && roleBits < math.MaxInt32 {
			parsedOptions.SiteRoleName = types.SiteRoles_name[int32(roleBits)]
		}
	}

	if proto.HasExtension(inputOpts, types.E_Cache) {
		cacheType := proto.GetExtension(inputOpts, types.E_Cache).(types.CacheType)
		parsedOptions.CacheType = int64(cacheType)
	}

	if proto.HasExtension(inputOpts, types.E_CacheDuration) {
		parsedOptions.CacheDuration = proto.GetExtension(inputOpts, types.E_CacheDuration).(int64)
	}

	if proto.HasExtension(inputOpts, types.E_Throttle) {
		parsedOptions.Throttle = proto.GetExtension(inputOpts, types.E_Throttle).(int64)
	}

	if proto.HasExtension(inputOpts, types.E_MultipartRequest) {
		parsedOptions.MultipartRequest = proto.GetExtension(inputOpts, types.E_MultipartRequest).(bool)
	}

	if proto.HasExtension(inputOpts, types.E_MultipartResponse) {
		parsedOptions.MultipartResponse = proto.GetExtension(inputOpts, types.E_MultipartResponse).(bool)
	}

	if proto.HasExtension(inputOpts, types.E_UseTx) {
		parsedOptions.UseTx = proto.GetExtension(inputOpts, types.E_UseTx).(bool)
	}

	return parsedOptions
}

func parseTag(field reflect.StructField, fieldName string) string {
	tagValue := field.Tag.Get(fieldName)
	if tagValue == "" {
		tagValue = strings.ToLower(field.Name[:1]) + field.Name[1:]
	} else if commaIndex := strings.Index(tagValue, ","); commaIndex != -1 {
		tagValue = tagValue[:commaIndex]
	}
	return tagValue
}

// Common utility function to handle setting field values
func setProtoFieldValue(msg proto.Message, jsonName string, value string) {
	reflectMsg := msg.ProtoReflect()
	descriptor := reflectMsg.Descriptor()
	fields := descriptor.Fields()

	for i := range fields.Len() {
		field := fields.Get(i)
		if field.JSONName() == jsonName && field.Kind() == protoreflect.StringKind {
			reflectMsg.Set(field, protoreflect.ValueOfString(value))
			return
		}
	}
}

// ParseProtoQueryParams sets proto message fields from URL query parameters
func ParseProtoQueryParams(msg proto.Message, req *http.Request) {
	queryParams := req.URL.Query()
	if len(queryParams) == 0 {
		return
	}

	reflectMsg := msg.ProtoReflect()
	descriptor := reflectMsg.Descriptor()
	fields := descriptor.Fields()

	for i := range fields.Len() {
		field := fields.Get(i)
		jsonName := field.JSONName()

		if values, ok := queryParams[jsonName]; ok && len(values) > 0 {
			if field.Kind() == protoreflect.StringKind {
				reflectMsg.Set(field, protoreflect.ValueOfString(values[0]))
			}
		}
	}
}

// ParseProtoPathParams sets proto message fields from path parameters
func ParseProtoPathParams(msg proto.Message, methodParams []string, req *http.Request) {
	requestParams := strings.Split(req.URL.Path[4:], "/")
	if len(methodParams) == 0 || len(methodParams) != len(requestParams) {
		return
	}

	for i := range len(methodParams) {
		paramName := methodParams[i]
		if strings.HasPrefix(paramName, "{") {
			// Extract the parameter name without braces
			paramName = strings.Trim(paramName, "{}")

			// Set the field value if it exists
			setProtoFieldValue(msg, paramName, requestParams[i])
		}
	}
}
