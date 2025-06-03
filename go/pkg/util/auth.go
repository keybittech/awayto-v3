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
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	if E_KC_PUBLIC_KEY != nil {
		return E_KC_PUBLIC_KEY, nil
	}

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
		E_KC_PUBLIC_KEY = parsed
		return parsed, nil
	}

	log.Fatal(ErrCheck(errors.New("key could not be parsed")))
	return nil, nil
}

func GetValidTokenChallenge(req *http.Request, code, codeVerifier, ua, tz, ip string) (*types.UserSession, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {E_KC_USER_CLIENT},
		"client_secret": {E_KC_USER_CLIENT_SECRET},
		"redirect_uri":  {E_APP_HOST_URL + "/auth/callback"},
		"code":          {code},
		"code_verifier": {codeVerifier},
	}
	return FetchAndValidateToken(req, data, ua, tz, ip)
}

func GetValidTokenRefresh(req *http.Request, refreshToken, ua, tz, ip string) (*types.UserSession, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {E_KC_USER_CLIENT},
		"client_secret": {E_KC_USER_CLIENT_SECRET},
		"refresh_token": {refreshToken},
	}
	return FetchAndValidateToken(req, data, ua, tz, ip)
}

// Data here should provide a refresh token or code challenge with verifier
func FetchAndValidateToken(req *http.Request, data url.Values, ua, tz, ip string) (*types.UserSession, error) {
	SetForwardingHeaders(req)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := PostFormData(req.Context(), E_KC_OPENID_TOKEN_URL, req.Header, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, ErrCheck(err)
	}

	var tokens types.OIDCToken
	if err := protojson.Unmarshal(resp, &tokens); err != nil {
		return nil, ErrCheck(err)
	}

	return ValidateToken(&tokens, ua, tz, ip)
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
		UserAgent:        userAgent,
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

type ScoreValueTypes interface{ string | int32 | int64 }

// Return number of non-matching pairs
func ScoreValues[T ScoreValueTypes](values [][]T) int8 {
	var score int8
	for _, row := range values {
		var prev T
		for i, value := range row {
			if i == 0 {
				prev = value
				continue
			}

			if value != prev {
				score++
			}

			prev = value
		}
	}
	return score
}
