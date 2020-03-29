package libp2p

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
)

type IdentityInfo struct {
	PrivKey crypto.PrivKey
	PubKey  crypto.PubKey
	ID      peer.ID
}

func InitIdentity(conf config.Config) (*IdentityInfo, error) {
	var (
		priv crypto.PrivKey
		pub  crypto.PubKey
	)
	privKeyStr, ok := conf.Get("priv_key").(string)
	if ok && len(privKeyStr) > 0 {
		log.Info("Reusing private key from config")
		pkBytes, err := base64.StdEncoding.DecodeString(privKeyStr)
		if err != nil {
			return nil, err
		}
		priv, err = crypto.UnmarshalPrivateKey(pkBytes)
		if err != nil {
			return nil, err
		}
		pub = priv.GetPublic()
		if pub == nil {
			return nil, errors.New("Public component nil")
		}
	} else {
		log.Info("Generating new identity for host")
		var e error
		priv, pub, e = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		if e != nil {
			return nil, e
		}
	}
	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return &IdentityInfo{
		PrivKey: priv,
		PubKey:  pub,
		ID:      pid,
	}, nil
}

func NewP2PHost(conf config.Config, id *IdentityInfo) (host.Host, error) {
	portVal, ok := conf.Get("p2p_port").(int32)
	if !ok {
		return nil, errors.New("P2P port absent")
	}

	return libp2p.New(
		context.Background(),
		libp2p.Identity(id.PrivKey),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", portVal),
		),
	)
}
