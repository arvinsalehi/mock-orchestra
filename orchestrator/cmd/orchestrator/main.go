package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"orchestrator/internal/config"
	"orchestrator/internal/core"
	"orchestrator/internal/discovery/mdns"
	"orchestrator/internal/registry"
	"orchestrator/internal/runner"
	"orchestrator/internal/sshclient"
)

type SessionStartPayload struct {
	SessionID   string `json:"session_id"`
	TestPlan    string `json:"test_plan"`    // e.g. "suite_full"
	TargetGroup string `json:"target_group"` // e.g. "lab_bench_A"
	BuildNumber string `json:"build_number"` // for UI traceability
}

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1) Registry (in-memory; later swap for Redis/DB-backed)
	reg := registry.NewMemory()

	// 2) Discovery (mDNS/DNS-SD)
	disc := mdns.New(mdns.Config{
		Service:     cfg.MDNSService,
		Domain:      cfg.MDNSDomain,
		DefaultPort: cfg.DeviceSSHPort,
		Logger:      log.Default(),
	})
	if err := disc.Start(ctx); err != nil {
		log.Fatalf("mDNS discovery start failed: %v", err)
	}
	go func() {
		for d := range disc.Updates() {
			reg.Upsert(d)
		}
	}()

	// 3) Runner (SSH identify + deploy + execute)
	sshc := sshclient.New(sshclient.Config{
		User:       cfg.SSHUser,
		KeyPath:    cfg.SSHKeyPath,
		Timeout:    8 * time.Second,
		StrictHost: cfg.SSHStrictHostKey,
		KnownHosts: cfg.SSHKnownHostsPath,
	})
	run := runner.New(runner.Config{
		RemoteDir:      cfg.RemoteDir,
		LocalTestsRoot: cfg.LocalTestsRoot,
		Logger:         log.Default(),
	}, sshc)

	// 4) MQTT start-session subscriber
	mc := mqtt.NewClient(mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL).
		SetClientID(cfg.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second),
	)

	if tok := mc.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatalf("MQTT connect failed: %v", tok.Error())
	}
	log.Printf("[MQTT] connected broker=%s", cfg.MQTTBrokerURL)

	if tok := mc.Subscribe(cfg.MQTTTopicStart, 1, func(_ mqtt.Client, m mqtt.Message) {
		var p SessionStartPayload
		if err := json.Unmarshal(m.Payload(), &p); err != nil {
			log.Printf("[Start] invalid payload: %v", err)
			return
		}

		// Select devices: by group if provided, else all discovered
		var devices []core.Device
		if p.TargetGroup != "" {
			devices = reg.ByGroup(p.TargetGroup)
		} else {
			devices = reg.All()
		}

		if len(devices) == 0 {
			log.Printf("[Start] no devices available (group=%q)", p.TargetGroup)
			return
		}

		log.Printf("[Start] session=%s build=%s plan=%s group=%s devices=%d",
			p.SessionID, p.BuildNumber, p.TestPlan, p.TargetGroup, len(devices),
		)

		// Dispatch runs concurrently (one goroutine per device).
		for _, d := range devices {
			d := d
			go func() {
				err := run.DeployAndRun(ctx, d, runner.RunCommand{
					SessionID:   p.SessionID,
					BuildNumber: p.BuildNumber,
					Plan:        p.TestPlan,
				})
				if err != nil {
					log.Printf("[Run] device=%s addr=%s err=%v", d.DeviceID, d.Addr, err)
				}
			}()
		}
	}); tok.Wait() && tok.Error() != nil {
		log.Fatalf("MQTT subscribe failed topic=%s err=%v", cfg.MQTTTopicStart, tok.Error())
	}
	log.Printf("[MQTT] subscribed topic=%s", cfg.MQTTTopicStart)

	// 5) Block until shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("[Shutdown] stopping...")
	cancel()
	mc.Disconnect(250)
	_ = disc.Stop()
}
