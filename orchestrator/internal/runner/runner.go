package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"

	"orchestrator/internal/core"
	"orchestrator/internal/deploy"
	"orchestrator/internal/sshclient"
)

type Config struct {
	RemoteDir      string
	LocalTestsRoot string
	Logger         *log.Logger
}

type RunCommand struct {
	SessionID   string
	BuildNumber string
	Plan        string
}

type Runner struct {
	cfg Config
	ssh *sshclient.Client
}

func New(cfg Config, sshc *sshclient.Client) *Runner {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Runner{cfg: cfg, ssh: sshc}
}

// DeployAndRun (Plan B): send a high-level plan to the device by uploading a bundle + executing a standard start script.
// The device-side scripts should publish live results to MQTT.
func (r *Runner) DeployAndRun(ctx context.Context, d core.Device, cmd RunCommand) error {
	client, err := r.ssh.Dial(d.Addr)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer client.Close()

	product := d.Product
	serial := d.Serial

	// Identify via SSH if missing metadata
	if product == "" || serial == "" {
		p, s, _ := r.identify(r.ssh, d.Addr)
		if product == "" {
			product = p
		}
		if serial == "" {
			serial = s
		}
	}
	r.cfg.Logger.Printf("[Identify] device=%s product=%s serial=%s addr=%s", d.DeviceID, product, serial, d.Addr)

	localFolder := filepath.Join(r.cfg.LocalTestsRoot, product, cmd.Plan)
	if _, err := os.Stat(localFolder); err != nil {
		return fmt.Errorf("local tests not found: %s", localFolder)
	}

	// SFTP upload bundle.tar.gz
	sftpc, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("sftp new client: %w", err)
	}
	defer sftpc.Close()

	if err := deploy.UploadDirAsTarGz(sftpc, localFolder, r.cfg.RemoteDir); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	// Extract + run standardized start.sh (bundle must include start.sh)
	run := fmt.Sprintf(
		"bash -lc 'set -euo pipefail; "+
			"mkdir -p %[1]s; "+
			"tar -xzf %[1]s/bundle.tar.gz -C %[1]s; "+
			"chmod +x %[1]s/start.sh; "+
			"%[1]s/start.sh --session %q --device %q --build %q --plan %q'"+
			"rm -rf %[1]s/*; ",
		r.cfg.RemoteDir, cmd.SessionID, d.DeviceID, cmd.BuildNumber, cmd.Plan,
	)

	out, code, err := sshclient.RunCombined(client, run)
	r.cfg.Logger.Printf("[Exec] device=%s exit=%d err=%v output=%s", d.DeviceID, code, err, trim(out, 2000))
	if err != nil {
		return fmt.Errorf("remote exec (exit=%d): %w", code, err)
	}
	return nil
}

func (r *Runner) identify(client *sshclient.Client, addr string) (product, serial string, err error) {
	c, err := client.Dial(addr)
	if err != nil {
		return "", "", err
	}
	defer c.Close()

	out, _, _ := sshclient.RunCombined(c, "cat /etc/product_id 2>/dev/null || true")
	product = strings.TrimSpace(out)

	out, _, _ = sshclient.RunCombined(c, "cat /etc/device_serial 2>/dev/null || true")
	serial = strings.TrimSpace(out)
	return product, serial, nil
}

// identify uses the established ssh connection to read product/serial files.
func (r *Runner) identifySSH(client *sshclient.Client, addr string) (string, string, error) {
	c, err := client.Dial(addr)
	if err != nil {
		return "", "", err
	}
	defer c.Close()
	return r.identifyConn(c)
}

func (r *Runner) identifyConn(client interface{}) (string, string, error) { return "", "", nil }

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(trimmed)"
}
