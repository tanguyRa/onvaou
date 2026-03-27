package crypto

import (
	"encoding/base64"
	"testing"
)

// Generate a valid test key (32 bytes base64 encoded)
func testKey() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     testKey(),
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			key:     "not-valid-base64!@#",
			wantErr: true,
		},
		{
			name:    "wrong size key (16 bytes)",
			key:     base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEncryptor(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	enc, err := NewEncryptor(testKey())
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple string",
			plaintext: "hello world",
		},
		{
			name:      "API key format",
			plaintext: "sk-ant-api03-abc123xyz456",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "unicode",
			plaintext: "hello 世界 🌍",
		},
		{
			name:      "long string",
			plaintext: string(make([]byte, 1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Ciphertext should be different from plaintext
			if ciphertext == tt.plaintext && tt.plaintext != "" {
				t.Error("Ciphertext should be different from plaintext")
			}

			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	enc, _ := NewEncryptor(testKey())
	plaintext := "test message"

	ciphertext1, _ := enc.Encrypt(plaintext)
	ciphertext2, _ := enc.Encrypt(plaintext)

	// Same plaintext should produce different ciphertext (due to random nonce)
	if ciphertext1 == ciphertext2 {
		t.Error("Same plaintext should produce different ciphertext")
	}

	// Both should decrypt to the same plaintext
	decrypted1, _ := enc.Decrypt(ciphertext1)
	decrypted2, _ := enc.Decrypt(ciphertext2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to the same plaintext")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	enc, _ := NewEncryptor(testKey())

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!@#",
		},
		{
			name:       "too short",
			ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
		},
		{
			name:       "tampered ciphertext",
			ciphertext: base64.StdEncoding.EncodeToString(make([]byte, 100)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.ciphertext)
			if err == nil {
				t.Error("Decrypt() should fail for invalid ciphertext")
			}
		})
	}
}

func BenchmarkEncrypt(b *testing.B) {
	enc, _ := NewEncryptor(testKey())
	plaintext := "sk-ant-api03-abc123xyz456def789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encrypt(plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	enc, _ := NewEncryptor(testKey())
	ciphertext, _ := enc.Encrypt("sk-ant-api03-abc123xyz456def789")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decrypt(ciphertext)
	}
}
