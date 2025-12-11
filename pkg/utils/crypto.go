package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// DecryptPassword decrypts AES-256 encrypted password from frontend (CryptoJS format)
func DecryptPassword(encryptedPassword string, passphrase string) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	// CryptoJS prepends "Salted__" to the ciphertext
	if len(ciphertext) < 16 || string(ciphertext[:8]) != "Salted__" {
		return "", errors.New("invalid CryptoJS format")
	}

	// Extract salt (8 bytes after "Salted__")
	salt := ciphertext[8:16]
	ciphertext = ciphertext[16:]

	// Derive key and IV using EVP_BytesToKey (CryptoJS default)
	key, iv := evpBytesToKey(passphrase, salt, 32, 16)

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Decrypt using CBC mode (CryptoJS default)
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// evpBytesToKey implements OpenSSL's EVP_BytesToKey function
// This is what CryptoJS uses for key derivation
func evpBytesToKey(password string, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	var (
		concat   []byte
		lastHash []byte
		totalLen = keyLen + ivLen
	)

	for len(concat) < totalLen {
		hash := md5.New()
		hash.Write(lastHash)
		hash.Write([]byte(password))
		hash.Write(salt)
		lastHash = hash.Sum(nil)
		concat = append(concat, lastHash...)
	}

	return concat[:keyLen], concat[keyLen:totalLen]
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid padding size")
	}

	padLen := int(data[len(data)-1])
	if padLen > len(data) || padLen > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}

	// Verify padding
	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padLen], nil
}

// EncryptPassword encrypts password using CryptoJS-compatible format with environment-based salt
func EncryptPassword(password string, passphrase string, saltFront string, saltBack string) (string, error) {
	// Create deterministic salt from environment values (8 bytes)
	// Combine front and back salt and take first 8 bytes
	combinedSalt := saltFront + saltBack
	salt := []byte(combinedSalt)
	if len(salt) > 8 {
		salt = salt[:8]
	} else if len(salt) < 8 {
		// Pad with zeros if needed
		padding := make([]byte, 8-len(salt))
		salt = append(salt, padding...)
	}

	// Derive key and IV
	key, iv := evpBytesToKey(passphrase, salt, 32, 16)

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Add PKCS7 padding
	plaintext := pkcs7Pad([]byte(password), aes.BlockSize)

	// Encrypt using CBC mode
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	// Prepend "Salted__" and salt
	result := make([]byte, 0, 16+len(ciphertext))
	result = append(result, []byte("Salted__")...)
	result = append(result, salt...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// pkcs7Pad adds PKCS7 padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// HashPassword hashes password using bcrypt for storage
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies bcrypt hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
