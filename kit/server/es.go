package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/PuerkitoBio/rehttp"
	"gopkg.in/olivere/elastic.v5"
	"net/url"
	"time"
)

const (
	baseESInitDelay     = 25 * time.Millisecond
	maxESInitRetryCount = 10
	defaultESMaxRetries = 5
)

// MustInitES initializes an ElasticSearch client, or panics.
func MustInitES(esSpec *url.URL) *elastic.Client {
	var es *elastic.Client
	var err error

	utils.MustBackoff(baseESInitDelay, maxESInitRetryCount, func() error {
		options := []elastic.ClientOptionFunc{
			elastic.SetURL(esSpec.String()),
			elastic.SetMaxRetries(defaultESMaxRetries),
			elastic.SetSniff(false),
			elastic.SetHttpClient(MakeProdHTTPClient(func(attempt rehttp.Attempt) bool { return false })), // The ES lib already implements retries.
		}

		if esSpec.User != nil {
			password, _ := esSpec.User.Password()
			options = append(options, elastic.SetBasicAuth(esSpec.User.Username(), password))
		}
		esSpec.User = nil

		es, err = elastic.NewClient(options...)
		return err
	})

	return es
}
