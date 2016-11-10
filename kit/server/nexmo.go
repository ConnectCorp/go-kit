package server

import (
	"encoding/json"
	"github.com/PuerkitoBio/rehttp"
	"github.com/njern/gonexmo"
)

// MustInitNexmo initializes the Nexmo client, or panics.
func MustInitNexmo(commonCfg *CommonConfig, nexmoCfg *NexmoConfig) *nexmo.Client {
	nexmoClient, err := nexmo.NewClientFromAPI(nexmoCfg.NexmoAPIKey, nexmoCfg.NexmoAPISecret)
	if err != nil {
		panic(err)
	}
	nexmoClient.HttpClient = MakeHTTPClientForConfig(
		commonCfg,
		rehttp.RetryAll(
			rehttp.RetryMaxRetries(defaultMaxRetries),
			rehttp.RetryAny(
				rehttp.RetryTemporaryErr(),
				rehttp.RetryStatusInterval(500, 600),
			),
		),
	)

	return nexmoClient
}
