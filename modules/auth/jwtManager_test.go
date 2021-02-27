package auth

import (
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"testing"
	"time"
)

func TestNewJWTManager(t *testing.T) {
	cfg := jsonConf.DefaultConfig()
	_, err := NewJWTManager(cfg)
	if err == nil {
		t.Fatal("Expected error while creating new JWT Manager")
	}
	cfg.Set("JWTSecret", "dummySecret")
	_, err = NewJWTManager(cfg)
	if err != nil {
		t.Fatal("Failed creating new JWT Manager", err.Error())
	}
}

type authUser struct {
	role string
	mtdt map[string]interface{}
}

func (a *authUser) ID() string {
	return "dummyID"
}

func (a *authUser) Role() string {
	return a.role
}

func (a *authUser) Mtdt() map[string]interface{} {
	return a.mtdt
}

func TestGenerateVerify(t *testing.T) {
	cfg := jsonConf.DefaultConfig()
	cfg.Set("JWTSecret", "dummySecret")
	jm, err := NewJWTManager(cfg)
	if err != nil {
		t.Fatal("Failed creating new JWT Manager", err.Error())
	}
	token, err := jm.Generate(&authUser{role: "admin"}, time.Second*5)
	if err != nil {
		t.Fatal("Failed generating new token", err.Error())
	}
	claims, err := jm.Verify(token)
	if err != nil {
		t.Fatal("Failed verifying token", err.Error())
	}
	if claims.ID != "dummyID" || claims.Role != "admin" {
		t.Fatal("Invalid claims in token", claims)
	}
	<-time.After(time.Second * 6)
	_, err = jm.Verify(token)
	if err == nil {
		t.Fatal("Expected error verifying expired token")
	}
}
