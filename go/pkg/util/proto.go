package util

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type HandlerOptions struct {
	Invalidations     []string
	NoLogFields       []protoreflect.Name
	ServiceMethod     protoreflect.MethodDescriptor
	ServiceMethodType protoreflect.MessageType
	ServiceMethodName string
	ServiceMethodURL  string
	SiteRoleName      string
	Pattern           string
	CacheDuration     int64
	CacheType         int64
	SiteRole          int64
	Throttle          int64
	NumInvalidations  int32
	HasQueryParams    bool
	HasPathParams     bool
	MultipartResponse bool
	MultipartRequest  bool
	ShouldStore       bool
	ShouldSkip        bool
	UseTx             bool
}

func ParseHandlerOptions(md protoreflect.MethodDescriptor) *HandlerOptions {
	parsedOptions := &HandlerOptions{
		ServiceMethod:     md,
		ServiceMethodName: string(md.Name()),
	}

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

	parsedOptions.NoLogFields = append(parsedOptions.NoLogFields, DEFAULT_IGNORED_PROTO_FIELDS...)

	inputOpts := md.Options().(*descriptor.MethodOptions)

	httpRule, ok := proto.GetExtension(inputOpts, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		log.Fatalf("service method %s doesn't have an http rule", parsedOptions.ServiceMethodName)
	}

	var serviceMethodMethod, serviceMethodURL string

	switch {
	case httpRule.GetPost() != "":
		serviceMethodMethod = "POST"
		serviceMethodURL = httpRule.GetPost()
	case httpRule.GetGet() != "":
		serviceMethodMethod = "GET"
		serviceMethodURL = httpRule.GetGet()
	case httpRule.GetPut() != "":
		serviceMethodMethod = "PUT"
		serviceMethodURL = httpRule.GetPut()
	case httpRule.GetDelete() != "":
		serviceMethodMethod = "DELETE"
		serviceMethodURL = httpRule.GetDelete()
	case httpRule.GetPatch() != "":
		serviceMethodMethod = "PATCH"
		serviceMethodURL = httpRule.GetPatch()
	default:
	}

	parsedOptions.ServiceMethodURL = "/api" + serviceMethodURL
	parsedOptions.Pattern = serviceMethodMethod + " " + parsedOptions.ServiceMethodURL

	if strings.Contains(serviceMethodURL, "?") {
		parsedOptions.HasQueryParams = true
	}

	if strings.Contains(serviceMethodURL, "{") {
		parsedOptions.HasPathParams = true
	}

	if siteRoles, ok := proto.GetExtension(inputOpts, types.E_SiteRole).(types.SiteRoles); ok {
		roles := strings.Split(fmt.Sprint(siteRoles), ",")
		roleBits := StringsToBitmask(roles)
		parsedOptions.SiteRole = roleBits
		if roleBits > math.MinInt32 && roleBits < math.MaxInt32 {
			parsedOptions.SiteRoleName = types.SiteRoles_name[int32(roleBits)]
		}
	}

	if cacheType, ok := proto.GetExtension(inputOpts, types.E_Cache).(types.CacheType); ok {
		parsedOptions.CacheType = int64(cacheType)
	}

	parsedOptions.ShouldStore = parsedOptions.CacheType == int64(types.CacheType_STORE)
	parsedOptions.ShouldSkip = parsedOptions.CacheType == int64(types.CacheType_SKIP)

	if cacheDuration, ok := proto.GetExtension(inputOpts, types.E_CacheDuration).(int64); ok {
		parsedOptions.CacheDuration = cacheDuration
	}

	if throttle, ok := proto.GetExtension(inputOpts, types.E_Throttle).(int64); ok {
		parsedOptions.Throttle = throttle
	}

	if multipartRequest, ok := proto.GetExtension(inputOpts, types.E_MultipartRequest).(bool); ok {
		parsedOptions.MultipartRequest = multipartRequest
	}

	if multipartResponse, ok := proto.GetExtension(inputOpts, types.E_MultipartResponse).(bool); ok {
		parsedOptions.MultipartResponse = multipartResponse
	}

	if useTx, ok := proto.GetExtension(inputOpts, types.E_UseTx).(bool); ok {
		parsedOptions.UseTx = useTx
	}

	return parsedOptions
}

// Make wildcard patterns out of urls /path/{param} -> /path/*
func parseInvalidation(i string) string {
	return string(regexp.MustCompile("{[^}]+}").ReplaceAll([]byte(i), []byte("*")))
}

func ParseInvalidations(handlerOptions map[string]*HandlerOptions) {
	for _, opts := range handlerOptions {

		// Unless needing to permanently store results in redis, mutations should invalidate the GET
		if !opts.ShouldStore {
			opts.Invalidations = append(opts.Invalidations, parseInvalidation(opts.ServiceMethodURL))
		}

		inputOpts := opts.ServiceMethod.Options().(*descriptor.MethodOptions)

		if invalidates, ok := proto.GetExtension(inputOpts, types.E_Invalidates).([]string); ok {
			for _, invalidation := range invalidates {
				invalidateHandler, ok := handlerOptions[invalidation]
				if !ok {
					log.Fatalf("%s invalidates unknown handler %s", opts.ServiceMethodName, invalidation)
				}
				opts.Invalidations = append(opts.Invalidations, parseInvalidation(invalidateHandler.ServiceMethodURL))
			}
		}
	}
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

// Sets proto message fields from URL query parameters
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

// Sets proto message fields from path parameters
func ParseProtoPathParams(msg proto.Message, methodParams []string, req *http.Request) {
	requestParams := strings.Split(req.URL.Path, "/")
	if len(methodParams) == 0 || len(methodParams) != len(requestParams) {
		return
	}

	for i := range len(methodParams) {
		paramName := methodParams[i]
		if strings.HasPrefix(paramName, "{") {
			paramName = strings.Trim(paramName, "{}")

			setProtoFieldValue(msg, paramName, requestParams[i])
		}
	}
}
