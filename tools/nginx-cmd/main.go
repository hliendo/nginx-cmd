package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nginx-cmd/internal/config"
	"nginx-cmd/internal/nginx"
	"nginx-cmd/internal/pki"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nginx-cmd",
	Short: "Nginx Control Tool",
	Long: `A unified tool designed to manage Nginx domains, 
certificates, and internal trust stores in Docker-based architectures. 
It replaces traditional bash scripts with a robust, type-safe Go application.

Main features:
- Domain lifecycle management (add/remove/list)
- Internal PKI for SSL Bootstrap (Root/Intermediate CA)
- Container trust store integration
- In-container Nginx management (reload/test)

💡 MODO DE USO AVANZADO:
Tip: Ejecute subcomandos sin argumentos (ej. "nginx-cmd add") para ver sus flags 
opcionales habilitadas (como --grpc, --stream o --ws).`,
	Version: "1.0.0-kaizen",
}

var reloadCmd = &cobra.Command{
	Use:     "reload",
	Short:   "Reload Nginx configuration",
	Long:    `Performs a syntax check (nginx -t) and reloads the Nginx master process (nginx -s reload).`,
	Example: `  nginx-cmd reload`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := nginx.Reload(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Nginx reloaded successfully.")
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all configured domains",
	Long:    `Scans the Nginx domain configuration directory and lists all active .conf files with their SSL status.`,
	Example: `  nginx-cmd list`,
	Run: func(cmd *cobra.Command, args []string) {
		ca := pki.NewPKI()
		domains, err := nginx.ListDomains()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-30s | %-6s | %-12s | %-10s | %s\n", "DOMAIN", "TYPE", "SSL STATUS", "EXPIRY(d)", "TARGET")
		fmt.Println(strings.Repeat("-", 120))
		for _, d := range domains {
			sslStatus, expiry, _ := ca.CheckCertSource(d.Name)
			daysLeft := ""
			if !expiry.IsZero() {
				daysLeft = fmt.Sprintf("%v", int(time.Until(expiry).Hours()/24))
			}
			fmt.Printf("%-30s | %-6s | %-12s | %-10s | %s\n", d.Name, d.Type, sslStatus, daysLeft, d.Target)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Audit Nginx configurations and SSL certificates consistency",
	Long: `Systematically compares .conf files in conf.d/domains with certificates in /etc/letsencrypt.
Identifies:
- Configs without certificates.
- Orphan certificates (no config file).
- Bootstrap certificates needing Let's Encrypt activation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== NGINX-CMD INTEGRITY AUDIT ===")
		ca := pki.NewPKI()

		// 1. Check Configs -> Certs
		domains, _ := nginx.ListDomains()
		confMap := make(map[string]bool)
		for _, d := range domains {
			confMap[d.Name] = true
			status, expiry, _ := ca.CheckCertSource(d.Name)
			daysLeft := int(time.Until(expiry).Hours() / 24)

			if status == "None" {
				fmt.Printf("[FAIL] Domain %-20s has config but NO certificate.\n", d.Name)
			} else if status == "Bootstrap" {
				fmt.Printf("[WARN] Domain %-20s is using BOOTSTRAP cert (%d days left). Needs Let's Encrypt.\n", d.Name, daysLeft)
			} else {
				fmt.Printf("[ OK ] Domain %-20s is secure (Production: %d days left).\n", d.Name, daysLeft)
			}
		}

		// 2. Check Certs -> Configs (Orphans)
		files, _ := os.ReadDir(filepath.Join(config.GetCertsPath(), "live"))
		for _, f := range files {
			if f.IsDir() && f.Name() != "internal-pki" && !confMap[f.Name()] {
				fmt.Printf("[WARN] Orphan certificate folder detected: %s/live/%s (No corresponding .conf in domains/).\n", config.GetCertsPath(), f.Name())
			}
		}
		fmt.Println("=================================")
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [domain]",
	Short: "Remove a domain configuration and its certificates",
	Long: `Deletes the Nginx configuration file and flushes all associated certificates 
from /etc/letsencrypt (Live, Archive, and Renewal paths).`,
	Args:    cobra.ExactArgs(1),
	Example: `  nginx-cmd remove example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		fmt.Printf("Removing domain %s...\n", domain)

		configPath := filepath.Join(config.GetDomainsPath(), domain+".conf")
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing config: %v\n", err)
		}

		pathsToRemove := []string{
			filepath.Join(config.GetCertsPath(), "live", domain),
			filepath.Join(config.GetCertsPath(), "archive", domain),
			filepath.Join(config.GetCertsPath(), "renewal", domain+".conf"),
		}
		for _, path := range pathsToRemove {
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("Error removing %s: %v\n", path, err)
			}
		}

		if err := nginx.Reload(); err != nil {
			fmt.Printf("Error reloading nginx: %v\n", err)
		}

		fmt.Printf("Domain %s removed successfully.\n", domain)
	},
}

var addCmd = &cobra.Command{
	Use:   "add [domain] [type]",
	Short: "Add a new domain configuration",
	Long: `Creates a new Nginx server block and issues an internal SSL certificate.
Supported types:
- spa: Single Page Application (static files with /index.html routing)
- proxy: Reverse Proxy (forwards traffic to a target host/ip)
 
The command automatically ensures the Internal PKI is initialized and 
signs a leaf certificate for the domain.`,
	Args: cobra.MinimumNArgs(2),
	Example: `  nginx-cmd add example.com spa
  nginx-cmd add backend.com proxy --target 10.10.10.50:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		domainType := args[1]
		target, _ := cmd.Flags().GetString("target")
		ws, _ := cmd.Flags().GetBool("ws")
		grpcProxy, _ := cmd.Flags().GetBool("grpc")
		streamProxy, _ := cmd.Flags().GetBool("stream")

		fmt.Printf("Adding domain %s (%s)...\n", domain, domainType)

		ca := pki.NewPKI()
		if err := ca.EnsureCA(); err != nil {
			fmt.Printf("Error ensuring CA: %v\n", err)
			os.Exit(1)
		}

		certDir := filepath.Join(config.GetCertsPath(), "live", domain)
		if err := ca.IssueCert(domain, certDir); err != nil {
			fmt.Printf("Error issuing cert: %v\n", err)
			os.Exit(1)
		}

		if err := nginx.AddDomain(domain, domainType, target, ws, grpcProxy, streamProxy); err != nil {
			fmt.Printf("Error generating config: %v\n", err)
			os.Exit(1)
		}

		if err := nginx.Reload(); err != nil {
			fmt.Printf("Error reloading nginx: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Domain %s added successfully in bootstrap mode.\n", domain)

		// 🚀 AUTO-ESCALADO: Intentar obtener Certificado Real inmediatamente
		fmt.Printf("🔄 Attempting automatic production upgrade for %s...\n", domain)
		renewCmd.Run(cmd, []string{})
	},
}

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Install the internal Root CA into the container trust store",
	Long: `Copies the internal Root CA certificate to /usr/local/share/ca-certificates/ 
and runs 'update-ca-certificates'. 
Crucial for services running inside the container (e.g., Certbot, health checks) 
to trust internal certificates.`,
	Example: `  nginx-cmd trust`,
	Run: func(cmd *cobra.Command, args []string) {
		ca := pki.NewPKI()
		if err := ca.EnsureCA(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Installing internal Root CA into trust store...")
		rootCertPath := filepath.Join(ca.BaseDir, "root.crt")
		destPath := filepath.Join(config.GetTrustPath(), "internal-root.crt")

		input, err := os.ReadFile(rootCertPath)
		if err != nil {
			fmt.Printf("Error reading root cert: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(destPath, input, 0644); err != nil {
			fmt.Printf("Error copying cert: %v\n", err)
			os.Exit(1)
		}

		out, err := exec.Command("update-ca-certificates").CombinedOutput()
		if err != nil {
			fmt.Printf("Error updating trust store: %v\nOutput: %s\n", err, string(out))
			os.Exit(1)
		}
		fmt.Println("Trust store updated successfully.")
	},
}

var forceRenew bool
var dryRun bool

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew SSL certificates intelligently",
	Long: `Checks all domains and performs renewals.
If a domain is in 'Bootstrap' status, it attempts to obtain a Let's Encrypt certificate.
If a domain is in 'Production' status, it runs certbot renew.`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, err := nginx.ListDomains()
		if err != nil {
			fmt.Printf("Error listing domains: %v\n", err)
			os.Exit(1)
		}

		ca := pki.NewPKI()
		processedAny := false

		for i := range domains {
			d := &domains[i]
			// Update status
			status, expiry, _ := ca.CheckCertSource(d.Name)
			d.SSLStatus = status
			daysLeft := int(time.Until(expiry).Hours() / 24)

			fmt.Printf("Processing %s (Status: %s, Expiry: %d days)...\n", d.Name, d.SSLStatus, daysLeft)

			// BOOTSTRAP EMERGENCY: If no cert exists, issue an internal one so Nginx can at least START
			if d.SSLStatus == "None" {
				fmt.Printf("🚨 No certificate found for %s. Issuing internal Bootstrap certificate...\n", d.Name)
				certDir := filepath.Join(config.GetCertsPath(), "live", d.Name)
				if err := ca.IssueCert(d.Name, certDir); err != nil {
					fmt.Printf("❌ Failed to issue bootstrap cert: %v\n", err)
					continue
				}
				d.SSLStatus = "Bootstrap" // Update status for the next check
				processedAny = true
			}

			if d.SSLStatus == "Bootstrap" || (d.SSLStatus == "Production" && forceRenew) {
				// PRE-FLIGHT CHECK: Avoid Let's Encrypt ban if domain is not delegated
				fmt.Printf("🔍 Running Pre-flight DNS check for %s...\n", d.Name)
				ips, err := net.LookupIP(d.Name)
				if err != nil || len(ips) == 0 {
					fmt.Printf("⚠️ WARNING: DNS resolution failed for %s. Proceeding anyway (Certbot will decide).\n", d.Name)
				} else {
					fmt.Printf("📡 DNS OK: %s resolved to %v\n", d.Name, ips[0])
				}

				// 🛠️ LIMPIEZA INDUSTRIAL: Eliminar veneno de Bootstrap antes de Certbot
				livePath := filepath.Join(config.GetCertsPath(), "live", d.Name)
				if fi, err := os.Lstat(livePath); err == nil {
					if fi.Mode().IsDir() {
						fmt.Printf("🧹 Cleaning up Bootstrap directory in live/ to allow Certbot symlinks: %s\n", livePath)
						os.RemoveAll(livePath)
					}
				}

				// Reparar archivos de renovación corruptos (0 bytes)
				renewalFile := filepath.Join(config.GetCertsPath(), "renewal", d.Name+".conf")
				if fi, err := os.Stat(renewalFile); err == nil && fi.Size() == 0 {
					fmt.Printf("🩹 Removing corrupt renewal file (0 bytes): %s\n", renewalFile)
					os.Remove(renewalFile)
				}
				
				fmt.Printf("🚀 Attempting to obtain Let's Encrypt certificate for %s...\n", d.Name)
				webroot := filepath.Join(config.GetSitePath(), d.Name)

				// Ensure webroot exists
				os.MkdirAll(webroot, 0755)

				certbotArgs := []string{
					"certonly", "--webroot", "-w", webroot,
					"-d", d.Name,
					"--email", config.GetCertEmail(),
					"--agree-tos", "--no-eff-email", "--non-interactive",
				}
				if forceRenew {
					certbotArgs = append(certbotArgs, "--force-renewal")
				}
				if dryRun {
					fmt.Println("🔍 [DRY RUN] Simulating certbot request...")
					certbotArgs = append(certbotArgs, "--dry-run")
				}

				cbCmd := exec.Command("certbot", certbotArgs...)
				output, err := cbCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ Failed to obtain certificate for %s: %v\nOutput: %s\n", d.Name, err, string(output))
				} else {
					fmt.Printf("✅ Certificate obtained for %s!\n", d.Name)
					processedAny = true
				}
			}
		}

		// Run generic renewal for existing production certs
		fmt.Println("🔄 Running general certbot renew...")
		renewArgs := []string{"renew", "--non-interactive"}
		if forceRenew {
			renewArgs = append(renewArgs, "--force-renewal")
		}
		if dryRun {
			renewArgs = append(renewArgs, "--dry-run")
		}
		cbRenew := exec.Command("certbot", renewArgs...)
		output, _ := cbRenew.CombinedOutput()
		fmt.Println(string(output))

		if processedAny || strings.Contains(string(output), "Congratulations") {
			fmt.Println("Reloading Nginx...")
			if err := nginx.Reload(); err != nil {
				fmt.Printf("Error reloading nginx: %v\n", err)
			}
		}
	},
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [domain]",
	Short: "Inject an internal PKI certificate for a domain",
	Long: `Forces the creation of an internal Bootstrap certificate. 
Useful when Nginx won't start because certificates are missing.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		ca := pki.NewPKI()
		if err := ca.EnsureCA(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		certDir := filepath.Join(config.GetCertsPath(), "live", domain)
		if err := ca.IssueCert(domain, certDir); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Bootstrap certificate issued for %s.\n", domain)
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(trustCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(renewCmd)
	rootCmd.AddCommand(bootstrapCmd)

	addCmd.Flags().StringP("target", "p", "", "Target host:port for reverse proxy")
	addCmd.Flags().Bool("ws", false, "Enable WebSocket support (Reverse Proxy only)")
	addCmd.Flags().Bool("grpc", false, "Enable gRPC protocol support (Reverse Proxy only)")
	addCmd.Flags().Bool("stream", false, "Enable HTTP Streaming/SSE buffering-off (Reverse Proxy only)")
	renewCmd.Flags().BoolVarP(&forceRenew, "force", "f", false, "Force renewal of all certificates")
	renewCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate certificate renewal")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
