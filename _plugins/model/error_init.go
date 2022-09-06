/* Error Init model plugin that raises an error in InitPlugin
 */

package main

import "errors"

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	return errors.New("Some error")
}

// ProcessRequest always returns 0 probability of attack
func ProcessRequest(transactionID, req string) (float64, error) {
	return 0.0, errors.New("Some error")
}
