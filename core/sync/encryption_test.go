package sync

import (
	"bytes"
	"testing"
)

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "simple text",
			plaintext: []byte("hello world"),
		},
		{
			name:      "empty plaintext",
			plaintext: []byte(""),
		},
		{
			name:      "large data",
			plaintext: bytes.Repeat([]byte("a"), 10000),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if len(encrypted) == 0 {
				t.Error("Encrypt() returned empty ciphertext")
			}

			// Nonce is prepended, so ciphertext should be longer than plaintext
			if len(encrypted) <= len(tt.plaintext) {
				t.Errorf("Encrypt() ciphertext length %d should be > plaintext length %d", len(encrypted), len(tt.plaintext))
			}

			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptor_DifferentNonce(t *testing.T) {
	key := make([]byte, 32)
	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := []byte("same plaintext")

	encrypted1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt() error = %v", err)
	}

	encrypted2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt() error = %v", err)
	}

	// Different nonces should produce different ciphertexts
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Encrypting the same plaintext twice should produce different ciphertexts (different nonces)")
	}
}

func TestEncryptor_InvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{name: "empty key", key: []byte{}},
		{name: "too short key", key: []byte("short")},
		{name: "wrong size key", key: make([]byte, 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEncryptor(tt.key)
			if err == nil {
				t.Errorf("NewEncryptor() with %s should have failed", tt.name)
			}
		})
	}
}

func TestEncryptor_DecryptInvalidCiphertext(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error = %v", err)
	}

	// Tampered ciphertext
	plaintext := []byte("test data")
	ciphertext, _ := encryptor.Encrypt(plaintext)
	ciphertext[0] ^= 0xFF // corrupt the nonce

	_, err = encryptor.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypt() with tampered ciphertext should fail")
	}
}

func TestEncryptor_DecryptTooShort(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error = %v", err)
	}

	// Ciphertext shorter than nonce size
	_, err = encryptor.Decrypt([]byte{0x00})
	if err == nil {
		t.Error("Decrypt() with too-short ciphertext should fail")
	}
}

func TestEncryptor_DifferentKeyDecryptFails(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	encryptor1, _ := NewEncryptor(key1)
	encryptor2, _ := NewEncryptor(key2)

	plaintext := []byte("secret data")
	encrypted, _ := encryptor1.Encrypt(plaintext)

	_, err := encryptor2.Decrypt(encrypted)
	if err == nil {
		t.Error("Decrypt() with wrong key should fail")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("GenerateKey() length = %d, want 32", len(key))
	}

	// Generate two keys and verify they are different
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("second GenerateKey() error = %v", err)
	}

	if bytes.Equal(key, key2) {
		t.Error("GenerateKey() should produce different keys on each call")
	}
}

func TestEncryptor_ValidKeySizes(t *testing.T) {
	keySizes := []int{16, 24, 32}

	for _, size := range keySizes {
		t.Run("key_size_"+string(rune('0'+size/10))+string(rune('0'+size%10)), func(t *testing.T) {
			key := make([]byte, size)
			encryptor, err := NewEncryptor(key)
			if err != nil {
				t.Fatalf("NewEncryptor() with %d-byte key error = %v", size, err)
			}

			plaintext := []byte("test")
			encrypted, err := encryptor.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Error("round-trip failed")
			}
		})
	}
}
