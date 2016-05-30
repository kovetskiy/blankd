package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kovetskiy/executil"
)

var (
	reHeaderStatus = regexp.MustCompile(
		`^(HTTP/[\d.]+\s)?([\d]{3}).*`,
	)
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
	path, err := handler.dumpRequest(request)
	if err != nil {
		logger.Fatal(err)
	}

	cmd := exec.Command(handler.program, path)
	stdout, _, err := executil.Run(cmd)
	if err != nil {
		if !executil.IsExitError(err) {
			logger.Fatal(err)
			return
		}

		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	parts := strings.SplitN(string(stdout), "\n\n", 2)

	var body string
	if len(parts) > 0 {
		body = parts[1]
	}

	headers := strings.Split(parts[0], "\n")
	for i, header := range headers {
		if i == 0 {
			continue
		}

		keyValue := strings.SplitN(header, ":", 2)
		if len(keyValue) < 2 {
			continue
		}

		response.Header().Set(keyValue[0], keyValue[1])
	}

	if len(headers) > 0 {
		statusMatches := reHeaderStatus.FindStringSubmatch(headers[0])
		if len(statusMatches) != 0 {
			code, err := strconv.Atoi(statusMatches[1])
			if err != nil {
				log.Fatal(err)
			}

			response.WriteHeader(code)
		} else {
			logger.Fatalf("expected http status, but found: %s", statusMatches)
		}
	}

	response.Write([]byte(body))
}

func (handler *HTTPHandler) dumpRequest(
	request *http.Request,
) (string, error) {
	path, err := ioutil.TempDir(os.TempDir(), "soul")
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(time.Now().UnixNano())))
	requestID := hex.EncodeToString(hasher.Sum(nil))

	var body bytes.Buffer
	_, err = io.Copy(&body, request.Body)
	if err != nil {
		return "", err
	}

	err = request.ParseForm()
	if err != nil {
		return "", err
	}

	// I dunno why this missing
	request.Header.Add("Host", request.Host)

	dir := map[string]string{
		"_id":            requestID,
		"method":         strings.ToUpper(request.Method),
		"host":           request.Host,
		"uri/raw":        request.RequestURI,
		"uri/path":       request.URL.RawPath,
		"uri/query":      request.URL.RawQuery,
		"uri/fields":     getFields(request.URL.Query(), "="),
		"headers/raw":    getFields(request.Header, ": "),
		"headers/fields": getFields(request.Header, "="),
		"cookies/raw":    getCookiesRaw(request.Cookies()),
		"cookies/fields": getCookiesFields(request.Cookies(), "="),
		"body/raw":       body.String(),
		"body/fields":    getFields(request.Form, "="),
		"raw": getURIHeader(request) + getFields(request.Header, ": ") +
			"\n\n" + body.String(),
	}

	for filename, contents := range dir {
		fullpath := filepath.Join(path, filename)

		err = os.MkdirAll(filepath.Dir(fullpath), 0755)
		if err != nil {
			return "", fmt.Errorf(
				"can't mkdir %s: %s", filepath.Dir(fullpath), err,
			)
		}

		err = ioutil.WriteFile(fullpath, []byte(contents), 0644)
		if err != nil {
			return "", fmt.Errorf("can't write %s: %s", filename, err)
		}
	}

	return path, nil
}

func getURIHeader(request *http.Request) string {
	return fmt.Sprintf(
		"%s %s HTTP/%d.%d\n",
		request.Method,
		request.RequestURI,
		request.ProtoMajor,
		request.ProtoMinor,
	)
}

func getCookiesFields(cookies []*http.Cookie, delimiter string) string {
	var fields []string
	for _, cookie := range cookies {
		fields = append(fields, cookie.Name+delimiter+cookie.Value)
	}
	sort.Strings(fields)
	return strings.Join(fields, "\n")
}

func getCookiesRaw(cookies []*http.Cookie) string {
	var fields []string
	for _, cookie := range cookies {
		fields = append(fields, cookie.Raw)
	}
	return strings.Join(fields, "\n")
}

func getFields(values map[string][]string, delimiter string) string {
	fields := []string{}
	for key, keyValues := range values {
		for _, value := range keyValues {
			fields = append(fields, key+delimiter+value)
		}
	}
	sort.Strings(fields)
	return strings.Join(fields, "\n")
}
