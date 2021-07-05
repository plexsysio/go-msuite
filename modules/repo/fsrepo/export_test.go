package fsrepo

var Opener = opener

func GetRefCnt() int {
	return opener.refCnt
}

func GetActive() *fsRepo {
	return opener.active
}
