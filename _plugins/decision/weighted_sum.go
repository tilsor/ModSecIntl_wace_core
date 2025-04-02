// Decision Plugin that uses weighted sum algorithm to decide if a transaction should be blocked

package main

import (
	"context"
	"fmt"
	"strconv"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
	pm "github.com/tilsor/ModSecIntl_wace_lib/pluginmanager"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var wafWeight float64
var threshold float64

func InitPlugin(params map[string]string, meter metric.Meter) error {
	stringWafWeight, ok := params["waf_weight"]
	if !ok {
		return fmt.Errorf("waf_weight parameter not found")
	}
	var err error
	wafWeight, err = strconv.ParseFloat(stringWafWeight, 64)
	if err != nil {
		return fmt.Errorf("error parsing waf_weight parameter: %v", err)
	}

	stringThreshold, ok := params["threshold"]
	if !ok {
		threshold = 0.5
	} else {
		threshold, err = strconv.ParseFloat(stringThreshold, 64)
		if err != nil {
			return fmt.Errorf("error parsing threshold parameter: %v", err)
		}
	}

	// Create counter for plugin register
	ctx := context.Background()
	pluginCounter, err := meter.Int64Counter("plugin_register")
	if err != nil {
		return err
	}
	pluginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("plugin_name", "weighted_sum"), attribute.String("plugin_type", "decision")))
	return nil
}

func CheckResults(decisionInput pm.DecisionInput) (bool, error) {
	var weightedSum float64 = 0
	var weightsSum float64 = 0
	for key, value := range decisionInput.Results {
		weightedSum += value.ProbAttack * decisionInput.ModelWeight[key]
		weightsSum += decisionInput.ModelWeight[key]
	}

	stringInboundBlocking, ok := decisionInput.WAFdata["inbound_blocking"]
	if !ok {
		return false, fmt.Errorf("inbound_blocking parameter not found")
	}
	stringInboundThreshold, ok := decisionInput.WAFdata["inbound_threshold"]
	if !ok {
		return false, fmt.Errorf("inbound_threshold parameter not found")
	}

	as, err := strconv.ParseFloat(stringInboundBlocking, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing anomaly score: %v", err)
	}
	it, err := strconv.ParseFloat(stringInboundThreshold, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing anomaly score threshold: %v", err)
	}

	logger := lg.Get()
	logger.TPrintf(lg.DEBUG, decisionInput.TransactionId, "weighted_sum | anomaly score: %v anomaly score threshold: %v", as, it)

	if as >= it {
		weightedSum += wafWeight
	} else {
		weightedSum += (as / it) * wafWeight
	}
	weightsSum += wafWeight

	weightedSum /= weightsSum

	logger.TPrintf(lg.DEBUG, decisionInput.TransactionId, "weighted_sum | weighted sum: %v threshold: %v", weightedSum, threshold)
	return weightedSum > threshold, nil
}
