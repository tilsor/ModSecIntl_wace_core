/*
The main package of WACE.
*/
package main

import (
	// _ "net/http/pprof" // DEBUG
	// "net/http" // DEBUG
	// "runtime/debug" // DEBUG

	"context"
	"flag"
	"fmt"
	"os"
	// "runtime"
	"strconv"
	"time"

	"strings"
	// "sync"
	comm "wace/comm"
	// cf "wace/configstore"
	cf "github.com/tilsor/ModSecIntl_wace_lib/configstore"

	wace "github.com/tilsor/ModSecIntl_wace_lib"

	lg "github.com/tilsor/ModSecIntl_logging/logging"

	"gopkg.in/yaml.v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// "go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	// "go.opentelemetry.io/otel/attribute"
)

// Analyze(modelsTypeAsString, transactionId, payload string, models []string)

type WaceModels struct {
	reqHeadModelIDs  []string
	reqBodyModelIDs  []string
	reqModelIDs      []string
	respHeadModelIDs []string
	respBodyModelIDs []string
	respModelIDs     []string
}

type generalConfig struct {
	otelURL              string
	waceModels           *WaceModels
	waceDecisions        []string
	earlyBlocking        bool
	crsVersion           string
	ruleIdsForExceptions map[string]int
	logPath              string
	logLevel             lg.LogLevel
	listenAddress        string
	listenPort           string
	histogramType		 string
}

// WaceGeneralConfigFileData holds the general configuration data from the config file
type WaceGeneralConfigFileData struct {
	cf.ConfigFileData    `yaml:",inline"`
	Options              map[string]string `yaml:"options"`
	RuleIdsForExceptions map[string]int    `yaml:"ruleidsforexceptions"`
}

// WaceAppConfigFileData holds the application configuration data from the config file
type WaceAppConfigFileData struct {
	ModelIds   []string `yaml:"modelids"`
	DecisionId string   `yaml:"decisionid"`
	Options    map[string]string
}

// LoadConfig loads the general configuration from the config file to memory
func (g *generalConfig) LoadConfig(configFilePath string) error {
	var file, err = os.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	return g.LoadGeneralConfigYaml(file)
}

// LoadGeneralConfigYaml loads the general configuration from the config file to memory
func (g *generalConfig) LoadGeneralConfigYaml(config []byte) error {
	var inConf WaceGeneralConfigFileData

	err := yaml.Unmarshal(config, &inConf)
	if err != nil {
		return err
	}
	for key, value := range inConf.Options {
		if key == "early_blocking" {
			g.earlyBlocking = value == "true"
		} else if key == "crs_version" {
			g.crsVersion = value
		} else if key == "otelurl" {
			g.otelURL = value
		} else if key == "listenaddress" {
			g.listenAddress = value
		} else if key == "listenport" {
			g.listenPort = value
		} else if key == "histogram_kind" {
			g.histogramType = value
		}
	}
	if g.ruleIdsForExceptions == nil {
		g.ruleIdsForExceptions = make(map[string]int)
	}
	for key, value := range inConf.RuleIdsForExceptions {
		g.ruleIdsForExceptions[key] = value
	}

	err = cf.Get().SetConfig(inConf.ConfigFileData)
	if err != nil {
		return err
	}

	gConfig.logPath = inConf.ConfigFileData.Logpath
	gConfig.logLevel, err = lg.StringToLogLevel(inConf.ConfigFileData.Loglevel)
	if err != nil {
		return err
	}

	g.waceModels = NewWaceDefaultModelsConfig()
	for _, decision := range inConf.Decisionplugins {
		g.waceDecisions = append(g.waceDecisions, decision.ID)
	}

	return err
}

// NewWaceDefaultModelsConfig creates the default WaceModels with the models stored in the WACE ConfigStore
func NewWaceDefaultModelsConfig() *WaceModels {
	conf := cf.Get()
	reqHeadModelIDs := []string{}
	reqBodyModelIDs := []string{}
	reqModelIDs := []string{}
	respHeadModelIDs := []string{}
	respBodyModelIDs := []string{}
	respModelIDs := []string{}
	for _, model := range conf.ModelPlugins {
		if model.PluginType.String() == "RequestHeaders" {
			reqHeadModelIDs = append(reqHeadModelIDs, model.ID)
		} else if model.PluginType.String() == "RequestBody" {
			reqBodyModelIDs = append(reqBodyModelIDs, model.ID)
		} else if model.PluginType.String() == "AllRequest" {
			reqModelIDs = append(reqModelIDs, model.ID)
		} else if model.PluginType.String() == "ResponseHeaders" {
			respHeadModelIDs = append(respHeadModelIDs, model.ID)
		} else if model.PluginType.String() == "ResponseBody" {
			respBodyModelIDs = append(respBodyModelIDs, model.ID)
		} else if model.PluginType.String() == "AllResponse" {
			respModelIDs = append(respModelIDs, model.ID)
		}
	}
	return &WaceModels{reqHeadModelIDs, reqBodyModelIDs, reqModelIDs, respHeadModelIDs, respBodyModelIDs, respModelIDs}
}

func initTransaction(transactionID string) int32 {
	wace.InitTransaction(transactionID)
	return 0
}

func analyzeRequest(transactionID, request string, models []string) int32 {
	wace.Analyze("AllRequest", transactionID, request, models)
	return 0
}

func analyzeReqLineAndHeaders(transactionID, requestLine, requestHeaders string, models []string) int32 {
	wace.Analyze("RequestHeaders", transactionID, requestLine, models)
	return 0
}

func analyzeRequestBody(transactionID, requestBody string, models []string) int32 {
	wace.Analyze("RequestBody", transactionID, requestBody, models)
	return 0
}

func analyzeResponse(transactionID, response string, models []string) int32 {
	wace.Analyze("AllResponse", transactionID, response, models)
	return 0
}

func analyzeRespLineAndHeaders(transactionID, statusLine, responseHeaders string, models []string) int32 {
	wace.Analyze("ResponseHeaders", transactionID, responseHeaders, models)
	return 0
}

func analyzeResponseBody(transactionID, responseBody string, models []string) int32 {
	wace.Analyze("ResponseBody", transactionID, responseBody, models)
	return 0
}

func checkTransaction(transactionID, decisionPlugin string, wafParams map[string]string) (bool, error) {
	return wace.CheckTransaction(transactionID, decisionPlugin, wafParams)
}

func closeTransaction(transactionID string, metrics map[string]string) int32 {
	for i, v := range metrics {
		if i == "Response_code" {
			processed, err := meter.Int64Counter("http.client.request.processed.total")
			if err != nil {
				logger.TPrintln(lg.ERROR, transactionID, "Error getting request counter: "+err.Error())
			} else {
				vInt, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					logger.TPrintln(lg.ERROR, transactionID, "Error getting response code: "+err.Error())
				} else {
					processed.Add(ctx, 1, metric.WithAttributes(semconv.HTTPResponseStatusCode(int(vInt))))
					
					logger.TPrintln(lg.DEBUG, transactionID, "Metric "+i+" : "+v)
				}
			}
		} else {
			duration, err := meter.Float64Histogram("http.client." + strings.ToLower(i) + ".duration.milliseconds")
			if err != nil {
				logger.TPrintln(lg.ERROR, transactionID, "Error getting request histogram: "+err.Error())
			} else {
				if s, err := strconv.ParseFloat(v, 64); err == nil {
					duration.Record(ctx, s/1000)
					logger.TPrintln(lg.DEBUG, transactionID, "Metric "+i+" : "+v)
				}
			}
		}
	}

	if useManualReader { // TODO: use configstore
		//Collect Metrics
		collectedMetrics := &metricdata.ResourceMetrics{}
		globalManualReader.Collect(ctx,collectedMetrics)

		//Export Metrics
		globalMetricExporter.Export(ctx,collectedMetrics)
	}

	wace.CloseTransaction(transactionID)
	return 0
}

var gConfig *generalConfig
var ctx = context.Background()
var meter metric.Meter
var logger = lg.Get()

// func callGC(){ // DEBUG
// 	logger.Println(lg.DEBUG, "Calling Go Garbage Collector...")

// 	runtime.GC()
// 	debug.FreeOSMemory()

// 	// fmt.Println(runtime.NumGoroutine())

// 	// var m runtime.MemStats
// 	// runtime.ReadMemStats(&m)
// 	// fmt.Println(m.Mallocs-m.Frees)
// }

func main() {

	// debug.SetGCPercent(1) // DEBUG
	// debug.SetMemoryLimit(16192000)

	// go func() {
    //     http.ListenAndServe("localhost:9000", nil) // DEBUG
    // }()

	// go func() { // DEBUG
	// 	for {
	// 		time.Sleep(time.Microsecond * 5000000)
	// 		callGC()
	// 	}
	// } ()

	flag.Parse()
	handlers := comm.Handlers{
		SendRequest:            analyzeRequest,
		SendReqLineAndHeaders:  analyzeReqLineAndHeaders,
		SendRequestBody:        analyzeRequestBody,
		SendResponse:           analyzeResponse,
		SendRespLineAndHeaders: analyzeRespLineAndHeaders,
		SendResponseBody:       analyzeResponseBody,
		Check:                  checkTransaction,
		Init:                   initTransaction,
		Close:                  closeTransaction,
	}

	logger.Println(lg.DEBUG, "Opening wace configuration file...")
	// Load the configuration
	configFilePath := flag.Arg(0)

	gConfig = new(generalConfig)
	err := gConfig.LoadConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error loading general config: %v", err)
		os.Exit(1)
	}

	InitMetrics(ctx, gConfig.otelURL, gConfig.histogramType)
	wace.Init(getWaceMeter())

	err = logger.LoadLogger(gConfig.logPath, gConfig.logLevel)
	if err != nil {
		logger.Printf(lg.ERROR, "ERROR: could not open wace log file: %v", err)
		os.Exit(1)
	}
	logger.Printf(lg.DEBUG, "Writing logs to %s from now", gConfig.logPath)

	logger.Println(lg.DEBUG, "Server started, listening for connections...")
	err = comm.Listen(handlers, gConfig.listenAddress, gConfig.listenPort)
	if err != nil {
		logger.Printf(lg.ERROR, "ERROR: wace server failed: %v", err)
	}
}

var serviceName = semconv.ServiceNameKey.String("wace-modsec-service")

// initConn creates a gRPC connection to the OpenTelemetry Collector. It returns the connection object and an error if the connection fails.
// This function is based on the example provided by OpenTelemetry Go contrib repository.
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go
func initConn(url string) (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient(url,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

// initMeterProvider initializes an OTLP exporter, and configures the corresponding meter provider.
func initMeterProvider(ctx context.Context, res *resource.Resource, url, histogram_kind string) (func(context.Context) error, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	if url != "" {
		conn, err := initConn(url)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection to Otel Collector: %w", err)
		}

		metricExporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
		}
	}

	globalMetricExporter = metricExporter

	var meterProvider *sdkmetric.MeterProvider
	meterProvider = &sdkmetric.MeterProvider{}

	if histogram_kind == "delta" {
		useManualReader = true // TODO: use configstore
		globalManualReader = sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(
			func (kind sdkmetric.InstrumentKind) metricdata.Temporality {
				switch kind {
				case sdkmetric.InstrumentKindHistogram:
					return metricdata.DeltaTemporality
				default:
					return metricdata.CumulativeTemporality
				}
			},
		))
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(globalManualReader),
			sdkmetric.WithResource(res),
		)
	} else {
		useManualReader = false // TODO: use configstore
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(2*time.Second))),
			sdkmetric.WithResource(res),
		)
	}

	// Check if MeterProvider is already setted
	// if otel.GetMeterProvider() != nil {
	// 	//fmt.Printf("MeterProvider already setted")
	// } else {
	// 	//fmt.Printf("MeterProvider not setted")
	// }

	globalMeterProvider = meterProvider
	meter = globalMeterProvider.Meter("wace-modsec")

	return meterProvider.Shutdown, nil
}

var useManualReader bool // TODO: use configstore
var globalMetricExporter sdkmetric.Exporter
var globalManualReader *sdkmetric.ManualReader
var globalMeterProvider *sdkmetric.MeterProvider

// getWaceMeter returns the meter for the WACE instrumentation.
func getWaceMeter() metric.Meter {
	return globalMeterProvider.Meter("wace")
}

// InitMetrics initializes the OpenTelemetry metrics instrumentation.
func InitMetrics(ctx context.Context, url, histogram_kind string) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			serviceName,
		),
	)
	if err != nil {
		panic(err)
	}

	_, err = initMeterProvider(ctx, res, url, histogram_kind)
	if err != nil {
		panic(err)
	}
	// defer func() {
	// 	if err := shutdownMeterProvider(ctx); err != nil {
	// 		panic(err) // TODO handle error
	// 	}
	// }()
}
