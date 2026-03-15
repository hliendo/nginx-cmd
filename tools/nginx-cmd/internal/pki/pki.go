package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"nginx-cmd/internal/config"
)

// PKI handles the internal certificate authority
type PKI struct {
	BaseDir string
}

func NewPKI() *PKI {
	return &PKI{BaseDir: config.GetPKIDir()}
}

// EnsureCA checks if Root and Intermediate CAs exist, generating them if not
func (p *PKI) EnsureCA() error {
	if err := os.MkdirAll(p.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create PKI directory: %v", err)
	}

	rootPath := filepath.Join(p.BaseDir, "root.crt")
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		if err := p.generateRoot(); err != nil {
			return err
		}
	}

	interPath := filepath.Join(p.BaseDir, "intermediate.crt")
	if _, err := os.Stat(interPath); os.IsNotExist(err) {
		if err := p.generateIntermediate(); err != nil {
			return err
		}
	}

	return nil
}

func (p *PKI) generateRoot() error {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: config.GetPKIRootCN(),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	return p.saveCertAndKey("root", derBytes, priv)
}

func (p *PKI) generateIntermediate() error {
	rootCert, rootKey, err := p.loadCertAndKey("root")
	if err != nil {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: config.GetPKIIntermediateCN(),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0), // 5 years
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCert, &priv.PublicKey, rootKey)
	if err != nil {
		return err
	}

	return p.saveCertAndKey("intermediate", derBytes, priv)
}

// IssueCert signs a leaf certificate for a domain using the Intermediate CA
func (p *PKI) IssueCert(domain, outputDir string) error {
	interCert, interKey, err := p.loadCertAndKey("intermediate")
	if err != nil {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:    []string{domain},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 3, 0), // 3 months emulating Certbot
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, interCert, &priv.PublicKey, interKey)
	if err != nil {
		return err
	}

	// Save to the destination directory (simulating certbot path)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := p.savePEM(filepath.Join(outputDir, "fullchain.pem"), "CERTIFICATE", derBytes); err != nil {
		return err
	}

	// Prepend intermediate to fullchain (matches Certbot behavior)
	interPEM, _ := p.loadPEM("intermediate", "CERTIFICATE")
	fullChainFile, _ := os.OpenFile(filepath.Join(outputDir, "fullchain.pem"), os.O_APPEND|os.O_WRONLY, 0644)
	fullChainFile.Write(interPEM)
	fullChainFile.Close()

	return p.savePEM(filepath.Join(outputDir, "privkey.pem"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(priv))
}

func (p *PKI) saveCertAndKey(name string, der []byte, key *rsa.PrivateKey) error {
	if err := p.savePEM(filepath.Join(p.BaseDir, name+".crt"), "CERTIFICATE", der); err != nil {
		return err
	}
	return p.savePEM(filepath.Join(p.BaseDir, name+".key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func (p *PKI) savePEM(path, blockType string, b []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: b})
}

func (p *PKI) loadCertAndKey(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(filepath.Join(p.BaseDir, name+".crt"))
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	keyPEM, err := os.ReadFile(filepath.Join(p.BaseDir, name+".key"))
	if err != nil {
		return nil, nil, err
	}
	block, _ = pem.Decode(keyPEM)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func (p *PKI) loadPEM(name, blockType string) ([]byte, error) {
	return os.ReadFile(filepath.Join(p.BaseDir, name+".crt"))
}

// CheckCertSource returns the source of the certificate (None, Bootstrap, Production) and its expiration date.
func (p *PKI) CheckCertSource(domain string) (string, time.Time, error) {
	path := filepath.Join(config.GetCertsPath(), "live", domain, "fullchain.pem")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "None", time.Time{}, nil
	}

	certPEM, err := os.ReadFile(path)
	if err != nil {
		return "Error", time.Time{}, err
	}

	// Decode the first certificate in the chain
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "Invalid", time.Time{}, fmt.Errorf("failed to decode PEM for domain %s", domain)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "Error", time.Time{}, err
	}

	// Internal Root CA has CN: "Internal Root CA"
	// Internal Intermediate CA has CN: "Internal Intermediate CA"
	issuerCN := cert.Issuer.CommonName
	source := "Production"
	if issuerCN == config.GetPKIRootCN() || issuerCN == config.GetPKIIntermediateCN() {
		source = "Bootstrap"
	}

	return source, cert.NotAfter, nil
}
