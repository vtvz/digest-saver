package token

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"os"
)

type Manager struct {
	filename string
}

func NewManager(filename string) *Manager {
	return &Manager{filename}
}

func (m *Manager) GetToken() (*oauth2.Token, error) {
	token, err := os.ReadFile(m.filename)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	if err := json.Unmarshal(token, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func (m *Manager) SaveToken(token *oauth2.Token) error {
	toSave, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(m.filename, toSave, 0644)
	if err != nil {
		return err
	}

	return nil
}
