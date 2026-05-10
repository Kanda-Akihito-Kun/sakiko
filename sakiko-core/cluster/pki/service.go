package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultRSAKeyBits = 2048
)

type Config struct {
	CAValidity   time.Duration
	LeafValidity time.Duration
}

type Service struct {
	cfg Config

	lock      sync.RWMutex
	materials *MasterMaterials
}

type MasterMaterials struct {
	ServerName      string
	CACertPEM       string
	CAKeyPEM        string
	ServerCertPEM   string
	ServerKeyPEM    string
	CAExpiresAt     time.Time
	ServerExpiresAt time.Time

	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

type IssuedKnightCertificate struct {
	KnightID             string
	KnightName           string
	ClientCertificatePEM string
	CACertificatePEM     string
	MasterServerName     string
}

func New(cfg Config) *Service {
	if cfg.CAValidity <= 0 {
		cfg.CAValidity = 365 * 24 * time.Hour
	}
	if cfg.LeafValidity <= 0 {
		cfg.LeafValidity = 90 * 24 * time.Hour
	}
	return &Service{cfg: cfg}
}

func (s *Service) EnsureMasterMaterials(serverName string) (MasterMaterials, error) {
	if s == nil {
		return MasterMaterials{}, fmt.Errorf("pki service is nil")
	}

	serverName = strings.TrimSpace(serverName)

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.materials != nil && s.materials.ServerName == serverName {
		return cloneMaterials(*s.materials), nil
	}

	materials, err := generateMasterMaterials(serverName, s.cfg)
	if err != nil {
		return MasterMaterials{}, err
	}
	s.materials = &materials
	return cloneMaterials(materials), nil
}

func (s *Service) CurrentMasterMaterials() (MasterMaterials, bool) {
	if s == nil {
		return MasterMaterials{}, false
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.materials == nil {
		return MasterMaterials{}, false
	}
	return cloneMaterials(*s.materials), true
}

func (s *Service) LoadMasterMaterials(materials MasterMaterials) error {
	if s == nil {
		return fmt.Errorf("pki service is nil")
	}

	loaded, err := parseMasterMaterials(materials)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.materials = &loaded
	return nil
}

func (s *Service) IssueKnightCertificate(csrPEM string, knightID string, knightName string) (IssuedKnightCertificate, error) {
	if s == nil {
		return IssuedKnightCertificate{}, fmt.Errorf("pki service is nil")
	}

	s.lock.RLock()
	materials := s.materials
	s.lock.RUnlock()
	if materials == nil {
		return IssuedKnightCertificate{}, fmt.Errorf("master PKI materials are not ready")
	}

	csr, err := parseCSRPEM(csrPEM)
	if err != nil {
		return IssuedKnightCertificate{}, err
	}
	if err := csr.CheckSignature(); err != nil {
		return IssuedKnightCertificate{}, fmt.Errorf("invalid CSR signature: %w", err)
	}

	clientKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          randomSerialNumber(),
		Subject:               certificateSubject("sakiko-knight", knightID, knightName),
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(s.cfg.LeafValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           clientKeyUsage,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, materials.caCert, csr.PublicKey, materials.caKey)
	if err != nil {
		return IssuedKnightCertificate{}, fmt.Errorf("issue knight certificate: %w", err)
	}

	return IssuedKnightCertificate{
		KnightID:             strings.TrimSpace(knightID),
		KnightName:           strings.TrimSpace(knightName),
		ClientCertificatePEM: encodePEM("CERTIFICATE", der),
		CACertificatePEM:     materials.CACertPEM,
		MasterServerName:     materials.ServerName,
	}, nil
}

func generateMasterMaterials(serverName string, cfg Config) (MasterMaterials, error) {
	now := time.Now().UTC()

	caKey, err := rsa.GenerateKey(rand.Reader, defaultRSAKeyBits)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("generate CA key: %w", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          randomSerialNumber(),
		Subject:               pkix.Name{CommonName: "sakiko-remote-ca"},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(cfg.CAValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("create CA certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("parse CA certificate: %w", err)
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, defaultRSAKeyBits)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("generate server key: %w", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber:          randomSerialNumber(),
		Subject:               pkix.Name{CommonName: "sakiko-master"},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(cfg.LeafValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	if ip := net.ParseIP(serverName); ip != nil {
		serverTemplate.IPAddresses = []net.IP{ip}
	} else if strings.TrimSpace(serverName) != "" {
		serverTemplate.DNSNames = []string{strings.TrimSpace(serverName)}
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("create server certificate: %w", err)
	}

	return MasterMaterials{
		ServerName:      serverName,
		CACertPEM:       encodePEM("CERTIFICATE", caDER),
		CAKeyPEM:        encodePEM("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(caKey)),
		ServerCertPEM:   encodePEM("CERTIFICATE", serverDER),
		ServerKeyPEM:    encodePEM("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey)),
		CAExpiresAt:     caTemplate.NotAfter,
		ServerExpiresAt: serverTemplate.NotAfter,
		caCert:          caCert,
		caKey:           caKey,
	}, nil
}

func parseCSRPEM(raw string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(raw)))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("CSR PEM is required")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse CSR: %w", err)
	}
	return csr, nil
}

func encodePEM(blockType string, der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}))
}

func randomSerialNumber() *big.Int {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return serial
}

func certificateSubject(prefix string, knightID string, knightName string) pkix.Name {
	commonName := strings.TrimSpace(prefix)
	if value := strings.TrimSpace(knightID); value != "" {
		commonName = commonName + "-" + value
	}

	subject := pkix.Name{CommonName: commonName}
	if value := strings.TrimSpace(knightName); value != "" {
		subject.Organization = []string{value}
	}
	return subject
}

func cloneMaterials(materials MasterMaterials) MasterMaterials {
	return MasterMaterials{
		ServerName:      materials.ServerName,
		CACertPEM:       materials.CACertPEM,
		CAKeyPEM:        materials.CAKeyPEM,
		ServerCertPEM:   materials.ServerCertPEM,
		ServerKeyPEM:    materials.ServerKeyPEM,
		CAExpiresAt:     materials.CAExpiresAt,
		ServerExpiresAt: materials.ServerExpiresAt,
		caCert:          materials.caCert,
		caKey:           materials.caKey,
	}
}

func parseMasterMaterials(materials MasterMaterials) (MasterMaterials, error) {
	caBlock, _ := pem.Decode([]byte(strings.TrimSpace(materials.CACertPEM)))
	if caBlock == nil || caBlock.Type != "CERTIFICATE" {
		return MasterMaterials{}, fmt.Errorf("master CA certificate PEM is required")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("parse master CA certificate: %w", err)
	}

	caKeyBlock, _ := pem.Decode([]byte(strings.TrimSpace(materials.CAKeyPEM)))
	if caKeyBlock == nil || caKeyBlock.Type != "RSA PRIVATE KEY" {
		return MasterMaterials{}, fmt.Errorf("master CA private key PEM is required")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("parse master CA private key: %w", err)
	}

	serverBlock, _ := pem.Decode([]byte(strings.TrimSpace(materials.ServerCertPEM)))
	if serverBlock == nil || serverBlock.Type != "CERTIFICATE" {
		return MasterMaterials{}, fmt.Errorf("master server certificate PEM is required")
	}
	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	if err != nil {
		return MasterMaterials{}, fmt.Errorf("parse master server certificate: %w", err)
	}

	serverKeyBlock, _ := pem.Decode([]byte(strings.TrimSpace(materials.ServerKeyPEM)))
	if serverKeyBlock == nil || serverKeyBlock.Type != "RSA PRIVATE KEY" {
		return MasterMaterials{}, fmt.Errorf("master server private key PEM is required")
	}
	if _, err := x509.ParsePKCS1PrivateKey(serverKeyBlock.Bytes); err != nil {
		return MasterMaterials{}, fmt.Errorf("parse master server private key: %w", err)
	}

	loaded := cloneMaterials(materials)
	loaded.caCert = caCert
	loaded.caKey = caKey
	if loaded.CAExpiresAt.IsZero() {
		loaded.CAExpiresAt = caCert.NotAfter
	}
	if loaded.ServerExpiresAt.IsZero() {
		loaded.ServerExpiresAt = serverCert.NotAfter
	}
	return loaded, nil
}
