package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Meta struct {
	req      *http.Request
	resp     *http.Response
	err      error
	t        time.Time
	id       string
	sess     int64
	bodyPath string
	from     string
}

func fprintf(nr *int64, err *error, w io.Writer, pat string, a ...interface{}) {
	if *err != nil {
		return
	}
	var n int
	n, *err = fmt.Fprintf(w, pat, a...)
	*nr += int64(n)
}

func write(nr *int64, err *error, w io.Writer, b []byte) {
	if *err != nil {
		return
	}
	var n int
	n, *err = w.Write(b)
	*nr += int64(n)
}

func (m *Meta) WriteTo(w io.Writer) (nr int64, err error) {
	if m.req != nil {
		fprintf(&nr, &err, w, "==========================\r\nType: request\r\n")
	} else if m.resp != nil {
		fprintf(&nr, &err, w, "--------------------------\r\nType: response\r\n")
	}
	fprintf(&nr, &err, w, "ReceivedAt: %v\r\n", m.t)
	fprintf(&nr, &err, w, "Id: %s\r\n", m.id)
	fprintf(&nr, &err, w, "Session: %d\r\n", m.sess)
	fprintf(&nr, &err, w, "From: %v\r\n", m.from)
	if m.err != nil {
		// note the empty response
		fprintf(&nr, &err, w, "Error: %v\r\n\r\n\r\n\r\n", m.err)
	} else if m.req != nil {
		fprintf(&nr, &err, w, "\r\n")
		buf, err2 := httputil.DumpRequest(m.req, false)
		if err2 != nil {
			return nr, err2
		}
		write(&nr, &err, w, buf)

		linhas := strings.Split(string(buf), "\r\n")
		resumo := strings.Join(linhas[:2], "\r\n")
		fmt.Println(resumo, "\r\n...........................................")
	} else if m.resp != nil {
		fprintf(&nr, &err, w, "\r\n")
		buf, err2 := httputil.DumpResponse(m.resp, false)
		if err2 != nil {
			return nr, err2
		}
		write(&nr, &err, w, buf)
	}
	return
}
