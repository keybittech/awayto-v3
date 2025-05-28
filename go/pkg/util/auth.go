package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

func (c *Cache) GetSessionFromCookie(req *http.Request) *types.ConcurrentUserSession {
	sessionId := GetSessionIdFromCookie(req)
	if sessionId == "" {
		return nil
	}

	session, ok := c.UserSessions.Get(sessionId)
	if !ok {
		return nil
	}

	return session
}

func GetSessionIdFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}

	sessionId, err := VerifySigned("session_id", cookie.Value)
	if err != nil {
		ErrorLog.Printf("could not verify session id signature: %v", err)
		return ""
	}
	return sessionId
}

// GenerateCodeVerifier creates a cryptographically random code verifier
// Length should be between 43-128 characters for RFC 7636 compliance
func GenerateCodeVerifier() string {
	// Using 32 bytes = 43 characters when base64url encoded (32 * 4/3 = ~43)
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateCodeChallenge creates a S256 code challenge from the verifier
func GenerateCodeChallenge(verifier string) string {
	// S256: base64url(sha256(ascii(code_verifier)))
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateSessionId() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func FetchPublicKey() (*rsa.PublicKey, error) {
	resp, err := Get(E_KC_URL, nil)
	if err != nil {
		log.Fatal(ErrCheck(err))
	}

	var result types.KeycloakRealmInfo
	if err := protojson.Unmarshal(resp, &result); err != nil {
		log.Fatal(ErrCheck(err))
	}

	block, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + result.PublicKey + "\n-----END PUBLIC KEY-----"))
	if block == nil {
		log.Fatal(ErrCheck(errors.New("empty pem block")))
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(ErrCheck(err))
	}

	if parsed, ok := pubKey.(*rsa.PublicKey); ok {
		return parsed, nil
	}

	log.Fatal(ErrCheck(errors.New("key could not be parsed")))
	return nil, nil
}

type KeycloakUserWithClaims struct {
	types.KeycloakUser
	jwt.StandardClaims
	ResourceAccess map[string]struct {
		Roles []string `json:"roles,omitempty"`
	} `json:"resource_access,omitempty"`
}

func ValidateToken(tokens *types.OIDCToken, userAgent, timezone, anonIp string) (*types.UserSession, error) {
	parsedToken, err := jwt.ParseWithClaims(tokens.GetAccessToken(), &KeycloakUserWithClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("bad signing method")
		}
		return E_KC_PUBLIC_KEY, nil
	})
	if err != nil {
		return nil, ErrCheck(err)
	}

	if !parsedToken.Valid {
		return nil, ErrCheck(errors.New("invalid token during parse"))
	}

	claims, ok := parsedToken.Claims.(*KeycloakUserWithClaims)
	if !ok {
		return nil, ErrCheck(errors.New("could not parse claims"))
	}

	var roleBits int32
	if clientAccess, clientAccessOk := claims.ResourceAccess[claims.Azp]; clientAccessOk {
		roleBits = StringsToSiteRoles(clientAccess.Roles)
	}

	session := &types.UserSession{
		UserSub:          claims.Subject,
		UserEmail:        claims.Email,
		SubGroupPaths:    claims.Groups,
		RoleBits:         roleBits,
		Timezone:         timezone,
		AnonIp:           anonIp,
		IdToken:          tokens.GetIdToken(),
		AccessToken:      tokens.GetAccessToken(),
		RefreshToken:     tokens.GetRefreshToken(),
		AccessExpiresAt:  time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).UnixNano(),
		RefreshExpiresAt: time.Now().Add(time.Duration(tokens.RefreshExpiresIn) * time.Second).UnixNano(),
	}

	return session, nil
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

func StringsToSiteRoles(roles []string) int32 {
	var bitmask int32
	for _, role := range roles {
		if bit, ok := types.SiteRoles_value[role]; ok {
			bitmask |= bit
		}
	}
	return bitmask
}

func SetForwardingHeaders(r *http.Request) {
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Forwarded-Proto", "https")
	r.Header.Add("X-Forwarded-Host", r.Host)
}

func GetUA(ua string) string {
	if ua == "" {
		return "unknown"
	}
	return ua
}
