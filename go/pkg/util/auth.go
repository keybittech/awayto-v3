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
	"net/url"
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
	return cookie.Value
}

func (c *Cache) RefreshAccessToken(req *http.Request) error {
	sessionId := GetSessionIdFromCookie(req)
	if sessionId == "" {
		return ErrCheck(errors.New("no session id to refresh token"))
	}

	session, ok := c.UserSessions.Get(sessionId)
	if !ok {
		return ErrCheck(errors.New("failed to get user session to refresh token"))
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {E_KC_USER_CLIENT},
		"client_secret": {E_KC_USER_CLIENT_SECRET},
		"refresh_token": {session.GetRefreshToken()},
	}

	SetForwardingHeaders(req)

	resp, err := PostFormData(E_KC_OPENID_TOKEN_URL, req.Header, data)
	if err != nil {
		return ErrCheck(err)
	}

	var tokens types.OIDCToken
	if err := protojson.Unmarshal(resp, &tokens); err != nil {
		return ErrCheck(err)
	}

	err = c.ValidateToken(&tokens, sessionId, req.Header.Get("X-Tz"), AnonIp(req.RemoteAddr))
	if err != nil {
		return ErrCheck(err)
	}

	return nil
}

func GenerateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateCodeChallenge(verifier string) string {
	// For simplicity, using plain challenge. In production, use S256
	return verifier
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

func (c *Cache) ValidateToken(tokens *types.OIDCToken, sessionId, timezone, anonIp string) error {
	parsedToken, err := jwt.ParseWithClaims(tokens.GetAccessToken(), &KeycloakUserWithClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("bad signing method")
		}
		return E_KC_PUBLIC_KEY, nil
	})
	if err != nil {
		return ErrCheck(err)
	}

	if !parsedToken.Valid {
		return ErrCheck(errors.New("invalid token during parse"))
	}

	claims, ok := parsedToken.Claims.(*KeycloakUserWithClaims)
	if !ok {
		return ErrCheck(errors.New("could not parse claims"))
	}

	var roleBits int32
	if clientAccess, clientAccessOk := claims.ResourceAccess[claims.Azp]; clientAccessOk {
		roleBits = StringsToSiteRoles(clientAccess.Roles)
	}

	session := &types.UserSession{
		UserSub:       claims.Subject,
		UserEmail:     claims.Email,
		SubGroupPaths: claims.Groups,
		RoleBits:      roleBits,
		ExpiresAt:     claims.ExpiresAt,
		Timezone:      timezone,
		AnonIp:        anonIp,
	}

	session.Token = tokens.AccessToken
	session.RefreshToken = tokens.RefreshToken
	session.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).UnixNano()
	session.RefreshExpiresAt = time.Now().Add(time.Duration(tokens.RefreshExpiresIn) * time.Second).UnixNano()

	c.UserSessions.Store(sessionId, types.NewConcurrentUserSession(session))

	return nil
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
