package pairing

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const defaultCodeTTL = 10 * time.Minute

type Service struct {
	now func() time.Time

	lock  sync.Mutex
	codes map[string]codeRecord
}

type codeRecord struct {
	code       interfaces.ClusterPairingCode
	createdAt  time.Time
	consumedAt time.Time
}

func New() *Service {
	return &Service{
		now:   time.Now,
		codes: map[string]codeRecord{},
	}
}

func (s *Service) CreateCode(req interfaces.ClusterCreatePairingCodeRequest) (interfaces.ClusterPairingCode, error) {
	if s == nil {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("pairing service is nil")
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = defaultCodeTTL
	}

	now := s.now().UTC()
	code := interfaces.ClusterPairingCode{
		Code:       randomCode(),
		KnightName: strings.TrimSpace(req.KnightName),
		ExpiresAt:  now.Add(ttl).Format(time.RFC3339),
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.pruneExpiredLocked(now)
	s.codes[code.Code] = codeRecord{
		code:      code,
		createdAt: now,
	}
	return code, nil
}

func (s *Service) ConsumeCode(rawCode string) (interfaces.ClusterPairingCode, error) {
	if s == nil {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("pairing service is nil")
	}

	code := strings.TrimSpace(rawCode)
	if code == "" {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("pairing code is required")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	now := s.now().UTC()
	s.pruneExpiredLocked(now)

	record, ok := s.codes[code]
	if !ok {
		return interfaces.ClusterPairingCode{}, fmt.Errorf("pairing code is invalid or expired")
	}
	delete(s.codes, code)
	record.consumedAt = now
	return record.code, nil
}

func (s *Service) pruneExpiredLocked(now time.Time) {
	for key, record := range s.codes {
		expiresAt, err := time.Parse(time.RFC3339, record.code.ExpiresAt)
		if err != nil || !expiresAt.After(now) {
			delete(s.codes, key)
		}
	}
}

func randomCode() string {
	var raw [8]byte
	_, _ = rand.Read(raw[:])
	return strings.ToUpper(hex.EncodeToString(raw[:]))
}
