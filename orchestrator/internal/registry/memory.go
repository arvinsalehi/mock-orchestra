package registry

import (
	"sync"
	"time"

	"orchestrator/internal/core"
)

type Memory struct {
	mu   sync.RWMutex
	byID map[string]core.Device
}

func NewMemory() *Memory {
	return &Memory{byID: map[string]core.Device{}}
}

func (m *Memory) Upsert(d core.Device) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d.LastSeen = time.Now()
	m.byID[d.DeviceID] = d
}

func (m *Memory) All() []core.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]core.Device, 0, len(m.byID))
	for _, d := range m.byID {
		out = append(out, d)
	}
	return out
}

func (m *Memory) ByGroup(group string) []core.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]core.Device, 0)
	for _, d := range m.byID {
		if d.Group == group {
			out = append(out, d)
		}
	}
	return out
}
