//go:build js && wasm

package main

import (
	"encoding/base64"
	"strings"
	"syscall/js"

	"github.com/keybittech/awayto-v3/go/pkg/crypto"
)

func encryptRequest(this js.Value, args []js.Value) any {
	// args[0]: Server PubKey (base64), args[1]: sessionId, args[2]: data String/bytes
	if len(args) < 3 {
		return nil
	}

	pubKey, _ := base64.StdEncoding.DecodeString(args[0].String())

	sid := args[1].String()

	var plaintext []byte
	// handle (json) strings or (file) bytes
	if args[2].Type() == js.TypeString {
		plaintext = []byte(args[2].String())
	} else {
		length := args[2].Get("length").Int()
		plaintext = make([]byte, length)
		js.CopyBytesToGo(plaintext, args[2])
	}

	blob, secret, err := crypto.ClientEncrypt(pubKey, plaintext, sid)
	if err != nil {
		return nil
	}

	return map[string]any{
		"blob":   base64.StdEncoding.EncodeToString(blob),
		"secret": base64.StdEncoding.EncodeToString(secret),
	}
}

func decryptResponse(this js.Value, args []js.Value) any {
	// args[0]: vault secret string, args[1]: sessionId, args[2]: data b64
	if len(args) < 3 {
		return nil
	}

	secretStr := strings.TrimSpace(args[0].String())
	sid := args[1].String()
	blobStr := strings.TrimSpace(args[2].String())

	blob, err := base64.StdEncoding.DecodeString(blobStr)
	if err != nil {
		return nil
	}

	secret, err := base64.StdEncoding.DecodeString(secretStr)
	if err != nil {
		return nil
	}

	plaintext, err := crypto.ClientDecrypt(blob, secret, sid)
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
