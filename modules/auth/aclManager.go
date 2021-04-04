package auth

import (
	"encoding/json"
	"errors"
	"github.com/SWRMLabs/ss-store"
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

func NewAclManager(r repo.Repo) (ACL, error) {
	am := &aclManager{r.Store()}
	acls := map[string]string{}
	if ok := r.Config().Get("ACL", &acls); ok {
		for k, v := range acls {
			err := am.Configure(k, Role(v))
			if err != nil {
				return nil, err
			}
		}
	}
	return am, nil
}

func (a *aclManager) Configure(rsc string, role Role) error {
	r, ok := aclMap[role]
	if !ok {
		return errors.New("Invalid Role")
	}
	nacl := &Acl{
		Key:   rsc,
		Roles: r,
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
		// If there is no ACL configured, by default access is universal
		return true
	}
	r, ok := aclMap[role]
	if !ok {
		return false
	}
	return r >= nacl.Roles
}

func (a *aclManager) Allowed(rsc string) []Role {
	nacl := &Acl{
		Key: rsc,
	}
	err := a.st.Read(nacl)
	if err != nil {
		// If there is no ACL configured, by default access is universal
		return []Role{None, PublicRead, PublicWrite, AuthRead, AuthWrite, Admin}
	}
	roles := []Role{}
	for i := admin; i >= nacl.Roles; i-- {
		roles = append(roles, raclMap[i])
	}
	return roles
}
