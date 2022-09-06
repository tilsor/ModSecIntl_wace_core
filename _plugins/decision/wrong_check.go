/* Trivial Decision Plugin that always returns no attack
 */

package main

import (
	lg "github.com/tilsor/ModSecIntl_logging/logging"
)

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	logger := lg.Get()
	logger.Printf(lg.WARN, "[simple:InitPlugin] %v\n", params)
	return nil
}

// CheckResults returns true (block traffic) if WAF says so, and false
// in other case.
func CheckResults() (bool, error) {
	return false, nil
}
