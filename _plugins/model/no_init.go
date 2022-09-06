/* No Init model plugin that has no InitPlugin function
 */

package main

// ProcessRequest always returns 0 probability of attack
func ProcessRequest(transactionID, req string) (float64, error) {
	return 0.0, nil
}
