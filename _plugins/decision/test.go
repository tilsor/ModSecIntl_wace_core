/* Trivial Test Decision Plugin that always returns no attack
 */

package main

import (
	"strconv"

	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	logger := lg.Get()
	logger.Printf(lg.WARN, "[test:InitPlugin] %v\n", params)

	return nil
}

// CheckResults returns true (block traffic) if WAF says so, and false
// in other case.
func CheckResults(transactionID string, modelRes map[string]float64, modelWeight map[string]float64, modelThres map[string]float64, wafData map[string]string) (bool, error) {
	logger := lg.Get()

	logger.TPrintf(lg.WARN, transactionID, "[test:CheckResults]\n  modelRes: %v\n  modelWeight: %v\n  modelThres: %v\n  wafData: %v\n", modelRes, modelWeight, modelThres, wafData)

	if len(wafData) != 0 {
		as, err := strconv.Atoi(wafData["anomalyscore"])
		if err != nil {
			return false, err
		}
		it, err := strconv.Atoi(wafData["inboundthreshold"])
		if err != nil {
			return false, err
		}
		if as >= it { // modsec wants to block
			return true, nil
		}
	}
	return false, nil
}
