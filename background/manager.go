package background

import (
	"context"
	"time"

	"github.com/oklog/oklog/pkg/group"
)

type Manager struct {
	ctx          context.Context
	serviceGroup group.Group
	services     []IService
}

func NewManager(ctx context.Context, job ...IService) *Manager {
	return &Manager{
		ctx:      ctx,
		services: job,
	}
}

func (m *Manager) AddJob(name string, tickRate, timeout time.Duration, processor Processor, opts ...OptionFn) {
	m.services = append(m.services, NewJob(name, tickRate, timeout, processor, opts...))
}

func (m *Manager) HasJobs() bool {
	return len(m.services) > 0
}

func (m *Manager) Run() error {
	for i := range m.services {
		service := m.services[i]
		m.serviceGroup.Add(func() error { return service.Run(m.ctx) }, func(error) {})
	}

	return m.serviceGroup.Run()
}

func (m *Manager) Stop() {
	for _, service := range m.services {
		service.Stop()
	}
}
