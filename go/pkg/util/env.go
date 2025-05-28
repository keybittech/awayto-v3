package util

import (
	"bufio"
	"crypto/rsa"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	httpPortFlag       = flag.Int("httpPort", 7080, "Server HTTP port")
	httpsPortFlag      = flag.Int("httpsPort", 7443, "Server HTTPS port")
	rateLimitFlag      = flag.Int("rateLimit", 20, "Requests per second")
	rateLimitBurstFlag = flag.Int("rateLimitBurst", 20, "Requests per second burst")
	logLevelFlag       = flag.String("logLevel", "", "Set the logging to empty or debug")

	envVarStrs map[string]string

	E_GO_ENVFILE_LOC = os.Getenv("GO_ENVFILE_LOC")
	E_UNIX_SOCK_DIR  = os.Getenv("UNIX_SOCK_DIR")

	E_APP_HOST_URL, E_APP_HOST_NAME, E_API_PATH, E_BINARY_NAME, E_CERT_LOC, E_CERT_KEY_LOC, E_DB_DRIVER, E_KC_USER_CLIENT_SECRET,
	E_KC_OPENID_TOKEN_URL, E_KC_OPENID_AUTH_URL, E_KC_OPENID_LOGOUT_URL, E_KC_API_CLIENT, E_KC_USER_CLIENT, E_KC_REALM, E_KC_INTERNAL,
	E_KC_URL, E_KC_ADMIN_URL, E_LOG_LEVEL, E_LOG_DIR, E_PG_WORKER, E_PG_DB, E_PROJECT_DIR, E_REDIS_URL,
	E_TS_DEV_SERVER_URL, E_UNIX_SOCK_FILE, E_UNIX_PATH string

	E_API_PATH_LEN, E_GO_HTTP_PORT, E_GO_HTTPS_PORT, E_RATE_LIMIT, E_RATE_LIMIT_BURST int

	E_KC_PUBLIC_KEY *rsa.PublicKey
)

func ParseEnv() {
	loadEnvFile()

	E_APP_HOST_URL = ParseEnvFileVar[string]("APP_HOST_URL")
	E_APP_HOST_NAME = ParseEnvFileVar[string]("APP_HOST_NAME")
	E_API_PATH = ParseEnvFileVar[string]("API_PATH")
	E_API_PATH_LEN = len(E_API_PATH)
	E_BINARY_NAME = ParseEnvFileVar[string]("BINARY_NAME")
	E_CERT_LOC = ParseEnvFileVar[string]("CERT_LOC")
	E_CERT_KEY_LOC = ParseEnvFileVar[string]("CERT_KEY_LOC")
	E_DB_DRIVER = ParseEnvFileVar[string]("DB_DRIVER")
	E_GO_HTTP_PORT = ParseEnvFileVar[int]("GO_HTTP_PORT")
	E_GO_HTTPS_PORT = ParseEnvFileVar[int]("GO_HTTPS_PORT")
	E_KC_API_CLIENT = ParseEnvFileVar[string]("KC_API_CLIENT")
	E_KC_USER_CLIENT = ParseEnvFileVar[string]("KC_USER_CLIENT")
	E_KC_REALM = ParseEnvFileVar[string]("KC_REALM")
	E_KC_INTERNAL = ParseEnvFileVar[string]("KC_INTERNAL")
	E_LOG_LEVEL = ParseEnvFileVar[string]("LOG_LEVEL")
	E_LOG_DIR = ParseEnvFileVar[string]("LOG_DIR")
	E_PG_WORKER = ParseEnvFileVar[string]("PG_WORKER")
	E_PG_DB = ParseEnvFileVar[string]("PG_DB")
	E_PROJECT_DIR = ParseEnvFileVar[string]("PROJECT_DIR")
	E_RATE_LIMIT = ParseEnvFileVar[int]("RATE_LIMIT")
	E_RATE_LIMIT_BURST = ParseEnvFileVar[int]("RATE_LIMIT_BURST")
	E_REDIS_URL = ParseEnvFileVar[string]("REDIS_URL")
	E_TS_DEV_SERVER_URL = ParseEnvFileVar[string]("TS_DEV_SERVER_URL")
	E_UNIX_SOCK_FILE = ParseEnvFileVar[string]("UNIX_SOCK_FILE")
	E_UNIX_PATH = filepath.Join(E_UNIX_SOCK_DIR, E_UNIX_SOCK_FILE)
	E_KC_URL = E_KC_INTERNAL + "/realms/" + E_KC_REALM
	E_KC_ADMIN_URL = E_KC_INTERNAL + "/admin/realms/" + E_KC_REALM
	E_KC_OPENID_AUTH_URL = E_APP_HOST_URL + "/auth/realms/" + E_KC_REALM + "/protocol/openid-connect/auth"
	E_KC_OPENID_LOGOUT_URL = E_APP_HOST_URL + "/auth/realms/" + E_KC_REALM + "/protocol/openid-connect/logout"
	E_KC_OPENID_TOKEN_URL = E_KC_INTERNAL + "/realms/" + E_KC_REALM + "/protocol/openid-connect/token"

	publicKey, err := FetchPublicKey()
	if err != nil {
		log.Fatalf("could not get kc public key, err: %v", err)
	}
	E_KC_PUBLIC_KEY = publicKey

	userClientSecret, err := GetEnvFilePath("KC_USER_CLIENT_SECRET_FILE", 128)
	if err != nil {
		log.Fatalf("could not get user client secret, err: %v", err)
	}
	E_KC_USER_CLIENT_SECRET = userClientSecret

	flag.Parse()

	E_GO_HTTP_PORT = *httpPortFlag
	E_GO_HTTPS_PORT = *httpsPortFlag
	E_RATE_LIMIT = *rateLimitFlag
	E_RATE_LIMIT_BURST = *rateLimitBurstFlag
	E_LOG_LEVEL = *logLevelFlag

	loadSigningToken()
	makeLoggers()
}

func loadEnvFile() {
	if E_GO_ENVFILE_LOC == "" {
		log.Fatal("GO_ENVFILE_LOC must be set")
	}

	envVarStrs = make(map[string]string)

	cleanEnvPath := filepath.Clean(E_GO_ENVFILE_LOC)
	if strings.Contains(cleanEnvPath, "..") {
		log.Fatalf("invalid env path: path traversal attempt detected, %s", E_GO_ENVFILE_LOC)
	}

	file, err := os.Open(cleanEnvPath)
	if err != nil {
		log.Fatalf("could not load env file, path: %s", E_GO_ENVFILE_LOC)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		varParts := strings.SplitN(scanner.Text(), "=", 2)
		if len(varParts) != 2 {
			continue
		}
		envVarStrs[varParts[0]] = varParts[1]
	}

	// This will parse vars repeatedly 10 times as we can't depend on
	// ordering of env file or number of vars used in each other var
	for range 10 {
		for i, ev := range envVarStrs {
			var replaced string = ev
			matches := regexp.MustCompile(`\${[^}]+}`).FindAll([]byte(replaced), 10)

			for _, match := range matches {
				m := strings.Trim(string(match[1:]), "{}")
				replaced = strings.Replace(replaced, string(match), envVarStrs[m], 1)
			}

			envVarStrs[i] = replaced
		}
	}
}

func ParseEnvFileVar[T ConvertibleFromStringBytes](name string) T {
	var empty T
	targetProp, ok := envVarStrs[name]
	if !ok || targetProp == "" {
		return empty
	}
	return ConvertStringBytes[T]([]byte(targetProp), []byte{'"'})
}

func GetEnvFilePath(envFilePath string, byteSize uint16) (string, error) {
	file, err := GetCleanPath(envVarStrs[envFilePath], os.O_RDONLY)
	if err != nil {
		return "", ErrCheck(err)
	}

	fileBytes := make([]byte, byteSize)
	_, err = file.Read(fileBytes)
	if err != nil {
		return "", ErrCheck(err)
	}

	fileStr := string(fileBytes)

	return strings.Trim(fileStr, "\n"), nil
}
