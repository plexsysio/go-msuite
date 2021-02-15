package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/StreamSpace/ss-store"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"go.uber.org/fx"
	"net/http"
	"strconv"
	"time"
)

var Module = fx.Options(
	fx.Invoke(NewCDNService),
)

var DefaultTimeout time.Duration = time.Second * 10

func NewCDNService(
	p *ipfslite.Peer,
	st store.Store,
	mux *http.ServeMux,
) {
	svc := &cdn{p: p, st: st}
	mux.HandleFunc("/v1/cdn/put", svc.Put)
	mux.HandleFunc("/v1/cdn/get", svc.Get)
	mux.HandleFunc("/v1/cdn/list", svc.List)
}

type cdn struct {
	p  *ipfslite.Peer
	st store.Store
}

type FileObj struct {
	Name     string
	Size     int64
	Cid      string
	Uploader string
	Acl      string
	Created  int64
}

func (f *FileObj) GetId() string {
	return f.Cid
}

func (f *FileObj) GetNamespace() string {
	return "FileObj"
}

func (f *FileObj) Marshal() ([]byte, error) {
	return json.Marshal(f)
}

func (f *FileObj) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, f)
}

func (f *FileObj) Factory() store.SerializedItem {
	return f
}

func errorHTML(msg string, w http.ResponseWriter) {
	fmt.Fprintf(w, fmt.Sprintf("<html><body style='font-size:100px'>%s</body></html>", msg))
	return
}

func (c *cdn) Put(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorHTML("BadRequest: Failed parsing form", w)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorHTML("BadRequest: Failed parsing file", w)
		return
	}
	defer file.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	nd, err := c.p.AddFile(ctx, file, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: "+err.Error(), w)
		return
	}
	nf := &FileObj{
		Name:    handler.Filename,
		Size:    handler.Size,
		Cid:     nd.Cid().String(),
		Created: time.Now().Unix(),
	}
	err = c.st.Create(nf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: Failed creating metadata Err:"+err.Error(), w)
		return
	}
	resp, err := json.Marshal(nf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: Failed serializing response", w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (c *cdn) Get(w http.ResponseWriter, r *http.Request) {
	fileId := r.URL.Path
	cid, err := cid.Decode(fileId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorHTML("BadRequest: Failed parsing file ID Err:"+err.Error(), w)
		return
	}
	f := &FileObj{
		Cid: fileId,
	}
	err = c.st.Read(f)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorHTML("NotFound: Failed getting file metadata Err:"+err.Error(), w)
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	rdr, err := c.p.GetFile(ctx, cid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: Failed reading file Err:"+err.Error(), w)
		return
	}
	http.ServeContent(w, r, f.Name, time.Unix(f.Created, 0), rdr)
}

func (c *cdn) List(w http.ResponseWriter, r *http.Request) {
	pg, lim := 0, 0
	p := r.URL.Query().Get("page")
	var err error
	if len(p) == 0 {
		pg = 0
	} else {
		pg, err = strconv.Atoi(p)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errorHTML("BadRequest: Failed parsing page Err:"+err.Error(), w)
			return
		}
	}
	l := r.URL.Query().Get("limit")
	if len(l) == 0 {
		lim = 10
	} else {
		lim, err = strconv.Atoi(l)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errorHTML("BadRequest: Failed parsing limit Err:"+err.Error(), w)
			return
		}
	}
	items, err := c.st.List(&FileObj{}, store.ListOpt{
		Page:  int64(pg),
		Limit: int64(lim),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: Failed listing files Err:"+err.Error(), w)
		return
	}
	retList := []*FileObj{}
	for _, v := range items {
		retList = append(retList, v.(*FileObj))
	}
	resp, err := json.Marshal(retList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHTML("InternalError: Failed serializing response", w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
