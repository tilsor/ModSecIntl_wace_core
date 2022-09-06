/* Wrong Init model plugin that has wrong InitPlugin type
 */

package main

// InitPlugin intitalizes the plugins (does nothing in this case)
func InitPlugin(params map[string]string) error {
	return nil
}

// ProcessRequest always returns 0 probability of attack
func ProcessRequest(transactionID, req string, k int) (float64, error) {
	return 0.0, nil
}

func ProcessResponse() (float64, error) {
	return 0.0, nil
}
