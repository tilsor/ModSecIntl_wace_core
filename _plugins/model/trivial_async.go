/* Trivial Async Model Plugin that always returns 0 probability of attack and sleeps for a given time, default 1 second
 */

package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
	pm "github.com/tilsor/ModSecIntl_wace_lib/pluginmanager"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var sleepTime float64

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string, meter metric.Meter) error {
	logger := lg.Get()
	logger.Printf(lg.WARN, "[trivial_async:InitPlugin] %v\n", params)
	stringSleepTime, ok := params["sleep_time"]
	if !ok {
		sleepTime = 1.0
	} else {
		var err error
		sleepTime, err = strconv.ParseFloat(stringSleepTime, 64)
		if err != nil {
			return fmt.Errorf("error parsing sleep_time parameter: %v", err)
		}
	}
	ctx := context.Background()
	pluginCounter, err := meter.Int64Counter("plugin_register")
	if err != nil {
		return err
	}
	pluginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("plugin_name", "trivial_async"), attribute.String("plugin_type", "model")))
	return nil
}

func InitPluginAsync(params map[string]string, meter metric.Meter, natsManager func(func(pm.ModelInput) (pm.ModelResults, error))) error {
	InitPlugin(params, meter)
	natsManager(Process)
	return nil
}

func Process(input pm.ModelInput) (pm.ModelResults, error) {
	time.Sleep(time.Duration(sleepTime) * time.Second)
	logger := lg.Get()
	logger.TPrintf(lg.WARN, input.TransactionId, "[trivial_async:Process] \"%s\"\n", input.Payload)
	result := pm.ModelResults{
		ProbAttack: 0.0,
		Data:       make(map[string]interface{}),
	}
	return result, nil
}
