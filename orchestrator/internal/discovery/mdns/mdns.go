package mdns

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"

	"orchestrator/internal/core"
)

type Config struct {
	Service     string // e.g. "_testbench._tcp"
	Domain      string // "local."
	DefaultPort int

	Logger *log.Logger
}

type Discovery struct {
	cfg     Config
	updates chan core.Device

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(cfg Config) *Discovery {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Discovery{
		cfg:     cfg,
		updates: make(chan core.Device, 64),
	}
}

func (d *Discovery) Updates() <-chan core.Device { return d.updates }

// Start runs a browse loop and emits device upserts to Updates().
func (d *Discovery) Start(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	d.cancel = cancel

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)

	// Consumer
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-entries:
				if e == nil {
					continue
				}
				dev, ok := entryToDevice(e, d.cfg.DefaultPort)
				if !ok {
					continue
				}
				select {
				case d.updates <- dev:
				default:
					// If updates channel is full, drop to avoid blocking discovery
					d.cfg.Logger.Printf("[mDNS] drop update device=%s addr=%s", dev.DeviceID, dev.Addr)
				}
			}
		}
	}()

	// Browse loop (rebrowse periodically)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for ctx.Err() == nil {
			bctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			_ = resolver.Browse(bctx, d.cfg.Service, d.cfg.Domain, entries)
			<-bctx.Done()
			cancel()
			time.Sleep(2 * time.Second)
		}
	}()

	d.cfg.Logger.Printf("[mDNS] browsing service=%s domain=%s", d.cfg.Service, d.cfg.Domain)
	return nil
}

func (d *Discovery) Stop() error {
	if d.cancel != nil {
		d.cancel()
	}
	d.wg.Wait()
	close(d.updates)
	return nil
}

func parseTXT(txt []string) map[string]string {
	m := map[string]string{}
	for _, kv := range txt {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return m
}

func pickIP(e *zeroconf.ServiceEntry) net.IP {
	if len(e.AddrIPv4) > 0 {
		return e.AddrIPv4[0]
	}
	if len(e.AddrIPv6) > 0 {
		return e.AddrIPv6[0]
	}
	return nil
}

func entryToDevice(e *zeroconf.ServiceEntry, defaultPort int) (core.Device, bool) {
	ip := pickIP(e)
	if ip == nil {
		return core.Device{}, false
	}

	txt := parseTXT(e.Text)
	deviceID := txt["device_id"]
	if deviceID == "" {
		// Fallback: instance name, but TXT is strongly preferred
		deviceID = e.Instance
	}
	port := e.Port
	if port == 0 {
		port = defaultPort
	}

	return core.Device{
		DeviceID: deviceID,
		Product:  txt["product"],
		Serial:   txt["serial"],
		Group:    txt["group"],
		Addr:     net.JoinHostPort(ip.String(), itoa(port)),
		LastSeen: time.Now(),
		Meta:     txt,
	}, true
}

func itoa(n int) string {
	// tiny local helper to avoid importing strconv everywhere
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(b[i:])
}
