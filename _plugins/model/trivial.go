/* Trivial Model Plugin that always returns 0 probability of attack
 */

package main

import (
	"context"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
	pm "github.com/tiroa-tilsor/wacelib/pluginmanager"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string, meter metric.Meter) error {
	logger := lg.Get()
	logger.Printf(lg.WARN, "[trivial:InitPlugin] %v\n", params)
	// Create counter for plugin register
	ctx := context.Background()
	pluginCounter, err := meter.Int64Counter("plugin_register")
	if err != nil {
		return err
	}
	pluginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("plugin_name", "trivial"), attribute.String("plugin_type", "model")))
	return nil
}

func InitPluginAsync(params map[string]string, natsManager func(func(pm.ModelInput) (pm.ModelResults, error))) error {
	natsManager(Process)
	return nil
}

func Process(input pm.ModelInput) (pm.ModelResults, error) {
	logger := lg.Get()
	logger.TPrintf(lg.WARN, input.TransactionId, "[trivial:Process] \"%s\"\n", input.Payload)
	result := pm.ModelResults{
		ProbAttack: 0.0,
		Data:       make(map[string]interface{}),
	}
	return result, nil
}
