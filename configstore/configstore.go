/*
Package configstore handles the configuration of WACE. The
configuration file is parsed, checked for errors and loaded into
memory
*/
package configstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

// ModelPluginType is an enum listing the parts of a request or
// response that a model plugin can handle.
type ModelPluginType int

// TODO: Update design document with these types
const (
	RequestHeaders ModelPluginType = iota
	RequestBody
	AllRequest
	ResponseHeaders
	ResponseBody
	AllResponse
	Everything
)

func (t ModelPluginType) String() string {
	switch t {
	case RequestHeaders:
		return "RequestHeaders"
	case RequestBody:
		return "RequestBody"
	case AllRequest:
		return "AllRequest"
	case ResponseHeaders:
		return "ResponseHeaders"
	case ResponseBody:
		return "ResponseBody"
	case AllResponse:
		return "AllResponse"
	default:
		return "Everything"
	}
}

// StringToPluginType converts a string to the corresponding model plugin type
func StringToPluginType(textType string) (ModelPluginType, error) {
	switch textType {
	case "RequestHeaders":
		return RequestHeaders, nil
	case "RequestBody":
		return RequestBody, nil
	case "AllRequest":
		return AllRequest, nil
	case "ResponseHeaders":
		return ResponseHeaders, nil
	case "ResponseBody":
		return ResponseBody, nil
	case "AllResponse":
		return AllResponse, nil
	case "Everything":
		return Everything, nil
	}
	return -1, fmt.Errorf("invalid plugin type %s", textType)
}

type modelPluginConfig struct {
	ID         string
	Path       string
	Weight     float64
	Threshold  float64
	Params     map[string]string
	PluginType ModelPluginType
}
type decisionPluginConfig struct {
	ID              string
	Path            string
	WAFweight       float64
	DecisionBalance float64
	Params          map[string]string
}

// ConfigStore stores all wacecore configuration from the config file.
type ConfigStore struct {
	ModelPlugins    map[string]modelPluginConfig
	DecisionPlugins map[string]decisionPluginConfig
	LogPath         string
	LogLevel        lg.LogLevel
	ListenAddress   string
	ListenPort      string
}

var config *ConfigStore

// Get returns or creates the unique instance of configstore
func Get() *ConfigStore {
	if config == nil {
		config = new(ConfigStore)
	}
	return config
}

type configFileModelPlugin struct {
	ID         string
	Path       string
	Weight     float64
	Threshold  float64
	Params     map[string]string
	PluginType string `yaml:"plugintype"`
}

type configFileDecisionPlugin struct {
	ID              string
	Path            string
	wafweight       float64
	decisionbalance float64
	Params          map[string]string
}

type configFileData struct {
	Logpath         string
	Loglevel        string
	Listenaddress   string
	Listenport      string
	Modelplugins    []configFileModelPlugin
	Decisionplugins []configFileDecisionPlugin
}

func checkLogging(inConf configFileData) error {
	// check logpath
	if inConf.Logpath == "" {
		return fmt.Errorf("log path empty")
	}
	_, err := os.Stat(inConf.Logpath)
	if err != nil { // check if log file does not exists already
		// Attempt to create dummy file
		var d []byte
		err = ioutil.WriteFile(inConf.Logpath, d, 0644)
		if err == nil {
			err = os.Remove(inConf.Logpath) // delete it
		}
	}
	return err
}

// CheckConfig verifies if the configuration read from the config file
// is correct.
func checkConfig(inConf configFileData) error {
	err := checkLogging(inConf)
	if err != nil {
		return fmt.Errorf("invalid log path %s: %v", inConf.Logpath, err)
	}

	// check listenport
	if inConf.Listenport != "" {
		val, err := strconv.Atoi(inConf.Listenport)
		if err != nil || (val > 65535 || val <= 0) {
			return fmt.Errorf("listen port %s should be a number between 1 and 65535", inConf.Listenport)
		}
	}

	// check modelplugins
	for _, modelP := range inConf.Modelplugins {

		if modelP.Path != "" {
			if _, err := os.Stat(modelP.Path); err != nil {
				return fmt.Errorf("%s plugin path %s: %v", modelP.ID, modelP.Path, err)
			}
		} else {
			return fmt.Errorf("%s plugin path is empty, please provide a valid path", modelP.ID)
		}
		if modelP.PluginType == "" {
			return fmt.Errorf("%s plugin type cannot be empty, please provide a valid type", modelP.ID)
		}
	}
	// check decisionplugins
	for _, decisionP := range inConf.Decisionplugins {

		if decisionP.Path != "" {
			if _, err := os.Stat(decisionP.Path); err != nil {
				return fmt.Errorf("%s plugin path %s cannot be opened: %v", decisionP.ID, decisionP.Path, err)
			}
		} else {
			return fmt.Errorf("%s plugin path is empty, please provide a valid path", decisionP.ID)
		}
	}
	return nil
}

// LoadConfigYaml loads a configuration from a yaml string
func (cs *ConfigStore) LoadConfigYaml(config []byte) error {
	var inConf configFileData

	err := yaml.Unmarshal(config, &inConf)
	if err != nil {
		return err
	}

	err = checkConfig(inConf)
	if err != nil {
		return err
	}

	cs.LogPath = inConf.Logpath
	cs.ListenAddress = inConf.Listenaddress
	cs.LogLevel, err = lg.StringToLogLevel(inConf.Loglevel)
	if err != nil {
		return err
	}
	cs.ListenPort = inConf.Listenport

	cs.ModelPlugins = make(map[string]modelPluginConfig)
	for _, modelP := range inConf.Modelplugins {
		var modelConfig modelPluginConfig
		modelConfig.ID = modelP.ID
		modelConfig.Path = modelP.Path
		modelConfig.Weight = modelP.Weight
		modelConfig.Threshold = modelP.Threshold
		modelConfig.Params = modelP.Params
		modelConfig.PluginType, err = StringToPluginType(modelP.PluginType)
		if err != nil {
			return err
		}
		cs.ModelPlugins[modelConfig.ID] = modelConfig
	}

	cs.DecisionPlugins = make(map[string]decisionPluginConfig)
	for _, decisionP := range inConf.Decisionplugins {
		var decisionConfig decisionPluginConfig
		decisionConfig.ID = decisionP.ID
		decisionConfig.Path = decisionP.Path
		decisionConfig.WAFweight = decisionP.wafweight
		decisionConfig.DecisionBalance = decisionP.decisionbalance
		decisionConfig.Params = decisionP.Params
		cs.DecisionPlugins[decisionConfig.ID] = decisionConfig
	}

	return nil
}

// LoadConfig loads the configuration from the config file to memory
func (cs *ConfigStore) LoadConfig(configFilePath string) error {
	var file, err = ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	return cs.LoadConfigYaml(file)
}
