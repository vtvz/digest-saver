package config

import (
	"encoding/json"
	"os"
)

type Manager struct {
	filename string
}

func NewManager(filename string) *Manager {
	return &Manager{filename: filename}
}

func (m *Manager) Get() (*Config, error) {
	content, err := os.ReadFile(m.filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (m *Manager) Init() error {
	config := &Config{
		ServerAddr:             ":8080",
		RedirectUrl:            "http://localhost:8080/callback",
		ClientId:               "-- insert client id --",
		ClientSecret:           "-- insert client secret --",
		ReleaseTargetPlaylist:  "-- insert playlist id where save release radar tracks --",
		DiscoverTargetPlaylist: "-- insert playlist id where save weekly discover tracks --",
	}

	toSave, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(m.filename, toSave, 0644)
	if err != nil {
		return err
	}

	return nil
}
