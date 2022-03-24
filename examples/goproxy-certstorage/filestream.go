package main

import (
	"errors"
	"log"
	"os"

	"github.com/elazarl/goproxy"
)

type FileStream struct {
	tipo    string
	path    string
	reqURI  string
	session int64
	f       *os.File
}

func NewFileStream(tipo string, path string, ctx *goproxy.ProxyCtx) *FileStream {

	reqURI := ctx.Req.RequestURI
	if reqURI == "" {
		reqURI = ctx.Req.URL.RequestURI()
	}

	return &FileStream{tipo, path, reqURI, ctx.Session, nil}
}

func (fs *FileStream) Write(b []byte) (nr int, err error) {
	if fs.f == nil {
		fs.f, err = os.Create(fs.path)
		if err != nil {
			return 0, err
		}
	}
	log.Println("Write:", string(b))
	return fs.f.Write(b)
}

func (fs *FileStream) Close() error {
	// fmt.Println("Close", fs.path)
	if fs.f == nil {
		return errors.New("FileStream was never written into")
	}
	return fs.f.Close()
}
