package auth

import (
	"encoding/json"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/repo"
)

type Role string

type ACL interface {
	Configure(rsc string, role Role) error
	Delete(rsc string) error
	Authorized(rsc string, role Role) bool
	Allowed(rsc string) []Role
}

const (
	None        Role = "none"
	PublicRead  Role = "public_read"
	PublicWrite Role = "public_write"
	AuthRead    Role = "authenticated_read"
	AuthWrite   Role = "authenticated_write"
	Admin       Role = "admin"
)

const (
	noRole = iota
	pubRead
	pubWrite
	authRead
	authWrite
	admin
)

var aclMap = map[Role]int{
	None:        noRole,
	PublicRead:  pubRead,
	PublicWrite: pubWrite,
	AuthRead:    authRead,
	AuthWrite:   authWrite,
	Admin:       admin,
}

var raclMap = map[int]Role{
	noRole:    None,
	pubRead:   PublicRead,
	pubWrite:  PublicWrite,
	authRead:  AuthRead,
	authWrite: AuthWrite,
	admin:     Admin,
}

type Acl struct {
	Key   string
	Roles int
}

func (m *Acl) GetId() string {
	return m.Key
}

func (m *Acl) GetNamespace() string {
	return "Acl"
}

func (m *Acl) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Acl) Unmarshal(b []byte) error {
	return json.Unmarshal(b, m)
}

type aclManager struct {
	st store.Store
}

func NewAclManager(r repo.Repo) ACL {
	return &aclManager{r.Store()}
}

func (a *aclManager) Configure(rsc string, role Role) error {
	nacl := &Acl{
		Key:   rsc,
		Roles: aclMap[role],
	}
	return a.st.Update(nacl)
}

func (a *aclManager) Delete(rsc string) error {
	nacl := &Acl{
		Key: rsc,
	}
	return a.st.Delete(nacl)
}

func (a *aclManager) Authorized(rsc string, role Role) bool {
	nacl := &Acl{
		Key: rsc,
	}
	err := a.st.Read(nacl)
	if err != nil {
		// If there is no ACL configured, by default access is 'noAcl'
		return true
	}
	return aclMap[role] >= nacl.Roles
}

func (a *aclManager) Allowed(rsc string) []Role {
	nacl := &Acl{
		Key: rsc,
	}
	err := a.st.Read(nacl)
	if err != nil {
		// If there is no ACL configured, by default access is 'noAcl'
		return []Role{None}
	}
	roles := []Role{}
	for i := 1; i < nacl.Roles; i++ {
		roles = append(roles, raclMap[i])
	}
	return roles
}
