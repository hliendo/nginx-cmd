package nginx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nginx-cmd/internal/config"
)

// Domain represents an nginx domain configuration and its SSL status
type Domain struct {
	Name      string
	Path      string
	SSLStatus string // "None", "Bootstrap", "Production", "Error"
	Type      string // "SPA", "Proxy", "Unknown"
	Target    string // Target URL/IP for proxy
	IsOrphan  bool   // True if cert exists but config doesn't, or vice-versa
}

// ListDomains returns a list of configured domains
func ListDomains() ([]Domain, error) {
	domainsPath := config.GetDomainsPath()
	files, err := os.ReadDir(domainsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Domain{}, nil
		}
		return nil, fmt.Errorf("failed to read domains directory: %v", err)
	}

	var domains []Domain
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".conf") {
			d := Domain{
				Name: strings.TrimSuffix(file.Name(), ".conf"),
				Path: filepath.Join(domainsPath, file.Name()),
				Type: "Unknown",
			}

			// Intentar extraer metadatos de los comentarios
			content, err := os.ReadFile(d.Path)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				// Solo miramos el encabezado (primeras 10 líneas)
				for i := 0; i < 10 && i < len(lines); i++ {
					line := strings.TrimSpace(lines[i])
					if !strings.HasPrefix(line, "#") {
						continue
					}
					if strings.Contains(line, "Tipo: SPA") || strings.Contains(line, "Type / Tipo: SPA") {
						d.Type = "SPA"
					} else if strings.Contains(line, "Tipo: Reverse Proxy") || strings.Contains(line, "Type / Tipo: Reverse Proxy") {
						d.Type = "Proxy"
					}
					if strings.Contains(line, "Target:") && !strings.Contains(line, "Target / Destino:") {
						parts := strings.Split(line, "Target:")
						if len(parts) > 1 {
							d.Target = strings.TrimSpace(parts[1])
						}
					}
					if strings.Contains(line, "Target / Destino:") {
						parts := strings.Split(line, "Target / Destino:")
						if len(parts) > 1 {
							d.Target = strings.TrimSpace(parts[1])
						}
					}
				}
			}

			domains = append(domains, d)
		}
	}
	return domains, nil
}
