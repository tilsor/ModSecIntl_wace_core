/* Trivial Model Plugin that always returns 0 probability of attack
 */

package main

import (
	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	logger := lg.Get()
	logger.Printf(lg.WARN, "[trivial:InitPlugin] %v\n", params)
	return nil
}

// ProcessRequest always returns 0 probability of attack
func ProcessRequest(transactionID, req string) (float64, error) {
	logger := lg.Get()
	logger.TPrintf(lg.WARN, transactionID, "[trivial:ProcessRequest] \"%s\"\n", req)
	return 0.0, nil
}

// ProcessResponse always returns 0 probability of attack
func ProcessResponse(transactionID, resp string) (float64, error) {
	logger := lg.Get()
	logger.TPrintf(lg.WARN, transactionID, "[trivial:ProcessResponse] \"%s\"\n", resp)
	return 0.0, nil
}
