package sprydb

import (
	"errors"
)

type Manager struct {
	configs map[string]map[string]string
}

func NewManager() *Manager {
	return &Manager{
		configs: make(map[string]map[string]string),
	}
}

// 获取数据库连接
func (m *Manager) Connection(name string) (*Connection, error) {
	var (
		ok     bool
		config map[string]string
	)
	if name == "" {
		name = "default"
	}

	if config, ok = m.configs[name]; !ok {
		return nil, errors.New("No connection was found")
	}
	return NewConnection(config)
}

// 添加数据库连接
func (m *Manager) AddConnection(name string, config map[string]string) {
	if name == "" {
		name = "default"
	}
	m.configs[name] = config
}

// 添加多个数据库连接
func (m *Manager) AddMultiConnection(configs map[string]map[string]string) {
	m.configs = configs
}

// 根据文件添加配置
func (m *Manager) AddConnectionByFile(filePath string) {
	panic("implement me")
}

// 删除数据库连接
// 删除后会关闭该连接
func (m *Manager) DeleteConnection(name string) error {
	conn, err := m.Connection(name)
	if err != nil {
		return err
	}

	if err = conn.Close(); err != nil {
		return err
	}

	delete(m.configs, name)

	return nil
}
