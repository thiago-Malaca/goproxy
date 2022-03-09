package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
)

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8081", "proxy listen address")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.CertStore = NewCertStorage()

	proxy.Verbose = *verbose
	if proxy.Verbose {
		log.Printf("Servidor no ar! - Configurado para escutar o endereço: %s", *addr)
	}

	now := time.Now()
	sNow := now.Format("2006-01-02")

	// if err := os.MkdirAll(path, 0755); err != nil {
	// 	log.Fatal("Can't create dir", err)
	// }
	logger, err := NewLogger("db/" + sNow)
	if err != nil {
		log.Fatal("can't open log file", err)
	}

	// tr := &transport.Transport{Proxy: transport.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	// transport := &http.Transport{Proxy: transport.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// log.Println(req.URL.String())
		// fmt.Println("teste: ", ctx.Req.PostForm)
		// fmt.Println("LogRequest")
		// log.Println("LogRequest-req.Body: ", req.Body)

		tr := &transport.Transport{Proxy: transport.ProxyFromEnvironment, TLSClientConfig: &tls.Config{ServerName: req.Host}}

		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			ctx.UserData, resp, err = tr.DetailedRoundTrip(req)
			return
		})
		// id := uuid.New()
		// ctx.UserData.id = id.String()()

		// spew.Dump()

		// fmt.Println("req.ContentLength: ", req.ContentLength)

		// if req.ContentLength != 0 {
		logger.LogReq(req, ctx)
		// }

		return req, nil
	})

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		// body := path.Join(logger.path, fmt.Sprintf("%d_resp", ctx.Session))
		// contentType := resp.Header.Get("Content-Type")
		// log.Println(contentType)
		// logger.LogReq(ctx.Req, ctx)
		logger.LogResp(resp, ctx)
		return resp
	})

	log.Fatal(http.ListenAndServe(*addr, proxy))
}
