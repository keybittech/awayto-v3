package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var (
	DEFAULT_IGNORED_PROTO_FIELDS = []protoreflect.Name{
		protoreflect.Name("state"),
		protoreflect.Name("sizeCache"),
		protoreflect.Name("unknownFields"),
	}
	TitleCase    cases.Caser
	SigningToken []byte
)

const (
	LOGIN_SIGNATURE_NAME = "login_signature_name"
	DefaultPadding       = 5
)

func init() {
	TitleCase = cases.Title(language.Und, cases.NoLower)

	signingToken, err := EnvFile(os.Getenv("SIGNING_TOKEN_FILE"))
	if err != nil {
		println("Failed to get signing token")
		log.Fatal(err)
	}

	SigningToken = []byte(signingToken)
}

type NullConn struct{}

func (n NullConn) Read(b []byte) (int, error)         { return 0, io.EOF }
func (n NullConn) Write(b []byte) (int, error)        { return len(b), nil }
func (n NullConn) Close() error                       { return nil }
func (n NullConn) LocalAddr() net.Addr                { return nil }
func (n NullConn) RemoteAddr() net.Addr               { return nil }
func (n NullConn) SetDeadline(t time.Time) error      { return nil }
func (n NullConn) SetReadDeadline(t time.Time) error  { return nil }
func (n NullConn) SetWriteDeadline(t time.Time) error { return nil }

// NewNullConn returns a new no-op connection
func NewNullConn() net.Conn {
	return NullConn{}
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

func IsUUID(id string) bool {
	if len(id) != 36 {
		return false
	}

	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}

	for i := 0; i < 36; i++ {
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

func EnvFile(loc string) (string, error) {
	envFile, err := os.ReadFile(os.Getenv("PROJECT_DIR") + "/" + loc)
	if err != nil {
		return "", ErrCheck(err)
	}

	return strings.Trim(string(envFile), "\n"), nil
}

func AnonIp(ipAddr string) string {
	ipParts := strings.Split(ipAddr, ".")
	if len(ipParts) != 4 {
		return ""
	}
	ipParts[3] = "0"
	return strings.Join(ipParts, ".")
}

func StringIn(s string, ss []string) bool {
	for _, sv := range ss {
		if sv == s {
			return true
		}
	}
	return false
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
	mac := hmac.New(sha256.New, SigningToken)

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

	mac := hmac.New(sha256.New, SigningToken)
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

func StringsToBitmask(roles []string) int32 {
	var bitmask int32
	for _, role := range roles {
		if bit, ok := types.SiteRoles_value[role]; ok {
			bitmask |= bit
		}
	}
	return bitmask
}
