package nginx

import (
	"fmt"
	"nginx-cmd/internal/config"
	"os/exec"
)

// Reload triggers nginx -s reload within the container
func Reload() error {
	cmd := exec.Command(config.GetNginxBinPath(), "-s", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload nginx: %v, output: %s", err, string(output))
	}
	return nil
}

// Test checks nginx configuration syntax
func Test() error {
	cmd := exec.Command(config.GetNginxBinPath(), "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed: %v, output: %s", err, string(output))
	}
	return nil
}
