/*
The main package of WACE.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	comm "wace/comm"
	cf "wace/configstore"

	pm "wace/pluginmanager"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

var plugins *pm.PluginManager

// transactionSync is a struct to syncronize the analysis of a given
// transaction. Each time callPlugins is executed, the counter is
// incremented. At the end of each callPlugins execution, a message is
// sent through the channel, to signal checkTransaction that it has
// finished analyzing the request. checkTransaction waits for Counter
// number of messages in the channel, before calling the decision
// plugin and sending the result to the client.
type transactionSync struct {
	Channel chan string
	Counter int
}

var (
	// channel to receive a notification when all plugins finish
	// processing a transaction
	analysisMap   = make(map[string](transactionSync))
	analysisMutex = sync.RWMutex{}
)

func addTransactionAnalysis(transactionID string) {
	analysisMutex.Lock()
	sync, exists := analysisMap[transactionID]
	if !exists {
		analysisMap[transactionID] = transactionSync{
			Channel: make(chan string),
			Counter: 1,
		}
	} else {
		sync.Counter++
	}
	analysisMutex.Unlock()
}

// Call all appropriate plugins
func callPlugins(input string, models []string, t cf.ModelPluginType, transactionID string) {
	logger := lg.Get()

	// channel to receive the status of the execution of the analysis
	// of all the model plugins executed
	modelPlugStatus := make(chan pm.ModelStatus)

	for _, id := range models {
		logger.TPrintf(lg.DEBUG, transactionID, "%s | calling from core", id)
		switch t {
		case cf.RequestHeaders, cf.RequestBody, cf.AllRequest:
			go plugins.ProcessRequest(id, input, t, transactionID, modelPlugStatus)
		case cf.ResponseHeaders, cf.ResponseBody, cf.AllResponse:
			go plugins.ProcessResponse(id, input, t, transactionID, modelPlugStatus)
		}
	}
	logger.TPrintf(lg.DEBUG, transactionID, "core | waiting for %d model plugins to finish", len(models))
	for i := 0; i < len(models); i++ {
		// Await for the execution of the model plugins
		logger.TPrintf(lg.DEBUG, transactionID, "core | Waiting for model plugin %d...", i)
		status := <-modelPlugStatus
		if status.Err == nil {
			logger.TPrintf(lg.DEBUG, transactionID, "%s | success. Result: %.5f", status.ModelID, status.Res)
		} else {
			logger.TPrintf(lg.WARN, transactionID, "%s | %v", status.ModelID, status.Err)
		}
	}

	analysisMutex.RLock()
	analysisChan := analysisMap[transactionID].Channel
	analysisMutex.RUnlock()
	analysisChan <- "done"

}

func analyzeRequest(transactionID, request string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing whole request: [%s...]", strings.Split(request, "\n")[0])
	addTransactionAnalysis(transactionID)
	go callPlugins(request, models, cf.AllRequest, transactionID)
	return 0
}

func analyzeReqLineAndHeaders(transactionID, requestLine, requestHeaders string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing request line and headers: [%s...]", requestLine)
	addTransactionAnalysis(transactionID)
	go callPlugins(requestLine+requestHeaders, models, cf.RequestHeaders, transactionID)
	return 0
}

func analyzeRequestBody(transactionID, requestBody string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing request body: [%s...]", strings.Split(requestBody, "\n")[0])
	addTransactionAnalysis(transactionID)
	go callPlugins(requestBody, models, cf.RequestBody, transactionID)
	return 0
}

func analyzeResponse(transactionID, response string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing whole response: [%s...]", strings.Split(response, "\n")[0])
	addTransactionAnalysis(transactionID)
	go callPlugins(response, models, cf.AllResponse, transactionID)
	return 0
}

func analyzeRespLineAndHeaders(transactionID, statusLine, responseHeaders string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing response line and headers: [%s...]", statusLine)
	addTransactionAnalysis(transactionID)
	go callPlugins(statusLine+responseHeaders, models, cf.ResponseHeaders, transactionID)
	return 0
}

func analyzeResponseBody(transactionID, responseBody string, models []string) int32 {
	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | analyzing response body: [%s...]", strings.Split(responseBody, "\n")[0])
	addTransactionAnalysis(transactionID)
	go callPlugins(responseBody, models, cf.ResponseBody, transactionID)
	return 0
}

func checkTransaction(transactionID, decisionPlugin string, wafParams map[string]string) (bool, error) {

	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, transactionID, "core | checking transaction")

	analysisMutex.RLock()
	sync, exists := analysisMap[transactionID]
	analysisMutex.RUnlock()

	if !exists {
		return false, fmt.Errorf("transaction with id %s does not exist", transactionID)
	}

	logger.TPrintln(lg.DEBUG, transactionID, "core | waiting for all models to finish...")

	for i := 0; i < sync.Counter; i++ {
		<-sync.Channel
	}

	logger.TPrintln(lg.DEBUG, transactionID, "core | done, checking data...")
	res, err := plugins.CheckResult(transactionID, decisionPlugin, wafParams)
	analysisMutex.Lock()
	delete(analysisMap, transactionID)
	analysisMutex.Unlock()
	if err == nil {
		logger.TPrintf(lg.DEBUG, transactionID, "core | transaction checked successfully. Blocking transaction: %t", res)
	} else {
		logger.TPrintf(lg.ERROR, transactionID, "core | could not check transaction: %v", err)
	}
	return res, err
}

func main() {
	logger := lg.Get()

	flag.Parse()
	handlers := comm.Handlers{
		SendRequest:            analyzeRequest,
		SendReqLineAndHeaders:  analyzeReqLineAndHeaders,
		SendRequestBody:        analyzeRequestBody,
		SendResponse:           analyzeResponse,
		SendRespLineAndHeaders: analyzeRespLineAndHeaders,
		SendResponseBody:       analyzeResponseBody,
		Check:                  checkTransaction,
	}
	logger.Println(lg.DEBUG, "Opening wace configuration file...")
	// Load the configuration
	configFilePath := flag.Arg(0)
	if configFilePath == "" {
		logger.Println(lg.ERROR, "ERROR: Please specify the path to the WACE configuration file as an argument")
		os.Exit(1)
	}
	conf := cf.Get()
	err := conf.LoadConfig(configFilePath)
	if err != nil {
		logger.Printf(lg.ERROR, "ERROR: could not load configuration: %v", err)
		os.Exit(1)
	}
	logger.Printf(lg.DEBUG, "Configuration loaded successfully from %s", configFilePath)

	err = logger.LoadLogger(conf.LogPath, conf.LogLevel)
	if err != nil {
		logger.Printf(lg.ERROR, "ERROR: could not open wace log file: %v", err)
		os.Exit(1)

	}
	logger.Printf(lg.DEBUG, "Writing logs to %s from now", conf.LogPath)

	logger.Println(lg.DEBUG, "Loading plugin manager...")
	plugins = pm.New()
	logger.Println(lg.DEBUG, "Plugin manager loaded")

	logger.Println(lg.DEBUG, "Server started, listening for connections...")
	err = comm.Listen(handlers, conf.ListenAddress, conf.ListenPort)
	if err != nil {
		logger.Printf(lg.ERROR, "ERROR: wace server failed: %v", err)
	}
}
