package main

import (
	"errors"
	"os"

	"github.com/elazarl/goproxy"
)

type CtxProposta struct {
	tipo    string
	reqURI  string
	session int64
}

type FileStream struct {
	path        string
	ctxProposta *CtxProposta
	f           *os.File
}

func NewFileStream(tipo string, path string, ctx *goproxy.ProxyCtx) *FileStream {

	reqURI := ctx.Req.RequestURI
	if reqURI == "" {
		reqURI = ctx.Req.URL.RequestURI()
	}
	ctxProposta := &CtxProposta{tipo, reqURI, ctx.Session}

	return &FileStream{path, ctxProposta, nil}
}

func (fs *FileStream) Write(b []byte) (nr int, err error) {
	if fs.f == nil {
		fs.f, err = os.Create(fs.path)
		if err != nil {
			return 0, err
		}
	}
	SendData(fs.ctxProposta, b)

	return fs.f.Write(b)
}

func (fs *FileStream) Close() error {
	// fmt.Println("Close", fs.path)
	if fs.f == nil {
		return errors.New("FileStream was never written into")
	}
	return fs.f.Close()
}
