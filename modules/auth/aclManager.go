package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/plexsysio/gkvstore"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
)

type Role string

type ACL interface {
	Configure(ctx context.Context, rsc string, role Role) error
	Delete(ctx context.Context, rsc string) error
	Authorized(ctx context.Context, rsc string, role Role) bool
	Allowed(ctx context.Context, rsc string) []Role
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

func (m *Acl) GetID() string {
	return m.Key
}

func (*Acl) GetNamespace() string {
	return "acl"
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

func NewAclManager(r repo.Repo, shStore sharedStorage.Provider) (ACL, error) {
	fmt.Println("CALLED")
	var (
		st  gkvstore.Store
		err error
	)
	// If P2P mode is configured, share ACLs across nodes
	if shStore != nil {
		st, err = shStore.SharedStorage("acl", nil)
	} else {
		fmt.Println("LOCAL MODE")
		st = r.Store()
	}
	if err != nil {
		return nil, err
	}
	am := &aclManager{st}
	acls := map[string]string{}
	if ok := r.Config().Get("ACL", &acls); ok {
		for k, v := range acls {
			err := am.Configure(context.Background(), k, Role(v))
			if err != nil {
				return nil, err
			}
		}
	}
	return am, nil
}

func (a *aclManager) Configure(ctx context.Context, rsc string, role Role) error {
	r, ok := aclMap[role]
	if !ok {
		return errors.New("Invalid Role")
	}
	nacl := &Acl{
		Key:   rsc,
		Roles: r,
	}
	return a.st.Update(ctx, nacl)
}

func (a *aclManager) Delete(ctx context.Context, rsc string) error {
	nacl := &Acl{
		Key: rsc,
	}
	return a.st.Delete(ctx, nacl)
}

func (a *aclManager) Authorized(ctx context.Context, rsc string, role Role) bool {
	nacl := &Acl{
		Key: rsc,
	}
	err := a.st.Read(ctx, nacl)
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

func (a *aclManager) Allowed(ctx context.Context, rsc string) []Role {
	nacl := &Acl{
		Key: rsc,
	}
	err := a.st.Read(ctx, nacl)
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
