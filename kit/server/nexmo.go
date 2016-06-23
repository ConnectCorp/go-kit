package server

import (
	"github.com/njern/gonexmo"
)

// MustInitNexmo initializes the Nexmo client, or panics.
func MustInitNexmo(commonCfg *CommonConfig, nexmoCfg *NexmoConfig) *nexmo.Client {
	nexmoClient, err := nexmo.NewClientFromAPI(nexmoCfg.NexmoAPIKey, nexmoCfg.NexmoAPISecret)
	if err != nil {
		panic(err)
	}
	nexmoClient.HttpClient = MakeHTTPClientForConfig(commonCfg)
	return nexmoClient
}
