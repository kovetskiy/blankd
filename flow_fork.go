package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

type HTTPHandler struct {
	program string
}

func forkFlow(args map[string]interface{}) {
	var (
		listenAddress = args["-l"].(string)
		program       = args["-e"].(string)
		master        = os.Getppid()
	)

	logger.Debugf("starting listening at %s", listenAddress)

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Fatalf("can't listen: %s", err)
	}

	logger.Debugf("sending signal to %d", master)

	err = syscall.Kill(master, ListeningStartedSignal)
	if err != nil {
		logger.Fatalf("can't send signal to %d: %s", master, err)
	}

	server := http.Server{
		Handler: &HTTPHandler{
			program: program,
		},
	}

	logger.Debugf("serving connections")

	err = server.Serve(listener)
	if err != nil {
		logger.Fatal(err)
	}
}

func (handler *HTTPHandler) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	var details map[string][]byte
	dir, err := ioutil.TempDir(os.TempDir(), "soul")
	if err != nil {
		log.Fatal(err)
	}

	logger.Printf("XXXXXX flow_fork.go:63: dir: %#v\n", dir)
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(time.Now().UnixNano())))
	details["_id"] = hex.EncodeToString(hasher.Sum(nil))

	details["raw"], err = httputil.DumpRequest(request, true)
	if err != nil {
		log.Fatal(err)
	}

	details["method"] = strings.ToUppe(request.Method)
	details["host"] = request.Host
	details["uri/raw"] = request.RequestURI
	details["uri/path"] = request.URL.RawPath
	details["uri/query"] = request.URL.RawQuery
	details["uri/fields"] = getFields(request.URL.Query(), "=")
	details["headers/raw"] = getValuesKeyValue(request.Header, ": ")
	details["headers/fields"] = getValuesKeyValue(request.Header, "=")

	details["cookies/raw"] = getValuesKeyValue(request.Cookies(), "=")
	details["cookies/fields"] = getValuesKeyValue(request.Cookies(), "=")
}

func getCookiesKeyValue(cookies []*http.Cookie, delimiter) string {
	for _, cookie := range cookies {
		cookie.String()
	}
	return ""
}

func getValuesKeyValue(values url.Values, delimiter) string {
	fields := []string{}
	for key, keyValues := range values {
		for _, value := range keyValues {
			fields = append(fields, key+delimiter+value)
		}
	}
	return strings.Join(fields, "\n")
}

