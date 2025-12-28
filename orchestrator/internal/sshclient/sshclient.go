package sshclient

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	User       string
	KeyPath    string
	Timeout    time.Duration

	// If StrictHost is false, host keys are not validated (lab-only).
	StrictHost bool
	KnownHosts string
}

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 8 * time.Second
	}
	return &Client{cfg: cfg}
}

func (c *Client) signer() (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(c.cfg.KeyPath)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(keyBytes)
}

func (c *Client) hostKeyCallback() (ssh.HostKeyCallback, error) {
	if !c.cfg.StrictHost {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	// Minimal strict mode: require a configured known_hosts file.
	// (Implement parsing known_hosts later if needed.)
	if c.cfg.KnownHosts == "" {
		return nil, errors.New("SSH_STRICT_HOST_KEY enabled but SSH_KNOWN_HOSTS not set")
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Placeholder strict validation hook; replace with known_hosts parsing.
		// Returning error here would block connections.
		return fmt.Errorf("strict host key validation not implemented; set SSH_STRICT_HOST_KEY=false for now")
	}, nil
}

func (c *Client) Dial(addr string) (*ssh.Client, error) {
	signer, err := c.signer()
	if err != nil {
		return nil, err
	}
	cb, err := c.hostKeyCallback()
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User:            c.cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: cb,
		Timeout:         c.cfg.Timeout,
	}

	return ssh.Dial("tcp", addr, cfg)
}

func RunCombined(client *ssh.Client, cmd string) (string, int, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", -1, err
	}
	defer sess.Close()

	out, err := sess.CombinedOutput(cmd)
	exitCode := 0
	if err != nil {
		var ee *ssh.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitStatus()
		} else {
			exitCode = -1
		}
	}
	return string(out), exitCode, err
}
