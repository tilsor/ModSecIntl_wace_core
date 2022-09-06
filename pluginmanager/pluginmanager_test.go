package pluginmanager

import (
	"bytes"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
	cf "wace/configstore"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

var baseConfig = `---
logpath: "/dev/null"
loglevel: "WARN"
listenport: "50051"
`

var trivialPlugin = `  - id: "trivial"
    path: "../_plugins/model/trivial.so"
    weight: 1
    threshold: 0.5
    params:
      param1: "first value"
      param2: "second value"
      param3: "third value"
    plugintype: "Everything"
`

var testPlugin = `  - id: "test"
    path: "../_plugins/decision/test.so"
    wafweight: 0.5
    decisionbalance: 0.5
    params:
      test1: "test"
      test2: "testtest"
      test3: "testtesttest"
`

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

	logger := lg.Get()
	err := logger.LoadLogger("/dev/null", lg.ERROR)
	if err != nil {
		panic("Error loading logger")
	}
}

func TestPluginInit(t *testing.T) {
	cases := []struct{ id, conf string }{
		// 		{"invalid_path", `  - id: "invalid_path"
		//     path: "../_plugins/model/nonexistent.so"
		//     plugintype: "AllRequest"
		// `},
		{"no_init", `  - id: "no_init"
    path: "../_plugins/model/no_init.so"
    plugintype: "AllRequest"
`},
		{"wrong_init", `  - id: "wrong_init"
    path: "../_plugins/model/wrong_init.so"
    plugintype: "AllRequest"
`},
		{"error_init", `  - id: "error_init"
    path: "../_plugins/model/error_init.so"
    plugintype: "AllRequest"
`},
	}

	// Test model plugin initialization
	for _, c := range cases {
		config := baseConfig + "modelplugins:\n" + trivialPlugin + c.conf

		conf := cf.Get()
		err := conf.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Error loading config: %v", err)
		}
		plugins := New()
		if _, exists := plugins.modelPlugins["trivial"]; !exists {
			t.Errorf("trivial plugin not loaded")
		}
		if _, exists := plugins.modelPlugins[c.id]; exists {
			t.Errorf(c.id + " should not load")
		}
	}

	// Test decision plugin initialization
	for _, c := range cases {
		config := baseConfig + "modelplugins:\n" + trivialPlugin + "decisionplugins:\n" + testPlugin + c.conf

		conf := cf.Get()
		err := conf.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Error loading config: %v", err)
		}
		plugins := New()
		if _, exists := plugins.decisionPlugins["test"]; !exists {
			t.Errorf("test plugin not loaded")
		}
		if _, exists := plugins.decisionPlugins[c.id]; exists {
			t.Errorf(c.id + " should not load")
		}
	}

}

func TestPluginParams(t *testing.T) {
	config := baseConfig + "modelplugins:\n" + trivialPlugin + "decisionplugins:\n" + testPlugin

	conf := cf.Get()
	err := conf.LoadConfigYaml([]byte(config))
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	var buf bytes.Buffer
	logger := lg.Get()
	err = logger.LoadLoggerWriter(&buf, lg.INFO)
	if err != nil {
		t.Errorf("Error loading logger: %v", err)
	}

	plugins := New()

	if !strings.Contains(buf.String(), "[trivial:InitPlugin] map[param1:first value param2:second value param3:third value]") {
		t.Errorf("trivial plugin did not initialize correctly")
	}
	if !strings.Contains(buf.String(), "[test:InitPlugin] map[test1:test test2:testtest test3:testtesttest]") {
		t.Errorf("test plugin did not initialize correctly")
	}

	transactionID := generateRandomID()
	modelPlugStatus := make(chan ModelStatus)
	go plugins.ProcessRequest("trivial", "test request1", cf.AllRequest, transactionID, modelPlugStatus)
	<-modelPlugStatus
	if !strings.Contains(buf.String(), "[trivial:ProcessRequest] \"test request1\"") {
		t.Errorf("trivial plugin did not analyze request")
	}

	go plugins.ProcessResponse("trivial", "test response1", cf.AllResponse, transactionID, modelPlugStatus)
	<-modelPlugStatus
	if !strings.Contains(buf.String(), "[trivial:ProcessResponse] \"test response1\"") {
		t.Errorf("trivial plugin did not analyze response")
	}

	_, err = plugins.CheckResult(transactionID, "test", map[string]string{"anomalyscore": "100", "inboundthreshold": "10"})
	if err != nil {
		t.Errorf("Error checking result: %v", err)
	}
	if !strings.Contains(buf.String(), "[test:CheckResults]") {
		t.Errorf("test plugin did not execute correctly")
	}
	if !strings.Contains(buf.String(), "modelRes: map[trivial:") {
		t.Errorf("trivial result is not stored in modelRes")
	}
	if !strings.Contains(buf.String(), "modelWeight: map[trivial:1]") {
		t.Errorf("trivial weight is not stored in modelWeight")
	}
	if !strings.Contains(buf.String(), "modelThres: map[trivial:0.5]") {
		t.Errorf("trivial threshold is not stored in modelWeight")
	}
	if !strings.Contains(buf.String(), "wafData: map[anomalyscore:100 inboundthreshold:10]") {
		t.Errorf("waf params are not stored in wafData")
	}
}

func TestPluginType(t *testing.T) {
	cases := []struct {
		id                      string
		pluginType, requestType cf.ModelPluginType
		executes                bool
	}{
		{"req_headers-req_headers", cf.RequestHeaders, cf.RequestHeaders, true},
		{"req_headers-resp_headers", cf.RequestHeaders, cf.ResponseHeaders, false},
		{"req_headers-all_req", cf.RequestHeaders, cf.AllRequest, false},
		{"all_req-req_headers", cf.AllRequest, cf.RequestHeaders, false},
		{"all_req-all_resp", cf.AllRequest, cf.AllResponse, false},

		{"resp_headers-resp_headers", cf.ResponseHeaders, cf.ResponseHeaders, true},
		{"resp_headers-req_headers", cf.ResponseHeaders, cf.RequestHeaders, false},
		{"resp_headers-all_resp", cf.ResponseHeaders, cf.AllResponse, false},
		{"all_resp-resp_headers", cf.AllResponse, cf.ResponseHeaders, false},
		{"all_resp-all_req", cf.AllResponse, cf.AllRequest, false},

		{"everything-req_headers", cf.Everything, cf.RequestHeaders, true},
		{"everything-all_req", cf.Everything, cf.AllRequest, true},
		{"everything-resp_body", cf.Everything, cf.ResponseBody, true},
		{"everything-all_resp", cf.Everything, cf.AllResponse, true},
	}

	for _, c := range cases {
		config := baseConfig + "modelplugins:\n" +
			"  - id: \"" + c.id + "\"\n" +
			"    path: \"../_plugins/model/trivial.so\"\n" +
			"    plugintype: \"" + c.pluginType.String() + "\"\n"

		conf := cf.Get()
		err := conf.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Error loading config: %v", err)
		}

		old := log.Writer()
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(old)

		plugins := New()

		transactionID := generateRandomID()
		modelPlugStatus := make(chan ModelStatus)
		switch c.requestType {
		case cf.RequestHeaders, cf.RequestBody, cf.AllRequest:
			go plugins.ProcessRequest(c.id, "test request", c.requestType, transactionID, modelPlugStatus)
			<-modelPlugStatus
			if strings.Contains(buf.String(), "[trivial:ProcessRequest] \"test request\"") != c.executes {
				t.Errorf("case %s: expected to run trivial plugin: %v", c.id, c.executes)
			}
			if _, exists := plugins.results[transactionID]; exists != c.executes {
				t.Errorf("case %s: expected to store results: %v", c.id, c.executes)
			}
		case cf.ResponseHeaders, cf.ResponseBody, cf.AllResponse:
			go plugins.ProcessResponse(c.id, "test response", c.requestType, transactionID, modelPlugStatus)
			<-modelPlugStatus
			if strings.Contains(buf.String(), "[trivial:ProcessResponse] \"test response\"") != c.executes {
				t.Errorf("case %s: expected to run trivial plugin: %v", c.id, c.executes)
			}
			if _, exists := plugins.results[transactionID]; exists != c.executes {
				t.Errorf("case %s: expected to store results: %v", c.id, c.executes)
			}
		}

	}
}

func TestProcessRequestInvalid(t *testing.T) {
	cases := []struct{ id, conf string }{
		{"no_req", `  - id: "no_req"
    path: "../_plugins/model/no_req.so"
    plugintype: "Everything"
`},
		{"wrong_req", `  - id: "wrong_req"
    path: "../_plugins/model/wrong_req.so"
    plugintype: "Everything"
`},
		{"error_req", `  - id: "error_req"
    path: "../_plugins/model/error_req.so"
    plugintype: "Everything"
`},
	}

	// Test model plugin initialization
	for _, c := range cases {
		config := baseConfig + "modelplugins:\n" + trivialPlugin + c.conf

		conf := cf.Get()
		err := conf.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Error loading config: %v", err)
		}
		plugins := New()

		transactionID := generateRandomID()
		modelPlugStatus := make(chan ModelStatus)
		go plugins.ProcessRequest(c.id, "test request", cf.AllRequest, transactionID, modelPlugStatus)
		<-modelPlugStatus
		go plugins.ProcessResponse(c.id, "test response", cf.AllResponse, transactionID, modelPlugStatus)
		<-modelPlugStatus

		if _, exists := plugins.results[transactionID]; exists {
			t.Errorf("invalid test %s stored a result", c.id)
		}
	}

	config := baseConfig + "modelplugins:\n" + trivialPlugin

	conf := cf.Get()
	err := conf.LoadConfigYaml([]byte(config))
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	plugins := New()

	transactionID := generateRandomID()
	modelPlugStatus := make(chan ModelStatus)
	go plugins.ProcessRequest("nonexistent", "test request", cf.AllRequest, transactionID, modelPlugStatus)
	<-modelPlugStatus
	go plugins.ProcessResponse("nonexistent", "test response", cf.AllResponse, transactionID, modelPlugStatus)
	<-modelPlugStatus

	if _, exists := plugins.results[transactionID]; exists {
		t.Errorf("nonexistent test stored a result")
	}

}

func TestCheckResultInvalid(t *testing.T) {
	cases := []struct{ id, conf string }{
		{"no_check", `  - id: "no_check"
    path: "../_plugins/decision/no_check.so"
`},
		{"wrong_check", `  - id: "wrong_check"
    path: "../_plugins/decision/wrong_check.so"
`},
		{"error_check", `  - id: "error_check"
    path: "../_plugins/decision/error_check.so"
`},
	}

	// Test model plugin initialization
	for _, c := range cases {
		config := baseConfig + "modelplugins:\n" + trivialPlugin + "decisionplugins:\n" + c.conf

		conf := cf.Get()
		err := conf.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Error loading config: %v", err)
		}
		plugins := New()

		_, err = plugins.CheckResult(generateRandomID(), c.id, make(map[string]string))
		if err == nil {
			t.Errorf("invalid CheckResult %s did not rise an error", c.id)
		}
	}

	config := baseConfig + "modelplugins:\n" + trivialPlugin + "decisionplugins:\n" + testPlugin

	conf := cf.Get()
	err := conf.LoadConfigYaml([]byte(config))
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	plugins := New()

	_, err = plugins.CheckResult(generateRandomID(), "nonexitent", make(map[string]string))
	if err == nil {
		t.Errorf("nonexistent plugin did not rise an error")
	}

}
