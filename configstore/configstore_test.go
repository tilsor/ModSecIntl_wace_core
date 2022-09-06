package configstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var validConfig = []byte(`---
logpath: "/dev/stderr"
loglevel: "DEBUG"
listenport: "50051"
modelplugins:
  - id: "trivial"
    path: "../_plugins/model/trivial.so"
    weight: 1
    threshold: 0.5
    params:
      d: "sds"
      b: "dnid"
      e: "dofnno"
    # plugintype: "RequestHeaders"
    plugintype: "Everything"
  - id: "trivial2"
    path: "../_plugins/model/trivial2.so"
    weight: 2
    threshold: 0.1
    params:
      a: "sdsds"
      b: "sdfjdnid"
      c: "kfoskdofnno"
    plugintype: "Everything"
decisionplugins:
  - id: "test"
    path: "../_plugins/decision/test.so"
    wafweight: 0.5
    decisionbalance: 0.5
    params:
      ssdaf: "sdsds"
      dsfb: "sdfjdnid"
      csfd: "kfoskdofnno"
`)

func TestLoadConfigYamlEmpty(t *testing.T) {
	cs := Get()

	err := cs.LoadConfigYaml([]byte(`---`))
	if err == nil {
		t.Errorf("empty config does not return error")
	}
}

func TestLoadConfigYamlValid(t *testing.T) {
	cs := Get()

	err := cs.LoadConfigYaml(validConfig)
	if err != nil {
		t.Errorf("valid config returned error: %v", err)
	}
}

func TestLoadConfigYamlInvalid(t *testing.T) {
	cs := Get()

	err := cs.LoadConfigYaml([]byte(`()=)(/&/()~@#~½¬{[{½¬½---sfdjlskjfs#@~sjdfa`))

	if err == nil {
		t.Errorf("invalid config does not return error")
	}
}

func TestLoadConfigYamlLogLevel(t *testing.T) {
	cs := Get()

	values := []string{
		"a",
		"4",
		"0",
	}

	for _, v := range values {
		config := `---
listenport: 50051
logpath: "/dev/null"
loglevel: ` + v
		err := cs.LoadConfigYaml([]byte(config))
		if err == nil {
			t.Errorf("invalid log level %v does not return error", v)
		}
	}
}

func TestLoadConfigYamlListenPort(t *testing.T) {
	cs := Get()

	values := []string{
		"a",
		"9999999999999999",
		"0",
	}

	for _, v := range values {
		config := `---
loglevel: ERROR
logpath: "/dev/null"
listenport: ` + v
		err := cs.LoadConfigYaml([]byte(config))
		if err == nil {
			t.Errorf("invalid listen port %v does not return error", v)
		}
	}
}

func TestLoadConfigYamlPluginType(t *testing.T) {
	cs := Get()

	err := cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
modelplugins:
  - id: "testplugin"
    path: "../_plugins/model/trivial.so"
    plugintype: InvalidPluginType
`))
	if err == nil {
		t.Errorf("invalid plugin type does not return error")
	}

	err = cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
modelplugins:
  - id: "testplugin"
    path: "../_plugins/model/trivial.so"
    plugintype: ""
`))
	if err == nil {
		t.Errorf("empty plugin type does not return error")
	}

	err = cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
modelplugins:
  - id: "testplugin"
    path: "../_plugins/model/nonexistent.so"
    plugintype: "RequestHeaders"
`))
	if err == nil {
		t.Errorf("nonexistent model plugin path does not return error")
	}

	err = cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
modelplugins:
  - id: "testplugin"
    path: ""
    plugintype: "RequestHeaders"
`))
	if err == nil {
		t.Errorf("empty plugin path does not return error")
	}

	err = cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
decisionplugins:
  - id: "test"
    path: ""
`))
	if err == nil {
		t.Errorf("empty decision plugin path does not return error")
	}

	err = cs.LoadConfigYaml([]byte(`---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
decisionplugins:
  - id: "testplugin"
    path: "../_plugins/decision/nonexistent.so"
`))
	if err == nil {
		t.Errorf("nonexistent decision plugin path does not return error")
	}

	values := []string{
		"RequestHeaders",
		"RequestBody",
		"AllRequest",
		"ResponseHeaders",
		"ResponseBody",
		"AllResponse",
		"Everything",
	}

	for _, v := range values {
		config := `---
loglevel: ERROR
logpath: /dev/null
listenport: 50051
modelplugins:
  - id: "testplugin"
    path: "../_plugins/model/trivial.so"
    plugintype: "` + v + `"
`
		err = cs.LoadConfigYaml([]byte(config))
		if err != nil {
			t.Errorf("Plugin type %s returns error: %v", v, err)
		}

		if fmt.Sprint(cs.ModelPlugins["testplugin"].PluginType) != v {
			t.Errorf("Stored plugin type is %v, expected %v", cs.ModelPlugins["testplugin"].PluginType, v)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	cs := Get()

	err := cs.LoadConfig("")
	if err == nil {
		t.Errorf("empty config file path does not return error")
	}

	err = cs.LoadConfig("/dev/null")
	if err == nil {
		t.Errorf("empty config file contents does not return error")
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "configstore_test-")
	if err != nil {
		t.Errorf("cannot create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(validConfig); err != nil {
		t.Errorf("failed to write to temporary file: %v", err)
	}
	err = cs.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("valid config file returned error: %v", err)
	}
}

func TestInvalidLogging(t *testing.T) {
	cs := Get()

	err := cs.LoadConfigYaml([]byte(`---
loglevel: INVALIDLOGLEVEL
logpath: /dev/null
listenport: 50051
`))
	if err == nil {
		t.Errorf("invalid log level does not return error")
	}

	if _, err = os.Stat("./configstore_test.log"); err == nil {
		err = os.Remove("./configstore_test.log")
		if err != nil {
			t.Errorf("could not remove ./configstore_test.log")
		}
	}

	err = cs.LoadConfigYaml([]byte(`---
listenport: 50051
loglevel: ERROR
logpath: ./configstore_test.log`))

	if err != nil {
		t.Errorf("Error loading config  with nonexistent file: %v", err)
	}

	err = cs.LoadConfigYaml([]byte(`---
listenport: 50051
loglevel: ERROR
logpath: /usr/configstore_test.log`))

	if err == nil {
		t.Errorf("non existent log file in directory without permissions does not rise error")
	}

}
