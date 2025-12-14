package util

import (
	"fmt"
	"log"
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
	"google.golang.org/protobuf/reflect/protoregistry"
)

type HandlerOptionsConfig struct {
	Invalidations          []string
	NoLogFields            []protoreflect.Name
	ServiceMethod          protoreflect.MethodDescriptor
	ServiceMethodInputType protoreflect.MessageType
	ServiceMethodName      string
	ServiceMethodURL       string
	Pattern                string
	CacheDuration          uint32
	CacheType              types.CacheType
	NumInvalidations       uint32
	Throttle               uint32
	SiteRole               int32
	HasQueryParams         bool
	HasPathParams          bool
	MultipartResponse      bool
	MultipartRequest       bool
	ResetsGroup            bool
	ShouldStore            bool
	ShouldSkip             bool
	UseTx                  bool
}

type HandlerOptions struct {
	Invalidations          []string
	NoLogFields            []protoreflect.Name
	ServiceMethod          protoreflect.MethodDescriptor
	ServiceMethodInputType protoreflect.MessageType
	ServiceMethodName      string
	ServiceMethodURL       string
	Pattern                string

	packedNumeric  uint32
	packedBooleans uint8
}

type UnpackedOptionsData struct {
	CacheDuration     uint32
	CacheType         types.CacheType
	NumInvalidations  uint32
	Throttle          uint32
	SiteRole          int32
	HasQueryParams    bool
	HasPathParams     bool
	MultipartResponse bool
	MultipartRequest  bool
	ResetsGroup       bool
	ShouldStore       bool
	ShouldSkip        bool
	UseTx             bool
}

const (
	cacheDurationBits    = 8  // max 256 second cache duration
	cacheTypeBits        = 2  // DEFAULT, SKIP, STORE
	numInvalidationsBits = 4  // an endpoint could invalidate 16 others
	throttleBits         = 8  // prevent endpoint use up to every 256 seconds
	siteRoleBits         = 10 // 10 supported role groups, i.e. APP_GROUP_ADMIN

	siteRoleShift         = 0
	throttleShift         = siteRoleShift + siteRoleBits
	numInvalidationsShift = throttleShift + throttleBits
	cacheTypeShift        = numInvalidationsShift + numInvalidationsBits
	cacheDurationShift    = cacheTypeShift + cacheTypeBits

	cacheDurationMask    = (1 << cacheDurationBits) - 1
	cacheTypeMask        = (1 << cacheTypeBits) - 1
	numInvalidationsMask = (1 << numInvalidationsBits) - 1
	throttleMask         = (1 << throttleBits) - 1
	siteRoleMask         = (1 << siteRoleBits) - 1
)

const (
	useTxBit             = 1 << 0
	shouldSkipBit        = 1 << 1
	shouldStoreBit       = 1 << 2
	resetsGroupBit       = 1 << 3
	multipartRequestBit  = 1 << 4
	multipartResponseBit = 1 << 5
	hasPathParamsBit     = 1 << 6
	hasQueryParamsBit    = 1 << 7
)

func NewHandlerOptions(config HandlerOptionsConfig) (*HandlerOptions, error) {
	var packedNumeric uint32
	var packedBooleans uint8

	if config.CacheDuration > cacheDurationMask {
		return nil, fmt.Errorf("CacheDuration %d out of range (max %d)", config.CacheDuration, cacheDurationMask)
	}
	if config.CacheType > cacheTypeMask {
		return nil, fmt.Errorf("CacheType %d out of range (max %d)", config.CacheType, cacheTypeMask)
	}
	if config.NumInvalidations > numInvalidationsMask {
		return nil, fmt.Errorf("NumInvalidations %d out of range (max %d)", config.NumInvalidations, numInvalidationsMask)
	}
	if config.Throttle > throttleMask {
		return nil, fmt.Errorf("Throttle %d out of range (max %d)", config.Throttle, throttleMask)
	}
	if config.SiteRole > siteRoleMask {
		return nil, fmt.Errorf("SiteRole %d out of range (max %d)", config.SiteRole, siteRoleMask)
	}

	cacheType, err := Itoui32(int(config.CacheType))
	if err != nil {
		return nil, fmt.Errorf("CacheType %d could not parse as uint32", config.CacheType)
	}

	siteRole, err := Itoui32(int(config.SiteRole))
	if err != nil {
		return nil, fmt.Errorf("SiteRole %d could not parse as uint32", config.SiteRole)
	}

	packedNumeric |= (config.CacheDuration << cacheDurationShift)
	packedNumeric |= (cacheType << cacheTypeShift)
	packedNumeric |= (config.NumInvalidations << numInvalidationsShift)
	packedNumeric |= (config.Throttle << throttleShift)
	packedNumeric |= (siteRole << siteRoleShift)

	if config.UseTx {
		packedBooleans |= useTxBit
	}
	if config.ShouldSkip {
		packedBooleans |= shouldSkipBit
	}
	if config.ShouldStore {
		packedBooleans |= shouldStoreBit
	}
	if config.ResetsGroup {
		packedBooleans |= resetsGroupBit
	}
	if config.MultipartRequest {
		packedBooleans |= multipartRequestBit
	}
	if config.MultipartResponse {
		packedBooleans |= multipartResponseBit
	}
	if config.HasPathParams {
		packedBooleans |= hasPathParamsBit
	}
	if config.HasQueryParams {
		packedBooleans |= hasQueryParamsBit
	}

	return &HandlerOptions{
		Invalidations:          config.Invalidations,
		NoLogFields:            config.NoLogFields,
		ServiceMethod:          config.ServiceMethod,
		ServiceMethodInputType: config.ServiceMethodInputType,
		ServiceMethodName:      config.ServiceMethodName,
		ServiceMethodURL:       config.ServiceMethodURL,
		Pattern:                config.Pattern,
		packedNumeric:          packedNumeric,
		packedBooleans:         packedBooleans,
	}, nil
}

func (h *HandlerOptions) Unpack() UnpackedOptionsData {
	var data UnpackedOptionsData

	data.CacheDuration = (h.packedNumeric >> cacheDurationShift) & cacheDurationMask

	cacheType, err := Itoi32(int((h.packedNumeric >> cacheTypeShift) & cacheTypeMask))
	if err != nil {
		panic(ErrCheck(err))
	}

	siteRole, err := Itoi32(int((h.packedNumeric >> siteRoleShift) & siteRoleMask))
	if err != nil {
		panic(ErrCheck(err))
	}

	data.CacheType = types.CacheType(cacheType)
	data.NumInvalidations = (h.packedNumeric >> numInvalidationsShift) & numInvalidationsMask
	data.Throttle = (h.packedNumeric >> throttleShift) & throttleMask
	data.SiteRole = siteRole

	data.UseTx = (h.packedBooleans & useTxBit) != 0
	data.ShouldSkip = (h.packedBooleans & shouldSkipBit) != 0
	data.ShouldStore = (h.packedBooleans & shouldStoreBit) != 0
	data.ResetsGroup = (h.packedBooleans & resetsGroupBit) != 0
	data.MultipartRequest = (h.packedBooleans & multipartRequestBit) != 0
	data.MultipartResponse = (h.packedBooleans & multipartResponseBit) != 0
	data.HasPathParams = (h.packedBooleans & hasPathParamsBit) != 0
	data.HasQueryParams = (h.packedBooleans & hasQueryParamsBit) != 0

	return data
}

func ParseHandlerOptions(md protoreflect.MethodDescriptor) *HandlerOptions {
	parsedOptions := HandlerOptionsConfig{
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

	if siteRoles, ok := proto.GetExtension(inputOpts, types.E_SiteRole).([]types.SiteRoles); ok {
		roles := strings.Split(fmt.Sprint(siteRoles), ",")
		roleBits := StringsToSiteRoles(roles)
		parsedOptions.SiteRole = roleBits
	}

	if cacheType, ok := proto.GetExtension(inputOpts, types.E_Cache).(types.CacheType); ok {
		parsedOptions.CacheType = cacheType
	}

	parsedOptions.ShouldStore = parsedOptions.CacheType == types.CacheType_STORE
	parsedOptions.ShouldSkip = parsedOptions.CacheType == types.CacheType_SKIP

	if cacheDuration, ok := proto.GetExtension(inputOpts, types.E_CacheDuration).(uint32); ok {
		parsedOptions.CacheDuration = cacheDuration
	}

	if throttle, ok := proto.GetExtension(inputOpts, types.E_Throttle).(uint32); ok {
		parsedOptions.Throttle = throttle
	}

	if multipartRequest, ok := proto.GetExtension(inputOpts, types.E_MultipartRequest).(bool); ok {
		parsedOptions.MultipartRequest = multipartRequest
	}

	if multipartResponse, ok := proto.GetExtension(inputOpts, types.E_MultipartResponse).(bool); ok {
		parsedOptions.MultipartResponse = multipartResponse
	}

	if resetsGroup, ok := proto.GetExtension(inputOpts, types.E_ResetsGroup).(bool); ok {
		parsedOptions.ResetsGroup = resetsGroup
	}

	if useTx, ok := proto.GetExtension(inputOpts, types.E_UseTx).(bool); ok {
		parsedOptions.UseTx = useTx
	}

	hops, err := NewHandlerOptions(parsedOptions)
	if err != nil {
		log.Fatalf("error making new handler options %v", err)
	}

	return hops
}

func GenerateOptions() map[string]*HandlerOptions {
	opts := make(map[string]*HandlerOptions)
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if fd.Services().Len() == 0 {
			return true
		}

		services := fd.Services().Get(0)

		for i := 0; i <= services.Methods().Len()-1; i++ {
			serviceMethod := services.Methods().Get(i)
			handlerOpts := ParseHandlerOptions(serviceMethod)

			handlerOpts.ServiceMethodInputType = GetMessageType(serviceMethod.Input().FullName())

			opts[handlerOpts.ServiceMethodName] = handlerOpts
		}

		return true
	})

	// Do this after when all opts are configured
	ParseInvalidations(opts)

	return opts
}

// Make wildcard patterns out of urls /path/{param} -> /path/*
func serviceMethodWildcard(i string) string {
	return string(regexp.MustCompile("{[^}]+}").ReplaceAll([]byte(i), []byte("*")))
}

func ParseInvalidations(handlerOptions map[string]*HandlerOptions) {
	for _, opts := range handlerOptions {
		// Unless needing to permanently store results in redis, mutations should invalidate the GET
		if !opts.Unpack().ShouldStore {
			opts.Invalidations = append(opts.Invalidations, serviceMethodWildcard(opts.ServiceMethodURL))
		}

		inputOpts := opts.ServiceMethod.Options().(*descriptor.MethodOptions)

		if invalidates, ok := proto.GetExtension(inputOpts, types.E_Invalidates).([]string); ok {
			for _, invalidation := range invalidates {
				invalidateHandler, ok := handlerOptions[invalidation]
				if !ok {
					log.Fatalf("%s invalidates unknown handler %s", opts.ServiceMethodName, invalidation)
				}
				opts.Invalidations = append(opts.Invalidations, serviceMethodWildcard(invalidateHandler.ServiceMethodURL))
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

func GetMessageType(messageName protoreflect.FullName) protoreflect.MessageType {
	var messageDescriptor protoreflect.MessageType
	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		if mt.Descriptor().FullName() == protoreflect.FullName(messageName) {
			messageDescriptor = mt
			return false
		}
		return true
	})
	return messageDescriptor
}
