package main

import(
	"testing"
	"bytes"
	"net/http"
	"reflect"
)

func  WriteHeaderTest (t *testing.T) {
	status := 505
	var rw http.ResponseWriter
	var b bytes.Buffer
	b.Write([]byte("Hello"))
	log := logWriter{
		ResponseWriter:	rw,
		statusCode:	400,
		response:	b,
	}
	log.WriteHeader(status)
	if log.statusCode != status {
		t.Errorf("statusCode was not set")
	}
}

func  WriteTest (t *testing.T) {
	var rw http.ResponseWriter
	var b bytes.Buffer
	b.Write([]byte("Hello"))
	log := logWriter{
		ResponseWriter:	rw,
		statusCode:	400,
		response:	b,
	}
	statusExpl := []byte("The server encountered an internal error or misconfiguration")
	log.Write(statusExpl)
	var statusExplByteBuff bytes.Buffer
	statusExplByteBuff.Write(statusExpl)
	if !reflect.DeepEqual(log.response, statusExplByteBuff) {
		t.Errorf("response was not set")
	}
}
