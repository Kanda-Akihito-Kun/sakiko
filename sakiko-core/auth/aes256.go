package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/protocol"
)

type Config struct {
	Secret       string
	AllowedDrift time.Duration
	NonceTTL     time.Duration
}

type Verifier struct {
	key          []byte
	allowedDrift time.Duration
	nonceTTL     time.Duration
	nonceStore   map[string]time.Time
	lock         sync.Mutex
}

func NewVerifier(cfg Config) (*Verifier, error) {
	secret := strings.TrimSpace(cfg.Secret)
	if secret == "" {
		return nil, fmt.Errorf("auth secret is required")
	}
	if cfg.AllowedDrift <= 0 {
		cfg.AllowedDrift = 30 * time.Second
	}
	if cfg.NonceTTL <= 0 {
		cfg.NonceTTL = 5 * time.Minute
	}
	sum := sha256.Sum256([]byte(secret))
	return &Verifier{
		key:          sum[:],
		allowedDrift: cfg.AllowedDrift,
		nonceTTL:     cfg.NonceTTL,
		nonceStore:   map[string]time.Time{},
	}, nil
}

func (v *Verifier) NewChallenge() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (v *Verifier) Sign(env protocol.Envelope) (string, error) {
	body, err := canonicalBody(env)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(body)
	iv := deriveIV(env.Nonce, env.Timestamp)
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad(digest[:], aes.BlockSize)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return base64.StdEncoding.EncodeToString(out), nil
}

func (v *Verifier) Verify(env protocol.Envelope, now time.Time) error {
	if strings.TrimSpace(env.Signature) == "" {
		return fmt.Errorf("missing signature")
	}
	if env.Timestamp == 0 {
		return fmt.Errorf("missing timestamp")
	}
	if strings.TrimSpace(env.Nonce) == "" {
		return fmt.Errorf("missing nonce")
	}

	diff := now.Sub(time.UnixMilli(env.Timestamp))
	if diff < 0 {
		diff = -diff
	}
	if diff > v.allowedDrift {
		return fmt.Errorf("timestamp out of range")
	}

	if err := v.claimNonce(env.Nonce, now); err != nil {
		return err
	}

	signed, err := v.Sign(env)
	if err != nil {
		return err
	}
	if signed != env.Signature {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

type signingEnvelope struct {
	Version   string          `json:"version,omitempty"`
	RequestID string          `json:"requestId,omitempty"`
	Event     string          `json:"event"`
	Timestamp int64           `json:"ts"`
	Nonce     string          `json:"nonce"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

func canonicalBody(env protocol.Envelope) ([]byte, error) {
	return json.Marshal(signingEnvelope{
		Version:   env.Version,
		RequestID: env.RequestID,
		Event:     env.Event,
		Timestamp: env.Timestamp,
		Nonce:     env.Nonce,
		Payload:   env.Payload,
	})
}

func deriveIV(nonce string, ts int64) []byte {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", nonce, ts)))
	return sum[:aes.BlockSize]
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	padSize := blockSize - (len(src) % blockSize)
	return append(src, bytes.Repeat([]byte{byte(padSize)}, padSize)...)
}

func (v *Verifier) claimNonce(nonce string, now time.Time) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for key, expiresAt := range v.nonceStore {
		if now.After(expiresAt) {
			delete(v.nonceStore, key)
		}
	}
	if expiresAt, ok := v.nonceStore[nonce]; ok && now.Before(expiresAt) {
		return fmt.Errorf("nonce replay detected")
	}
	v.nonceStore[nonce] = now.Add(v.nonceTTL)
	return nil
}
