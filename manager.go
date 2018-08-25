package sprydb

import (
	"errors"
)

type Manager struct {
	configs  map[string]map[string]string
}

func NewManager() *Manager {
	return &Manager{
		configs:make(map[string]map[string]string),
	}
}

func (m *Manager) Connection(name string) (*Connection, error) {
	var (
		ok bool
		config map[string]string
	)
	if name == "" {
		name = "default"
	}
	if config, ok = m.configs[name]; !ok{
		return nil, errors.New("No connection was found")
	}
	return NewConnection(config)
}

func (m *Manager) AddConnection(name string, config map[string]string) {
	if name == "" {
		name = "default"
	}
	m.configs[name] = config
}

func (m *Manager) AddMultiConnection(configs map[string]map[string]string){
	m.configs = configs
}

func (m *Manager) DeleteConnection(name string) {
	delete(m.configs, name)
}