package acl

import (
	"encoding/json"
)

type MethodRoles struct {
	Method string
	Roles  []string
}

func (m *MethodRoles) GetId() string {
	return m.Method
}

func (m *MethodRoles) GetNamespace() string {
	return "MethodRoles"
}

func (m *MethodRoles) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MethodRoles) Unmarshal(b []byte) error {
	return json.Unmarshal(b, m)
}
