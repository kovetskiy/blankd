package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
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
	root    string
}

func forkFlow(args map[string]interface{}) {
	var (
		listenAddress = args["-l"].(string)
		program       = args["-e"].(string)
		rootDirectory = args["-d"].(string)
		useTLS        = args["--tls"].(bool)
		master        = os.Getppid()
	)

	var listener net.Listener
	var err error
	if useTLS {
		_, _, err := executil.Run(
			exec.Command(
				"openssl", "req", "-x509", "-newkey", "rsa:1024",
				"-keyout", filepath.Join(rootDirectory, "tls.key"),
				"-out", filepath.Join(rootDirectory, "tls.crt"),
				"-days", "9999", "-nodes",
				"-subj", "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost",
			),
		)
		if err != nil {
			logger.Fatalf("can't generate tls certificate: %s", err)
		}

		certificate, err := tls.LoadX509KeyPair(
			filepath.Join(rootDirectory, "tls.crt"),
			filepath.Join(rootDirectory, "tls.key"),
		)
		if err != nil {
			logger.Fatalf("can't load certificate: %s", err)
		}

		logger.Debugf("starting listening at %s", listenAddress)

		listener, err = tls.Listen("tcp", listenAddress, &tls.Config{
			Certificates: []tls.Certificate{certificate},
		})
		if err != nil {
			logger.Fatalf("can't listen: %s", err)
		}
	} else {
		logger.Debugf("starting listening at %s", listenAddress)
		listener, err = net.Listen("tcp", listenAddress)
		if err != nil {
			logger.Fatalf("can't listen: %s", err)
		}
	}

	logger.Debugf("sending signal to %d", master)

	err = syscall.Kill(master, ListeningStartedSignal)
	if err != nil {
		logger.Fatalf("can't send signal to %d: %s", master, err)
	}

	server := http.Server{
		Handler: &HTTPHandler{
			program: program,
			root:    rootDirectory,
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
	defer func() {
		if err := recover(); err != nil {
			logger.Fatalf("PANIC: %s", err)
		}
	}()

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

		logger.Error(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	parts := strings.SplitN(string(stdout), "\n\n", 2)

	var body string
	if len(parts) == 2 {
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
		if headers[0] != "" {
			statusMatches := reHeaderStatus.FindStringSubmatch(headers[0])
			if len(statusMatches) != 0 {
				code, err := strconv.Atoi(statusMatches[2])
				if err != nil {
					logger.Fatal(err)
				}

				response.WriteHeader(code)
			} else {
				logger.Fatalf("expected http status, but found: '%s'", headers[0])
			}
		}
	}

	response.Write([]byte(body))
}

func (handler *HTTPHandler) dumpRequest(
	request *http.Request,
) (string, error) {
	dump, err := getRequestDump(request)
	if err != nil {
		return "", err
	}

	path, err := ioutil.TempDir(handler.root, dump["_id"])
	if err != nil {
		return "", err
	}

	for filename, contents := range dump {
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

func getRequestDump(request *http.Request) (map[string]string, error) {
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprint(time.Now().UnixNano())))
	requestID := hex.EncodeToString(hasher.Sum(nil))

	body := newBuffer(nil)

	if request.Body != nil {
		_, err := io.Copy(body, request.Body)
		if err != nil {
			return nil, err
		}

		request.Body = newBuffer(body.Bytes())
	}

	err := request.ParseForm()
	if err != nil {
		return nil, err
	}

	var headers bytes.Buffer
	err = request.Header.WriteSubset(&headers, nil)
	if err != nil {
		return nil, err
	}

	dump := map[string]string{
		"_id":            requestID,
		"method":         strings.ToUpper(request.Method),
		"host":           request.Host,
		"uri/raw":        request.RequestURI,
		"uri/path":       request.URL.Path,
		"uri/query":      request.URL.RawQuery,
		"uri/values":     getValues(request.URL.Query(), "="),
		"headers/raw":    string(headers.Bytes()),
		"headers/values": getValues(request.Header, "="),
		"cookies":        getCookies(request.Cookies()),
		"body/raw":       string(body.Bytes()),
		"body/values":    getValues(request.Form, "="),
		"raw": getURIHeader(request) + getValues(request.Header, ": ") +
			"\n\n" + body.String(),
	}

	return dump, nil
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

func getCookies(cookies []*http.Cookie) string {
	var values []string
	for _, cookie := range cookies {
		values = append(values, cookie.String())
	}

	sort.Strings(values)

	return strings.Join(values, "\n")
}

func getValues(raw map[string][]string, delimiter string) string {
	values := []string{}
	for key, keyValues := range raw {
		for _, value := range keyValues {
			values = append(values, key+delimiter+value)
		}
	}
	sort.Strings(values)
	return strings.Join(values, "\n")
}
