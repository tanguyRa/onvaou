# crypto

```tree
crypto/
├── README.md
├── crypto.go
│   ├── type Encryptor {gcm: cipher.AEAD}
│   ├── func NewEncryptor(base64Key string) (*Encryptor, error)
│   ├── func (*Encryptor) Encrypt(plaintext string) (string, error)
│   └── func (*Encryptor) Decrypt(ciphertext string) (string, error)
└── crypto_test.go
    ├── func testKey() string
    ├── func TestNewEncryptor(t *testing.T)
    ├── func TestEncryptDecrypt(t *testing.T)
    ├── func TestEncryptProducesDifferentCiphertext(t *testing.T)
    ├── func TestDecryptInvalidCiphertext(t *testing.T)
    ├── func BenchmarkEncrypt(b *testing.B)
    └── func BenchmarkDecrypt(b *testing.B)
```
