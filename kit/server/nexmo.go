package server

import (
	"encoding/json"
	"github.com/PuerkitoBio/rehttp"
	"github.com/njern/gonexmo"
	"io/ioutil"
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
				nexmoThrottleRetry,
			),
		),
	)

	return nexmoClient
}

func nexmoThrottleRetry(attempt rehttp.Attempt) bool {
	bytes, err := ioutil.ReadAll(attempt.Response.Body)
	if err == nil {
		return false
	}

	var resp struct {
		Status string `json:"status"`
	}

	err = json.Unmarshal(bytes, resp)
	if err == nil {
		return false
	}

	return resp.Status == "1"
}
