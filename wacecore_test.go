package main

import (
	"errors"
	"math/rand"
	"testing"
	"time"
	cf "wace/configstore"
	pm "wace/pluginmanager"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

var requestLine = "POST /cgi-bin/process.cgi HTTP/1.1\n"
var requestHeaders = `User-Agent: Mozilla/4.0 (compatible; MSIE5.01; Windows NT)
Host: www.tutorialspoint.com
Content-Type: application/x-www-form-urlencoded
Content-Length: length
Accept-Language: en-us
Accept-Encoding: gzip, deflate
Connection: Keep-Alive
`

var requestBody = "licenseID=string&content=string&/paramsXML=string\n"
var wholeRequest = requestLine + requestHeaders + "\n" + requestBody

var responseLine = "HTTP/1.1 200 OK\n"
var responseHeaders = `Date: Mon, 27 Jul 2009 12:28:53 GMT
Server: Apache/2.2.14 (Win32)
Last-Modified: Wed, 22 Jul 2009 19:15:56 GMT
Content-Length: 88
Content-Type: text/html
Connection: Closed
`
var responseBody = `<html>
<body>
<h1>Hello, World!</h1>
</body>
</html>
`
var wholeResponse = responseLine + responseHeaders + "\n" + responseBody

var config = []byte(`---
logpath: "/dev/null"
loglevel: DEBUG
listenport: "50051"
modelplugins:
  - id: "trivial"
    path: "_plugins/model/trivial.so"
    weight: 1
    threshold: 0.5
    params:
      d: "sds"
      b: "dnid"
      e: "dofnno"
    # plugintype: "RequestHeaders"
    plugintype: "Everything"
  - id: "trivial2"
    path: "_plugins/model/trivial2.so"
    weight: 2
    threshold: 0.1
    params:
      a: "sdsds"
      b: "sdfjdnid"
      c: "kfoskdofnno"
    plugintype: "Everything"
decisionplugins:
  - id: "simple"
    path: "_plugins/decision/simple.so"
    wafweight: 0.5
    decisionbalance: 0.5
`)

var configRoberta = []byte(`---
logpath: "/dev/null"
loglevel: DEBUG
listenport: "50051"
modelplugins:
  - id: "trivial"
    path: "_plugins/model/trivial.so"
    weight: 1
    threshold: 0.5
    params:
      d: "sds"
      b: "dnid"
      e: "dofnno"
    # plugintype: "RequestHeaders"
    plugintype: "Everything"
  - id: "trivial2"
    path: "_plugins/model/trivial2.so"
    weight: 2
    threshold: 0.1
    params:
      a: "sdsds"
      b: "sdfjdnid"
      c: "kfoskdofnno"
    plugintype: "Everything"
  - id: "roberta"
    path: "_plugins/model/roberta.so"
    weight: 1
    threshold: 0.5
    params:
      url: "localhost:9999"
      distance_threshold: -0.02
    plugintype: "AllRequest"
decisionplugins:
  - id: "simple"
    path: "_plugins/decision/simple.so"
    wafweight: 0.5
    decisionbalance: 0.5
`)

func generateRandomID() string {
	letters := "1234567890ABCDEF"
	id := ""
	for i := 0; i < 16; i++ {
		id += string(letters[rand.Intn(len(letters))])
	}

	return id
}

func init() {
	rand.Seed(time.Now().UnixNano())

	conf := cf.Get()
	err := conf.LoadConfigYaml(config)
	if err != nil {
		panic("Error loading config: " + err.Error())
	}

	logger := lg.Get()
	err = logger.LoadLogger(conf.LogPath, conf.LogLevel)
	if err != nil {
		panic("Error opening the wace log file: " + err.Error())
	}

	plugins = pm.New()
}

func TestAnalyzeRequestInParts(t *testing.T) {
	transactionID := generateRandomID()

	res := analyzeReqLineAndHeaders(transactionID, requestLine, requestHeaders, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeReqLineAndHeaders returned non-zero")
	}
	res = analyzeRequestBody(transactionID, requestBody, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeRequestBody returned non-zero")
	}

	_, err := checkTransaction(transactionID, "simple", make(map[string]string))
	if err != nil {
		t.Errorf("checkTransaction error: %v", err)
	}
}

func TestAnalyzeWholeRequest(t *testing.T) {
	transactionID := generateRandomID()

	res := analyzeRequest(transactionID, wholeRequest, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeRequest returned non-zero")
	}

	_, err := checkTransaction(transactionID, "simple",
		map[string]string{"anomalyscore": "200",
			"inboundthreshold": "100"})
	if err != nil {
		t.Errorf("checkTransaction error: %v", err)
	}
}

func TestAnalyzeResponseInParts(t *testing.T) {
	transactionID := generateRandomID()

	res := analyzeRespLineAndHeaders(transactionID, responseLine, responseHeaders, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeRespLineAndHeaders returned non-zero")
	}
	res = analyzeResponseBody(transactionID, responseBody, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeResponseBody returned non-zero")
	}

	_, err := checkTransaction(transactionID, "simple", make(map[string]string))
	if err != nil {
		t.Errorf("checkTransaction error: %v", err)
	}
}

func TestAnalyzeWholeResponse(t *testing.T) {
	transactionID := generateRandomID()

	res := analyzeResponse(transactionID, wholeResponse, []string{"trivial", "trivial2"})
	if res != 0 {
		t.Errorf("analyzeResponse returned non-zero")
	}

	_, err := checkTransaction(transactionID, "simple", make(map[string]string))
	if err != nil {
		t.Errorf("checkTransaction error: %v", err)
	}
}

func TestAnalyzeStress(t *testing.T) {
	for i := 0; i < 1000; i++ {
		transactionID := generateRandomID()
		analyzeRequest(transactionID, wholeRequest, []string{"trivial", "trivial2"})
		_, err := checkTransaction(transactionID, "simple", make(map[string]string))
		if err != nil {
			t.Errorf("checkTransaction error: %v", err)
		}
	}

}

func TestCheckInvalidTransaction(t *testing.T) {
	_, err := checkTransaction("INEXISTENT", "simple", make(map[string]string))
	if err == nil {
		t.Errorf("checkTransaction with inexistent transaction does not rise an error")
	}
}

func processRequest(models []string) error {
	transactionID := generateRandomID()

	res := analyzeRequest(transactionID, wholeRequest, models)
	if res != 0 {
		return errors.New("analyzeRequest returned non-zero")
	}

	_, err := checkTransaction(transactionID, "simple",
		map[string]string{"anomalyscore": "1",
			"inboundthreshold": "100"})
	return err
}

// func TestRoberta(t *testing.T) {
// 	conf := cf.Get()
// 	err := conf.LoadConfigYaml(configRoberta)
// 	if err != nil {
// 		panic("Error loading config: " + err.Error())
// 	}

// 	err = processRequest([]string{"roberta"})
// 	if err != nil {
// 		t.Errorf("callRoberta error: %v", err)
// 	}
// }

// func BenchmarkRoberta(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		processRequest([]string{"roberta"})
// 	}
// }

func BenchmarkTrivial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := processRequest([]string{"trivial"})
		if err != nil {
			b.Errorf("Error on Trivial benchmark: %v", err)
		}
	}
}
