/* Trivial Decision Plugin that always returns no attack
 */

package main

import (
	"strconv"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

func InitPlugin(params map[string]string) error {
	return nil
}

func CheckResults(transactionID string, modelRes map[string]float64, modelWeight map[string]float64, modelThres map[string]float64, WAFdata map[string]string) (bool, error) {
	logger := lg.Get()
	var totalModelW float64 = 0
	var modelDetectionCount int = 0
	var totalModelProb float64 = 0
	for key, value := range modelRes {
		logger.TPrintf(lg.DEBUG, transactionID, "simple | model_id: %v result: %v threshold: %v", key, value, modelThres[key])
		if value >= modelThres[key] {
			modelDetectionCount++
			totalModelW += modelWeight[key]
		}
	}
	// if we have some model results
	if modelDetectionCount > 0 {
		totalModelProb = totalModelW / float64(modelDetectionCount)
	}
	if len(WAFdata) != 0 {
		as, _ := strconv.Atoi(WAFdata["anomalyscore"])
		it, _ := strconv.Atoi(WAFdata["inboundthreshold"])
		logger.TPrintf(lg.DEBUG, transactionID, "ModSecurity | Anomaly score: %v Anomaly score threshold: %v ", as, it)

		if as >= it && totalModelProb > 0.5 { // modsec wants to block
			return true, nil
		}
	}
	return false, nil
}
