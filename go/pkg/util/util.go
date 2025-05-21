package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var (
	signingToken []byte

	DEFAULT_IGNORED_PROTO_FIELDS = []protoreflect.Name{
		protoreflect.Name("state"),
		protoreflect.Name("sizeCache"),
		protoreflect.Name("unknownFields"),
	}
	TitleCase cases.Caser = cases.Title(language.Und, cases.NoLower)
)

const (
	LOGIN_SIGNATURE_NAME = "login_signature_name"
	DefaultPadding       = 5
)

func loadSigningToken() {
	signToken, err := GetEnvFilePath("SIGNING_TOKEN_FILE", 128)
	if err != nil {
		println("Failed to get signing token")
		log.Fatal(err)
	}

	signingToken = []byte(signToken)
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func Atoi32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, ErrCheck(err)
	}
	return int32(i), nil
}

func Itoi32(i int) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, ErrCheck(errors.New("int32 conversion overflowed"))
	}

	return int32(i), nil
}

func Itoui32(i int) (uint32, error) {
	if i > math.MaxUint32 || i < 0 {
		return 0, ErrCheck(errors.New("uint32 conversion overflowed"))
	}

	return uint32(i), nil
}

func I64to32(i int64) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, ErrCheck(errors.New("unt64 > int32 conversion overflowed"))
	}

	return int32(i), nil
}

type ConvertibleFromStringBytes interface {
	string | int | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func ConvertStringBytes[T ConvertibleFromStringBytes](b []byte, trim ...[]byte) T {
	var zeroVal T

	if len(trim) > 0 {
		trimSlice := trim[0]
		b = bytes.Trim(b, string(trimSlice))
	}

	s := string(b)

	switch any(zeroVal).(type) {
	case string:
		return any(s).(T)
	case int:
		val, err := strconv.Atoi(s)
		if err != nil {
			return zeroVal
		}
		return any(val).(T)
	case int8:
		val8, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return zeroVal
		}
		return any(int8(val8)).(T)
	case uint8:
		valU8, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return zeroVal
		}
		return any(uint8(valU8)).(T)
	case int16:
		val16, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return zeroVal
		}
		return any(int16(val16)).(T)
	case uint16:
		valU16, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return zeroVal
		}
		return any(uint16(valU16)).(T)
	case int32:
		val32, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return zeroVal
		}
		return any(int32(val32)).(T)
	case uint32:
		valU32, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return zeroVal
		}
		return any(uint32(valU32)).(T)
	case int64:
		val64, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return zeroVal
		}
		return any(val64).(T)
	case uint64:
		valU64, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return zeroVal
		}
		return any(valU64).(T)
	}
	return zeroVal
}

func IsUUID(id string) bool {
	if len(id) != 36 {
		return false
	}

	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}

	for i := range 36 {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}

		b := id[i]
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')) {
			return false
		}
	}

	return true
}

func IsEpoch(id string) bool {
	if len(id) == 0 {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func PaddedLen(padTo int, length int) string {
	strLen := strconv.Itoa(length)
	for len(strLen) < padTo {
		strLen = "0" + strLen
	}
	return strLen
}

func AnonIp(ipAddr string) string {
	ipParts := strings.Split(ipAddr, ".")
	if len(ipParts) != 4 {
		return ""
	}
	ipParts[3] = "0"
	return strings.Join(ipParts, ".")
}

func StringOut(s string, ss []string) []string {
	if len(ss) == 0 {
		return ss
	}
	ns := make([]string, 0, len(ss)-1)
	for _, cs := range ss {
		if cs == s {
			continue
		}
		ns = append(ns, cs)
	}
	return ns
}

func WriteSigned(name, unsignedValue string) (string, error) {
	mac := hmac.New(sha256.New, signingToken)

	_, err := mac.Write([]byte(name))
	if err != nil {
		return "", ErrCheck(errors.New("invalid base64 signature encoding"))
	}

	_, err = mac.Write([]byte(unsignedValue))
	if err != nil {
		return "", ErrCheck(errors.New("invalid base64 signature encoding"))
	}

	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature) + unsignedValue, nil
}

func VerifySigned(name, signedValue string) (string, error) {
	if len(signedValue) < sha256.Size {
		return "", ErrCheck(errors.New("signed value too small"))
	}

	signatureEncoded := signedValue[:base64.StdEncoding.EncodedLen(sha256.Size)]
	signature, err := base64.StdEncoding.DecodeString(signatureEncoded)
	if err != nil {
		return "", ErrCheck(errors.New("invalid base64 signature encoding"))
	}

	value := signedValue[base64.StdEncoding.EncodedLen(sha256.Size):]

	mac := hmac.New(sha256.New, signingToken)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return "", ErrCheck(errors.New("invalid signature equality"))
	}

	return value, nil
}

func CookieExpired(req *http.Request) bool {
	cookie, err := req.Cookie("valid_signature")
	if err == nil && cookie.Value != "" {
		expiresAtStr, err := VerifySigned(LOGIN_SIGNATURE_NAME, cookie.Value)
		if err == nil {
			expiresAt, parseErr := strconv.ParseInt(expiresAtStr, 10, 64)
			if parseErr == nil && time.Now().Unix() < expiresAt {
				return false
			}
		}
	}
	return true
}

func StringsToSiteRoles(roles []string) types.SiteRoles {
	var bitmask int32
	for _, role := range roles {
		if bit, ok := types.SiteRoles_value[role]; ok {
			bitmask |= bit
		}
	}
	return types.SiteRoles(bitmask)
}
