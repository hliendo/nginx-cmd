package config

import (
	"os"
	"path/filepath"
)

// Default paths for Docker environment
const (
	DefaultNginxPath = "/opt/nginx"
	DefaultCertsPath = "/etc/letsencrypt"
	DefaultSitePath  = "/opt/site"
	DefaultTrustPath = "/usr/local/share/ca-certificates"
)

// GetNginxPath returns the base directory for Nginx
func GetNginxPath() string {
	return getEnv("NGINX_CMD_NGINX_PATH", DefaultNginxPath)
}

// GetCertsPath returns the base directory for SSL certificates
func GetCertsPath() string {
	return getEnv("NGINX_CMD_CERTS_PATH", DefaultCertsPath)
}

// GetSitePath returns the base directory for static sites
func GetSitePath() string {
	return getEnv("NGINX_CMD_SITE_PATH", DefaultSitePath)
}

// GetTrustPath returns the system trust store path
func GetTrustPath() string {
	return getEnv("NGINX_CMD_TRUST_PATH", DefaultTrustPath)
}

// GetDomainsPath returns the directory where domain configs are stored
func GetDomainsPath() string {
	return filepath.Join(GetNginxPath(), "conf/conf.d/domains")
}

// GetPKIDir returns the directory for the internal CA
func GetPKIDir() string {
	return filepath.Join(GetCertsPath(), "internal-pki")
}

// GetNginxBinPath returns the path to the nginx binary
func GetNginxBinPath() string {
	return filepath.Join(GetNginxPath(), "sbin/nginx")
}

// GetPKIRootCN returns the Common Name for the internal Root CA
func GetPKIRootCN() string {
	return getEnv("NGINX_CMD_PKI_ROOT_CN", "Internal Root CA")
}

// GetPKIIntermediateCN returns the Common Name for the internal Intermediate CA
func GetPKIIntermediateCN() string {
	return getEnv("NGINX_CMD_PKI_INTER_CN", "Internal Intermediate CA")
}

// GetCertEmail returns the email for Certbot registration
func GetCertEmail() string {
	return getEnv("NGINX_CMD_CERT_EMAIL", "admin@example.com")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
