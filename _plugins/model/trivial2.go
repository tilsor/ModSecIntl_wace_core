/* Trivial Model Plugin that always returns 0 probability of attack
 */

package main

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	return nil
}

// ProcessRequest always returns 0 probability of attack
func ProcessRequest(transactionID, req string) (float64, error) {
	// fmt.Println("[trivial2] Processing Request: " + req)
	return 0.0, nil
}

// ProcessResponse always returns 0 probability of attack
func ProcessResponse(transactionID, resp string) (float64, error) {
	// fmt.Println("[trivial2] Processing Request: " + req)
	return 0.0, nil
}
