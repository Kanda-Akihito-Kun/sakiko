package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"
)

func TestEnsureMasterMaterialsGeneratesCertificates(t *testing.T) {
	service := New(Config{})

	materials, err := service.EnsureMasterMaterials("example.com")
	if err != nil {
		t.Fatalf("EnsureMasterMaterials() error = %v", err)
	}
	if materials.CACertPEM == "" || materials.ServerCertPEM == "" {
		t.Fatalf("expected CA and server certificates, got %+v", materials)
	}
	if materials.ServerName != "example.com" {
		t.Fatalf("expected server name example.com, got %q", materials.ServerName)
	}
}

func TestIssueKnightCertificateSignsCSR(t *testing.T) {
	service := New(Config{})
	if _, err := service.EnsureMasterMaterials("203.0.113.10"); err != nil {
		t.Fatalf("EnsureMasterMaterials() error = %v", err)
	}

	csrPEM := generateCSRPEM(t, "sakiko-knight")
	issued, err := service.IssueKnightCertificate(csrPEM, "knight-1", "Tokyo")
	if err != nil {
		t.Fatalf("IssueKnightCertificate() error = %v", err)
	}
	if issued.KnightID != "knight-1" {
		t.Fatalf("expected knight ID knight-1, got %q", issued.KnightID)
	}
	if issued.ClientCertificatePEM == "" || issued.CACertificatePEM == "" {
		t.Fatalf("expected issued certificate bundle, got %+v", issued)
	}
}

func generateCSRPEM(t *testing.T, commonName string) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, defaultRSAKeyBits)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: commonName},
	}, key)
	if err != nil {
		t.Fatalf("CreateCertificateRequest() error = %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	}))
}
