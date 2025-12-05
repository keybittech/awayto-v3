package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/hkdf"
)

const (
	KEMCiphertextSize = 1568 // ML-KEM-1024 Ciphertext
	X25519PubKeySize  = 32   // Standard X25519 Public Key
	NonceSize         = 12   // AES-GCM Standard Nonce
	replayWindow      = 5 * time.Minute
	futureTolerance   = 2 * time.Minute
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
	VaultKey.X25519, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate X25519 key: %v", err)
	}

	// 3. Create a combined Public Key for the client
	mlkemPub := VaultKey.MLKEM.EncapsulationKey().Bytes()
	x25519Pub := VaultKey.X25519.PublicKey().Bytes()

	combinedPub := append(mlkemPub, x25519Pub...)
	EncodedVaultKey = base64.StdEncoding.EncodeToString(combinedPub)

	log.Println("Vault configured with Hybrid Encryption (ML-KEM-1024 + X25519).")
}

func packPlaintext(data []byte) []byte {
	buf := make([]byte, 8+len(data))
	binary.BigEndian.PutUint64(buf[:8], uint64(time.Now().UnixNano()))
	copy(buf[8:], data)
	return buf
}

func unpackAndVerifyTimestamp(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, errors.New("verifytimestamp: payload too short")
	}

	tsVal := int64(binary.BigEndian.Uint64(data[:8]))
	ts := time.Unix(0, tsVal)

	if time.Since(ts) > replayWindow {
		return nil, errors.New("verifytimestamp: packet replay detected (old stamp)")
	}

	if time.Until(ts) > futureTolerance {
		return nil, errors.New("verifytimestamp: packet from future (skewed)")
	}

	return data[8:], nil
}

func symEncrypt(key, plaintext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("symenc: could not make new cipher, %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("symenc: could not wrap block, %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("symenc: could not transfer into nonce, %v", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, aad), nil
}

func symDecrypt(key, ciphertext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("symdec: could not make new cipher, %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("symdec: could not wrap block, %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("symdec: ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, actualCiphertext, aad)
}

func deriveAESKey(input []byte) ([]byte, error) {
	kdf := hkdf.New(sha256.New, input, nil, []byte("AWAYTO_VAULT_HYBRID_V1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, fmt.Errorf("deriveaes: could not read kdf into key, %v", err)
	}
	return key, nil
}

// Flow: Client encrypts -> Server decrypts -> Server encrypts -> Client decrypts

// ClientEncrypt (Client Side)
func ClientEncrypt(serverPubKeyBytes, plaintext []byte, sid string) ([]byte, []byte, error) {
	// We expect the key to be ML-KEM-1024 Encapsulation Key size + 32 bytes (X25519)
	expectedSize := KEMCiphertextSize + X25519PubKeySize

	if len(serverPubKeyBytes) != expectedSize {
		return nil, nil, errors.New("client enc: invalid public key size")
	}

	// 1. Split Server Key
	serverKemPubBytes := serverPubKeyBytes[:KEMCiphertextSize]
	serverX25519PubBytes := serverPubKeyBytes[KEMCiphertextSize:]

	// 2. ML-KEM Encapsulation
	ek, err := mlkem.NewEncapsulationKey1024(serverKemPubBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: mlkem generate bad %v", err)
	}
	sharedSecretKEM, kemCT := ek.Encapsulate()

	// 3. X25519 Key Exchange
	// Generate ephemeral client key
	clientEphemeralKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: x25519 generate bad %v", err)
	}

	// Parse server's X25519 key
	serverX25519Pub, err := ecdh.X25519().NewPublicKey(serverX25519PubBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: x25519 parse bad %v", err)
	}

	// Compute ECDH secret
	sharedSecretX25519, err := clientEphemeralKey.ECDH(serverX25519Pub)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: ecdh compute bad %v", err)
	}

	// 4. Combine Secrets for Hybrid Security
	combinedSecret := append(sharedSecretKEM, sharedSecretX25519...)

	// 5. Derive AES Key
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: %v", err)
	}

	// 6. Encrypt Body
	packet := packPlaintext(plaintext)
	aad := []byte(sid)

	ciphertext, err := symEncrypt(aesKey, packet, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("client enc: %v", err)
	}

	// 7. Pack Blob: [KEM CT] + [Client Ephemeral X25519 Pub] + [AES Ciphertext]
	finalBlob := make([]byte, 0, len(kemCT)+X25519PubKeySize+len(ciphertext))
	finalBlob = append(finalBlob, kemCT...)
	finalBlob = append(finalBlob, clientEphemeralKey.PublicKey().Bytes()...)
	finalBlob = append(finalBlob, ciphertext...)

	// Return blob and the COMBINED secret (needed to decrypt response)
	return finalBlob, combinedSecret, nil
}

// ServerDecrypt (Server Side)
func ServerDecrypt(dk *HybridPrivateKey, blob []byte, sid string) ([]byte, []byte, error) {
	minSize := KEMCiphertextSize + X25519PubKeySize + NonceSize
	if len(blob) < minSize {
		return nil, nil, errors.New("server dec: payload too short")
	}

	// 1. Slice the blob
	// [KEM CT (1568)] [X25519 Pub (32)] [AES Body...]
	kemCT := blob[:KEMCiphertextSize]
	clientX25519PubBytes := blob[KEMCiphertextSize : KEMCiphertextSize+X25519PubKeySize]
	aesCT := blob[KEMCiphertextSize+X25519PubKeySize:]

	// 2. ML-KEM Decapsulation
	sharedSecretKEM, err := dk.MLKEM.Decapsulate(kemCT)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: mlkem decap bad, %v", err)
	}

	// 3. X25519 Decapsulation (ECDH)
	clientPub, err := ecdh.X25519().NewPublicKey(clientX25519PubBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: x25519 pub key bad creation, %v", err)
	}

	sharedSecretX25519, err := dk.X25519.ECDH(clientPub)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: ecdh bad, %v", err)
	}

	// 4. Reconstruct Combined Secret
	combinedSecret := append(sharedSecretKEM, sharedSecretX25519...)

	// 5. Derive AES Key
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: %v", err)
	}

	// 6. Decrypt Body
	aad := []byte(sid)
	decryptedPacket, err := symDecrypt(aesKey, aesCT, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: %v", err)
	}

	// 7. Verify timestamp
	plaintext, err := unpackAndVerifyTimestamp(decryptedPacket)
	if err != nil {
		return nil, nil, fmt.Errorf("server dec: %v", err)
	}

	return plaintext, combinedSecret, nil
}

// ServerEncrypt (Server Side)
func ServerEncrypt(combinedSecret, plaintext []byte, sid string) ([]byte, error) {
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, fmt.Errorf("server enc: %v", err)
	}

	aad := []byte(sid)

	return symEncrypt(aesKey, plaintext, aad)
}

// ClientDecrypt (Client Side)
func ClientDecrypt(ciphertext, combinedSecret []byte, sid string) ([]byte, error) {
	// Re-derive AES Key using the hybrid secret
	aesKey, err := deriveAESKey(combinedSecret)
	if err != nil {
		return nil, fmt.Errorf("client dec: %v", err)
	}

	aad := []byte(sid)

	return symDecrypt(aesKey, ciphertext, aad)
}
