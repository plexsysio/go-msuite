package fsrepo

import (
	"github.com/plexsysio/go-msuite/modules/repo"
)

var Opener = opener

func GetRefCnt() int {
	return opener.refCnt
}

func GetActive() repo.Repo {
	return opener.active
}
