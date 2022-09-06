/*
Package pluginmanager handles the communication with the model and
decision plugins
*/
package pluginmanager

import (
	"fmt"
	"plugin"
	"sync"
	cf "wace/configstore"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

// ResultData maps the model plugin ID with the corresponding analysis result.
type ResultData struct {
	pAttack map[string]float64 // maps the model id with the probabilty of attack
	weight  map[string]float64 // maps the model id with the weight for the model result
	thres   map[string]float64 // maps the model id with the threshold of the model
}

type modelPlugin struct {
	p          *plugin.Plugin
	pluginType cf.ModelPluginType
}

type decisionPlugin struct {
	p *plugin.Plugin
}

// ModelStatus stores whether there was an error while processing a
// request (response) by the modelID model plugin
type ModelStatus struct {
	ModelID string
	Res     float64
	Err     error
}

// PluginManager is the main plugin struct storing information of
// every plugin execution.
type PluginManager struct {
	modelPlugins    map[string]modelPlugin
	decisionPlugins map[string]decisionPlugin
	results         map[string]*ResultData // maps the transactionID with the results of the models
	resultsMutex    sync.RWMutex
}

// New creates a new PluginManager instance.
func New() *PluginManager {
	pm := new(PluginManager)
	conf := cf.Get()
	logger := lg.Get()
	pm.resultsMutex = sync.RWMutex{}

	// Loading of model plugins
	pm.modelPlugins = make(map[string]modelPlugin)
	for _, data := range conf.ModelPlugins {
		tp, err := plugin.Open(data.Path)
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		f, err := tp.Lookup("InitPlugin")
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		initPlugin, ok := f.(func(map[string]string) error)
		if !ok {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: invalid InitPlugin function type", data.ID)
			continue
		}
		err = initPlugin(data.Params)
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		modelPluginLoaded := modelPlugin{tp, data.PluginType}
		pm.modelPlugins[data.ID] = modelPluginLoaded
	}

	pm.decisionPlugins = make(map[string]decisionPlugin)
	// Loading of decision plugins
	for _, data := range conf.DecisionPlugins {
		tp, err := plugin.Open(data.Path)
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		f, err := tp.Lookup("InitPlugin")
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		initPlugin, ok := f.(func(map[string]string) error)
		if !ok {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: invalid InitPlugin function type", data.ID)
			continue
		}
		err = initPlugin(data.Params)
		if err != nil {
			logger.Printf(lg.WARN, "| %s | cannot load plugin: %v", data.ID, err)
			continue
		}
		decisionPluginLoaded := decisionPlugin{tp}
		pm.decisionPlugins[data.ID] = decisionPluginLoaded
	}
	pm.results = make(map[string]*ResultData)
	return pm
}

// ProcessRequest analyses the request with the given model, and
// writes the result to the channel modelPlugStatus
func (p *PluginManager) ProcessRequest(modelID, req string, t cf.ModelPluginType, transactID string, modelPlugStatus chan ModelStatus) {
	conf := cf.Get()

	mp, exists := p.modelPlugins[modelID]
	if !exists {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("model plugin not found")}
		return
	}

	// check if the plugin is capable of analyzing the indicated part of the transaction
	switch mp.pluginType {
	case cf.RequestHeaders, cf.RequestBody, cf.AllRequest:
		if mp.pluginType != t {
			modelPlugStatus <- ModelStatus{ModelID: modelID,
				Err: fmt.Errorf("plugin type %v cannot process a request with incompatible type %v", mp.pluginType, t)}
			return
		}
	case cf.Everything:
		// Process the request
	default:
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("plugin only handles responses")}
		return
	}

	pR, err := mp.p.Lookup("ProcessRequest")
	if err != nil {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("ProcessRequest lookup failed: %v", err)}
		return
	}
	processRequest, ok := pR.(func(string, string) (float64, error))
	if !ok {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("ProcessRequest lookup failed: invalid function type")}
		return
	}

	res, err := processRequest(transactID, req)

	if err == nil {
		// store the results
		p.resultsMutex.Lock()
		_, existData := p.results[transactID]
		if !existData {
			p.results[transactID] = new(ResultData)
		}
		if p.results[transactID].pAttack == nil {
			p.results[transactID].pAttack = make(map[string]float64)
		}
		if p.results[transactID].weight == nil {
			p.results[transactID].weight = make(map[string]float64)
		}
		if p.results[transactID].thres == nil {
			p.results[transactID].thres = make(map[string]float64)
		}
		p.results[transactID].pAttack[modelID] = res
		p.results[transactID].weight[modelID] = conf.ModelPlugins[modelID].Weight
		p.results[transactID].thres[modelID] = conf.ModelPlugins[modelID].Threshold
		p.resultsMutex.Unlock()
	}
	modelPlugStatus <- ModelStatus{ModelID: modelID, Res: res, Err: nil}
}

// ProcessResponse analyses the response with the given model
func (p *PluginManager) ProcessResponse(modelID, resp string, t cf.ModelPluginType, transactID string, modelPlugStatus chan ModelStatus) {
	mp, exists := p.modelPlugins[modelID]
	if !exists {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("model plugin not found")}
		return
	}

	// check if the plugin is capable of analyzing the indicated part of the transaction
	switch mp.pluginType {
	case cf.ResponseHeaders, cf.ResponseBody, cf.AllResponse:
		if mp.pluginType != t {
			modelPlugStatus <- ModelStatus{ModelID: modelID,
				Err: fmt.Errorf("plugin type %v cannot process a request with incompatible type %v", mp.pluginType, t)}
			return
		}
	case cf.Everything:
		// Process the response
	default:
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("plugin only handles requests")}
		return
	}

	pR, err := mp.p.Lookup("ProcessResponse")
	if err != nil {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("ProcessResponse lookup failed: %v", err)}
		return
	}
	processResponse, ok := pR.(func(string, string) (float64, error))
	if !ok {
		modelPlugStatus <- ModelStatus{ModelID: modelID, Err: fmt.Errorf("ProcessResponse lookup failed: invalid function type")}
		return
	}
	conf := cf.Get()

	res, err := processResponse(transactID, resp)
	if err == nil {
		p.resultsMutex.Lock()

		_, existData := p.results[transactID]
		if !existData {
			p.results[transactID] = new(ResultData)
		}
		if p.results[transactID].pAttack == nil {
			p.results[transactID].pAttack = make(map[string]float64)
		}
		if p.results[transactID].weight == nil {
			p.results[transactID].weight = make(map[string]float64)
		}
		if p.results[transactID].thres == nil {
			p.results[transactID].thres = make(map[string]float64)
		}
		p.results[transactID].pAttack[modelID] = res
		p.results[transactID].weight[modelID] = conf.ModelPlugins[modelID].Weight
		p.results[transactID].thres[modelID] = conf.ModelPlugins[modelID].Threshold
		p.resultsMutex.Unlock()
	}
	modelPlugStatus <- ModelStatus{ModelID: modelID, Res: res, Err: nil}
}

// CheckResult is in charge of calling the decision plugin with id decisionID over the
// transaction with id transactID
func (p *PluginManager) CheckResult(transactID, decisionID string, wafParams map[string]string) (bool, error) {
	logger := lg.Get()
	dp, exists := p.decisionPlugins[decisionID]
	if !exists {
		return false, fmt.Errorf("decision plugin %s not found", decisionID)
	}
	cR, err := dp.p.Lookup("CheckResults")
	if err != nil {
		return false, fmt.Errorf("CheckResults lookup failed for %s plugin: %v", decisionID, err.Error())
	}
	checkResults, ok := cR.(func(string, map[string]float64, map[string]float64, map[string]float64, map[string]string) (bool, error))
	if !ok {
		return false, fmt.Errorf("CheckResults lookup failed for %s plugin: invalid function type", decisionID)
	}

	p.resultsMutex.Lock()
	if p.results[transactID] == nil { // no analysis result found for the transaction
		p.resultsMutex.Unlock()
		return false, fmt.Errorf("no analysis data found for transaction %s", transactID)
	}
	res, err := checkResults(transactID, p.results[transactID].pAttack, p.results[transactID].weight, p.results[transactID].thres, wafParams)

	logger.TPrintf(lg.INFO, transactID, "%s | transaction checked. Block: %t ", decisionID, res)

	// clean of the result data after the check
	delete(p.results, transactID)
	p.resultsMutex.Unlock()

	return res, err
}
