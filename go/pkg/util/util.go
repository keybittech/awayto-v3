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
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	DEFAULT_IGNORED_PROTO_FIELDS = []string{"state", "sizeCache", "unknownFields"}
	TitleCase                    cases.Caser
	SigningToken                 []byte
)

const (
	LOGIN_SIGNATURE_NAME  = "login_signature_name"
	ForbiddenResponse     = `{ "error": { "status": 403 } }`
	InternalErrorResponse = `{ "error": { "status": 500 } }`
	DefaultPadding        = 5
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

func ExeTime(name string) (time.Time, func(start time.Time, info string)) {
	return time.Now(), func(start time.Time, info string) {
		AccessLog.Println(name + " " + time.Since(start).String() + " " + info)
	}
}

func WriteSigned(name, unsignedValue string) string {
	mac := hmac.New(sha256.New, SigningToken)
	mac.Write([]byte(name))
	mac.Write([]byte(unsignedValue))
	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature) + unsignedValue
}

func VerifySigned(name, signedValue string) error {
	if len(signedValue) < sha256.Size {
		return errors.New("signed value too small")
	}

	signatureEncoded := signedValue[:base64.StdEncoding.EncodedLen(sha256.Size)]
	signature, err := base64.StdEncoding.DecodeString(signatureEncoded)
	if err != nil {
		return errors.New("invalid base64 signature encoding")
	}

	value := signedValue[base64.StdEncoding.EncodedLen(sha256.Size):]

	mac := hmac.New(sha256.New, SigningToken)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return errors.New("invalid signature equality")
	}

	return nil
}
