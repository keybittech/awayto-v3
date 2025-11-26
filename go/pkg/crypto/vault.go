package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/hkdf"
)

const (
	// Sizes for parsing the blob
	KEMCiphertextSize = 1568 // ML-KEM-1024 Ciphertext
	X25519PubKeySize  = 32   // Standard X25519 Public Key
	NonceSize         = 12   // AES-GCM Standard Nonce
)

// HybridPrivateKey holds both the Post-Quantum and Classical keys
type HybridPrivateKey struct {
	MLKEM  *mlkem.DecapsulationKey1024
	X25519 *ecdh.PrivateKey
}

var VaultKey *HybridPrivateKey
var EncodedVaultKey string

func InitVault() {
	var err error
	VaultKey = &HybridPrivateKey{}

	// 1. Generate Post-Quantum Key (ML-KEM)
	VaultKey.MLKEM, err = mlkem.GenerateKey1024()
	if err != nil {
		log.Fatalf("Failed to generate ML-KEM key: %v", err)
	}

	// 2. Generate Classical Key (X25519)
	// This provides a safety net if ML-KEM is broken in the future.
	VaultKey.X25519, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate X25519 key: %v", err)
	}

	// 3. Create a combined Public Key for the client
	// Format: [ML-KEM PubKey] + [X25519 PubKey]
	mlkemPub := VaultKey.MLKEM.EncapsulationKey().Bytes()
	x25519Pub := VaultKey.X25519.PublicKey().Bytes()

	combinedPub := append(mlkemPub, x25519Pub...)
	EncodedVaultKey = base64.StdEncoding.EncodeToString(combinedPub)

	log.Println("Vault configured with Hybrid Encryption (ML-KEM-1024 + X25519).")
}

// EncryptForServer (Client Side)
// Performs a Hybrid Key Exchange.
func EncryptForServer(serverPubKeyBytes []byte, plaintext []byte) ([]byte, []byte, error) {
	// We expect the key to be ML-KEM-1024 Encapsulation Key size + 32 bytes (X25519)
	kemKeySize := 1568 // Size of ML-KEM-1024 Public Key
	expectedSize := kemKeySize + X25519PubKeySize

	if len(serverPubKeyBytes) != expectedSize {
		fmt.Println("did error here")
		return nil, nil, errors.New("invalid server public key size")
	}

	// 1. Split Server Key
	serverKemPubBytes := serverPubKeyBytes[:kemKeySize]
	serverX25519PubBytes := serverPubKeyBytes[kemKeySize:]

	// 2. ML-KEM Encapsulation
	ek, err := mlkem.NewEncapsulationKey1024(serverKemPubBytes)
	if err != nil {
		return nil, nil, err
	}
	sharedSecretKEM, kemCT := ek.Encapsulate()

	// 3. X25519 Key Exchange
	// Generate ephemeral client key
	clientEphemeralKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	// Parse server's X25519 key
	serverX25519Pub, err := ecdh.X25519().NewPublicKey(serverX25519PubBytes)
	if err != nil {
		return nil, nil, err
	}
	// Compute ECDH secret
	sharedSecretX25519, err := clientEphemeralKey.ECDH(serverX25519Pub)
	if err != nil {
		return nil, nil, err
	}

	// 4. Combine Secrets for Hybrid Security
	// Concatenate (KEM Secret || X25519 Secret)
	combinedSecret := append(sharedSecretKEM, sharedSecretX25519...)

	// 5. Derive AES Key
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, nil, err
	}

	// 6. Encrypt Body
	ciphertext, err := symEncrypt(aesKey, plaintext)
	if err != nil {
		return nil, nil, err
	}

	// 7. Pack Blob: [KEM CT] + [Client Ephemeral X25519 Pub] + [AES Ciphertext]
	finalBlob := make([]byte, 0, len(kemCT)+X25519PubKeySize+len(ciphertext))
	finalBlob = append(finalBlob, kemCT...)
	finalBlob = append(finalBlob, clientEphemeralKey.PublicKey().Bytes()...)
	finalBlob = append(finalBlob, ciphertext...)

	// Return blob and the COMBINED secret (needed to decrypt response)
	return finalBlob, combinedSecret, nil
}

// DecryptResponse (Client Side)
func DecryptResponse(ciphertext []byte, combinedSecret []byte) ([]byte, error) {
	// Re-derive AES Key using the hybrid secret
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, err
	}
	return symDecrypt(aesKey, ciphertext)
}

// DecryptFromClient (Server Side)
func DecryptFromClient(dk *HybridPrivateKey, blob []byte) ([]byte, []byte, error) {
	minSize := KEMCiphertextSize + X25519PubKeySize + NonceSize
	if len(blob) < minSize {
		return nil, nil, errors.New("vault: payload too short")
	}

	// 1. Slice the blob
	// [KEM CT (1568)] [X25519 Pub (32)] [AES Body...]
	kemCT := blob[:KEMCiphertextSize]
	clientX25519PubBytes := blob[KEMCiphertextSize : KEMCiphertextSize+X25519PubKeySize]
	aesCT := blob[KEMCiphertextSize+X25519PubKeySize:]

	// 2. ML-KEM Decapsulation
	sharedSecretKEM, err := dk.MLKEM.Decapsulate(kemCT)
	if err != nil {
		return nil, nil, err
	}

	// 3. X25519 Decapsulation (ECDH)
	clientPub, err := ecdh.X25519().NewPublicKey(clientX25519PubBytes)
	if err != nil {
		return nil, nil, err
	}
	sharedSecretX25519, err := dk.X25519.ECDH(clientPub)
	if err != nil {
		return nil, nil, err
	}

	// 4. Reconstruct Combined Secret
	combinedSecret := append(sharedSecretKEM, sharedSecretX25519...)

	// 5. Derive AES Key
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, nil, err
	}

	// 6. Decrypt Body
	plaintext, err := symDecrypt(aesKey, aesCT)
	if err != nil {
		return nil, nil, err
	}

	return plaintext, combinedSecret, nil
}

// EncryptForClient (Server Side)
func EncryptForClient(combinedSecret []byte, plaintext []byte) ([]byte, error) {
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, err
	}

	return symEncrypt(aesKey, plaintext)
}

// --- Shared Utilities (Unchanged logic, just helpers) ---

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
	// Updated info string to reflect protocol change
	kdf := hkdf.New(sha256.New, input, nil, []byte("AWAYTO_VAULT_HYBRID_V1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}
	return key, nil
}
