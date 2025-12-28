package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MQTTBrokerURL  string
	MQTTClientID   string
	MQTTTopicStart string

	MDNSService    string
	MDNSDomain     string
	DeviceSSHPort  int

	SSHUser            string
	SSHKeyPath         string
	SSHStrictHostKey   bool
	SSHKnownHostsPath  string

	LocalTestsRoot string
	RemoteDir      string
}

func env(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func envBool(k string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes"
}

func envInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func Load() Config {
	return Config{
		MQTTBrokerURL:  env("MQTT_BROKER_URL", "tcp://mosquitto:1883"),
		MQTTClientID:   env("MQTT_CLIENT_ID", "go-orchestrator-mdns-ssh"),
		MQTTTopicStart: env("MQTT_TOPIC_START", "hub/session/start"),

		MDNSService:   env("MDNS_SERVICE", "_testbench._tcp"),
		MDNSDomain:    env("MDNS_DOMAIN", "local."),
		DeviceSSHPort: envInt("DEVICE_SSH_PORT", 22),

		SSHUser:           env("SSH_USER", "pi"),
		SSHKeyPath:        env("SSH_KEY_PATH", "/run/secrets/ssh_key"),
		SSHStrictHostKey:  envBool("SSH_STRICT_HOST_KEY", false),
		SSHKnownHostsPath: env("SSH_KNOWN_HOSTS", "/etc/ssh/ssh_known_hosts"),

		LocalTestsRoot: env("LOCAL_TESTS_ROOT", "./tests"),
		RemoteDir:      env("REMOTE_DIR", "/tmp/test_run"),
	}
}
