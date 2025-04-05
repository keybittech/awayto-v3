package util

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type IdStruct struct {
	Id string `json:"id"`
}

type HandlerOptions struct {
	ServiceMethodURL  string
	SiteRole          string
	Pattern           string
	CacheType         types.CacheType
	CacheDuration     int32
	NoLogFields       []string
	MultipartRequest  bool
	MultipartResponse bool
	Throttle          int32
}

func ParseHandlerOptions(md protoreflect.MethodDescriptor) *HandlerOptions {
	parsedOptions := &HandlerOptions{}

	if md.Input().Fields().Len() > 0 {
		for i := 0; i < md.Input().Fields().Len(); i++ {
			field := md.Input().Fields().ByNumber(protowire.Number(i + 1))

			if proto.HasExtension(field.Options(), types.E_Nolog) {
				parsedOptions.NoLogFields = append(parsedOptions.NoLogFields, TitleCase.String(string(field.Name())))
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

	if proto.HasExtension(inputOpts, types.E_SiteRole) {
		parsedOptions.SiteRole = fmt.Sprint(proto.GetExtension(inputOpts, types.E_SiteRole))
	}

	if proto.HasExtension(inputOpts, types.E_Cache) {
		parsedOptions.CacheType = proto.GetExtension(inputOpts, types.E_Cache).(types.CacheType)
	}

	if proto.HasExtension(inputOpts, types.E_CacheDuration) {
		parsedOptions.CacheDuration = proto.GetExtension(inputOpts, types.E_CacheDuration).(int32)
	}

	if proto.HasExtension(inputOpts, types.E_Throttle) {
		parsedOptions.Throttle = proto.GetExtension(inputOpts, types.E_Throttle).(int32)
	}

	if proto.HasExtension(inputOpts, types.E_MultipartRequest) {
		parsedOptions.MultipartRequest = proto.GetExtension(inputOpts, types.E_MultipartRequest).(bool)
	}

	if proto.HasExtension(inputOpts, types.E_MultipartResponse) {
		parsedOptions.MultipartResponse = proto.GetExtension(inputOpts, types.E_MultipartResponse).(bool)
	}

	return parsedOptions
}

func parseTag(field reflect.StructField, fieldName string) string {
	tagValue := field.Tag.Get(fieldName)
	if tagValue == "" {
		tagValue = strings.ToLower(field.Name)
	} else if commaIndex := strings.Index(tagValue, ","); commaIndex != -1 {
		tagValue = tagValue[:commaIndex]
	}
	return tagValue
}

func ParseProtoQueryParams(pbVal reflect.Value, queryParams url.Values) {
	if len(queryParams) > 0 {
		for i := 0; i < pbVal.NumField(); i++ {
			f := pbVal.Type().Field(i)
			jsonTag := parseTag(f, "json")

			if values, ok := queryParams[jsonTag]; ok && len(values) > 0 {
				fv := pbVal.Field(i)
				if fv.IsValid() && fv.CanSet() && fv.Kind() == reflect.String {
					fv.SetString(values[0])
				}
			}
		}
	}
}

func ParseProtoPathParams(pbVal reflect.Value, methodParameters, requestParameters []string) {
	if len(methodParameters) > 0 && len(methodParameters) == len(requestParameters) {
		for i := 0; i < len(methodParameters); i++ {
			mp := methodParameters[i]

			if strings.HasPrefix(mp, "{") {
				mp = strings.TrimLeft(mp, "{")
				mp = strings.TrimRight(mp, "}")

				for k := 0; k < pbVal.NumField(); k++ {
					f := pbVal.Type().Field(k)
					jsonTag := parseTag(f, "json")

					if strings.Split(jsonTag, ",")[0] == mp {
						fv := pbVal.Field(k)
						if fv.IsValid() && fv.CanSet() && fv.Kind() == reflect.String {
							fv.SetString(requestParameters[i])
						}
					}
				}
			}
		}
	}
}
