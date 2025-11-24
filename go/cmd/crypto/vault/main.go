//go:build js && wasm

package main

import (
	"encoding/base64"
	"syscall/js"

	"github.com/keybittech/awayto-v3/go/pkg/crypto"
)

func encryptRequest(this js.Value, args []js.Value) any {
	// args[0]: Server PubKey (base64), args[1]: JSON String
	pubKey, _ := base64.StdEncoding.DecodeString(args[0].String())
	plaintext := []byte(args[1].String())

	blob, secret, err := crypto.EncryptForServer(pubKey, plaintext)
	if err != nil {
		return nil
	}

	return map[string]any{
		"blob":   base64.StdEncoding.EncodeToString(blob),
		"secret": base64.StdEncoding.EncodeToString(secret),
	}
}

func decryptResponse(this js.Value, args []js.Value) any {
	// args[0]: encryptedBlob b64string, args[1]: b64resultSecret string
	if len(args) < 2 {
		return nil
	}

	blob, err := base64.RawStdEncoding.DecodeString(args[0].String())
	if err != nil {
		return nil
	}

	secret, err := base64.RawStdEncoding.DecodeString(args[1].String())
	if err != nil {
		return nil
	}

	plaintext, err := crypto.DecryptResponse(blob, secret)
	if err != nil {
		return nil
	}

	return string(plaintext)
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("pqcEncrypt", js.FuncOf(encryptRequest))
	js.Global().Set("pqcDecrypt", js.FuncOf(decryptResponse))
	<-c
}
