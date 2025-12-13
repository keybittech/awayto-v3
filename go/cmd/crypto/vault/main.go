//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/keybittech/awayto-v3/go/pkg/crypto"
)

func bytesToJS(b []byte) js.Value {
	out := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(out, b)
	return out
}

func jsToBytes(v js.Value) []byte {
	if v.IsNull() || v.IsUndefined() {
		return nil
	}

	lenProp := v.Get("length")
	if lenProp.IsUndefined() || lenProp.IsNull() {
		return nil
	}

	b := make([]byte, lenProp.Int())
	js.CopyBytesToGo(b, v)
	return b
}

func encryptRequest(this js.Value, args []js.Value) any {
	// args[0]: PubKey (Uint8Array), args[1]: SessionId (String), args[2]: Data (Uint8Array)
	if len(args) < 3 {
		return nil
	}

	pubKey := jsToBytes(args[0])
	sid := args[1].String()
	data := jsToBytes(args[2])

	blob, secret, err := crypto.ClientEncrypt(pubKey, data, sid)
	if err != nil {
		return nil
	}

	return map[string]any{
		"blob":   bytesToJS(blob),
		"secret": bytesToJS(secret),
	}
}

func decryptResponse(this js.Value, args []js.Value) any {
	// args[0]: Secret (Uint8Array), args[1]: SessionId (String), args[2]: Blob (Uint8Array)
	if len(args) < 3 {
		return nil
	}

	secret := jsToBytes(args[0])
	sid := args[1].String()
	blob := jsToBytes(args[2])

	if secret == nil || blob == nil {
		return nil
	}

	plaintext, err := crypto.ClientDecrypt(blob, secret, sid)
	if err != nil {
		return nil
	}

	return bytesToJS(plaintext)
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("pqcEncrypt", js.FuncOf(encryptRequest))
	js.Global().Set("pqcDecrypt", js.FuncOf(decryptResponse))
	<-c
}
