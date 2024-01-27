package sapbtp

import (
	"sync"
	"time"
)

var btpTokens sync.Map

func GetToken(apiKey string) string {
	data, ok := btpTokens.Load(apiKey)
	if ok {
		tokenData := data.(tokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.AccessToken
		}
	}
	// TODO get token from BTP
	return ""
}
