package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
	"github.com/google/uuid"
)

type HttpLogger struct {
	id    string
	path  string
	c     chan *Meta
	errch chan error
}

func NewLogger(rawbasepath string) (*HttpLogger, error) {
	id := uuid.New()
	basepath := path.Join(rawbasepath, id.String())
	if err := os.MkdirAll(basepath, 0755); err != nil {
		log.Fatal("Can't create dir", err)
	}
	f, err := os.Create(path.Join(basepath, "log"))
	if err != nil {
		return nil, err
	}

	logger := &HttpLogger{id.String(), basepath, make(chan *Meta), make(chan error)}
	go func() {
		for m := range logger.c {
			if _, err := m.WriteTo(f); err != nil {
				log.Println("Can't write meta", err)
			}
		}
		logger.errch <- f.Close()
	}()
	return logger, nil
}

func (logger *HttpLogger) LogResp(resp *http.Response, ctx *goproxy.ProxyCtx) {

	folder := path.Join(logger.path, fmt.Sprintf("%d", ctx.Session))
	if err := os.MkdirAll(folder, 0755); err != nil {
		log.Fatal("Can't create dir", err)
	}
	file := path.Join(folder, "resp")
	from := ""
	if ctx.UserData != nil {
		from = ctx.UserData.(*transport.RoundTripDetails).TCPAddr.String()
	}
	if resp == nil {
		resp = emptyResp
	} else {
		resp.Body = NewTeeReadCloser(resp.Body, NewFileStream("response", file, ctx))
	}
	logger.LogMeta(&Meta{
		resp: resp,
		err:  ctx.Error,
		t:    time.Now(),
		id:   logger.id,
		sess: ctx.Session,
		from: from})
}

var emptyResp = &http.Response{}
var emptyReq = &http.Request{}

func (logger *HttpLogger) LogReq(req *http.Request, ctx *goproxy.ProxyCtx) {

	folder := path.Join(logger.path, fmt.Sprintf("%d", ctx.Session))
	if err := os.MkdirAll(folder, 0755); err != nil {
		log.Fatal("Can't create dir", err)
	}
	file := path.Join(folder, "req")
	if req == nil {
		req = emptyReq
	} else if req.ContentLength != 0 {
		req.Body = NewTeeReadCloser(req.Body, NewFileStream("request", file, ctx))
	}
	logger.LogMeta(&Meta{
		req:  req,
		err:  ctx.Error,
		t:    time.Now(),
		id:   logger.id,
		sess: ctx.Session,
		from: req.RemoteAddr})
}

func (logger *HttpLogger) LogMeta(m *Meta) {
	logger.c <- m
}

func (logger *HttpLogger) Close() error {
	close(logger.c)
	return <-logger.errch
}
