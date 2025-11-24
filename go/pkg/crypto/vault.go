package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log"

	"golang.org/x/crypto/hkdf"
)

const (
	KEMCiphertextSize = 1568 // ML-KEM-1024 Ciphertext size
	NonceSize         = 12   // AES-GCM Standard Nonce
)

var VaultKey *mlkem.DecapsulationKey1024
var EncodedVaultKey string

func InitVault() {
	var err error
	VaultKey, err = mlkem.GenerateKey1024()
	if err != nil {
		log.Fatalf("Failed to generate vault key err: %v", err)
	}

	EncodedVaultKey = base64.StdEncoding.EncodeToString(VaultKey.EncapsulationKey().Bytes())

	log.Println("Vault configured.")
}

// EncryptForServer (Client Side)
// Generates a shared secret, encapsulates it to the Server's PubKey,
// and encrypts the body.
func EncryptForServer(serverPubKeyBytes []byte, plaintext []byte) ([]byte, []byte, error) {
	// 1. Parse Server Public Key
	ek, err := mlkem.NewEncapsulationKey1024(serverPubKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	// 2. Encapsulate (Create Quantum Shared Secret)
	sharedSecret, kemCT := ek.Encapsulate()

	// 3. Derive AES Key
	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		return nil, nil, err
	}

	// 4. Encrypt Body
	ciphertext, err := symEncrypt(aesKey, plaintext)
	if err != nil {
		return nil, nil, err
	}

	// 5. Pack: [KEM CT] + [AES Ciphertext (includes nonce)]
	finalBlob := append(kemCT, ciphertext...)

	// Return blob and sharedSecret (so client can decrypt response)
	return finalBlob, sharedSecret, nil
}

// DecryptResponse (Client Side)
// Used in the WASM module to allow the JS client to decrypt
// responses from the rtk-query request.
func DecryptResponse(ciphertext []byte, kemSecret []byte) ([]byte, error) {
	// 1. Re-derive the same AES Key used for the request
	// We use the exact same KDF logic as the request phase.
	aesKey, err := deriveAESKey(kemSecret)
	if err != nil {
		return nil, err
	}

	// 2. Decrypt the body using the symmetric key
	// This assumes the server responded using symEncrypt(aesKey, body)
	return symDecrypt(aesKey, ciphertext)
}

// DecryptFromClient (Server Side)
// Decrypts incoming payloads from rtk-query.
func DecryptFromClient(dk *mlkem.DecapsulationKey1024, blob []byte) ([]byte, []byte, error) {
	if len(blob) < KEMCiphertextSize+NonceSize {
		return nil, nil, errors.New("vault: payload too short")
	}

	kemCT := blob[:KEMCiphertextSize]
	aesCT := blob[KEMCiphertextSize:]

	// 1. Decapsulate
	sharedSecret, err := dk.Decapsulate(kemCT)
	if err != nil {
		return nil, nil, err
	}

	// 2. Derive AES Key
	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		return nil, nil, err
	}

	// 3. Decrypt Body
	plaintext, err := symDecrypt(aesKey, aesCT)
	if err != nil {
		return nil, nil, err
	}

	return plaintext, sharedSecret, nil
}

// EncryptForClient (Server Side)
// Encrypts outgoing messages to be received by rtk-query.
func EncryptForClient(sharedSecret []byte, plaintext []byte) ([]byte, error) {
	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		return nil, err
	}

	encryptedBytes, err := symEncrypt(aesKey, plaintext)
	if err != nil {
		return nil, err
	}

	return encryptedBytes, nil
}

// SymEncrypt (Shared utility)
func symEncrypt(key, plaintext []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func symDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("vault: ciphertext too short")
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

func deriveAESKey(input []byte) ([]byte, error) {
	kdf := hkdf.New(sha256.New, input, nil, []byte("AWAYTO_VAULT_V1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}
	return key, nil
}
