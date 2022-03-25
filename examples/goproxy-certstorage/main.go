package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
	"github.com/joho/godotenv"
)

func And(pcond ...goproxy.ReqCondition) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {

		retorno := true
		for _, cond := range pcond {
			retorno = retorno && cond.HandleReq(req, ctx)
		}

		return retorno
	}
}

func Or(pcond ...goproxy.ReqCondition) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {

		retorno := false
		for _, cond := range pcond {
			retorno = retorno || cond.HandleReq(req, ctx)
		}

		return retorno
	}
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8081", "proxy listen address")
	flag.Parse()

	setCA(caCert, caKey)
	proxy := goproxy.NewProxyHttpServer()
	// proxy.CertStore = NewCertStorage()

	proxy.Verbose = *verbose
	if proxy.Verbose {
		log.Printf("Servidor no ar! - Configurado para escutar o endereço: %s", *addr)
	}

	now := time.Now()
	sNow := now.Format("2006-01-02")

	logger, err := NewLogger("db/" + sNow)
	if err != nil {
		log.Fatal("can't open log file", err)
	}

	condicoes := []goproxy.ReqCondition{
		// Or(
		// 	goproxy.UrlMatches(regexp.MustCompile(`/bcpa-vendas/consultaPrevia/iniciarProposta`)),
		// 	goproxy.UrlMatches(regexp.MustCompile(`/bcpa-vendas/consultaPrevia/iniciarProposta2`)),
		// ),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.html`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.js`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.css`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.gif`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.png`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.svg`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`.woff2`))),
		goproxy.Not(goproxy.UrlMatches(regexp.MustCompile(`/rb_`))),
	}

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest(And(condicoes...)).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

		// tr := &transport.Transport{Proxy: transport.ProxyFromEnvironment}
		tr := &transport.Transport{Proxy: transport.ProxyFromEnvironment, TLSClientConfig: &tls.Config{ServerName: req.Host}}

		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			ctx.UserData, resp, err = tr.DetailedRoundTrip(req)
			return
		})

		logger.LogReq(req, ctx)

		return req, nil
	})

	proxy.OnResponse(And(condicoes...)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		logger.LogResp(resp, ctx)
		return resp
	})

	log.Fatal(http.ListenAndServe(*addr, proxy))
}

func getEnv(key string) string {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("É necessário informar a variável de ambiente: %s", key))
	}

	return value
}
