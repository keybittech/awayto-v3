//go:build js && wasm

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/keybittech/awayto-v3/go/pkg/crypto"
)

func encryptRequest(this js.Value, args []js.Value) any {
	// args[0]: Server PubKey (base64), args[1]: JSON String
	pubKey, _ := base64.StdEncoding.DecodeString(args[0].String())

	var plaintext []byte

	// handle (json) strings or (file) bytes
	if args[1].Type() == js.TypeString {
		plaintext = []byte(args[1].String())
	} else {
		length := args[1].Get("length").Int()
		plaintext = make([]byte, length)
		js.CopyBytesToGo(plaintext, args[1])
	}

	blob, secret, err := crypto.EncryptForServer(pubKey, plaintext)
	if err != nil {
		fmt.Println("WASM: Into encryption issue")
		return nil
	}

	return map[string]any{
		"blob":   base64.StdEncoding.EncodeToString(blob),
		"secret": base64.StdEncoding.EncodeToString(secret),
	}
}

func decryptResponse(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return nil
	}

	// Ensure inputs are clean
	blobStr := strings.TrimSpace(args[0].String())
	secretStr := strings.TrimSpace(args[1].String())

	// Robust Decode: Tries Standard (with padding), falls back to Raw
	decodeSafe := func(in string) ([]byte, error) {
		if b, err := base64.StdEncoding.DecodeString(in); err == nil {
			return b, nil
		}
		if b, err := base64.RawStdEncoding.DecodeString(in); err == nil {
			return b, nil
		}
		return nil, errors.New("WASM: From server decoding failed")
	}

	blob, err := decodeSafe(blobStr)
	if err != nil {
		return nil
	}

	secret, err := decodeSafe(secretStr)
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
