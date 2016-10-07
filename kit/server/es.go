package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"gopkg.in/olivere/elastic.v2"
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
		es, err = elastic.NewClient(elastic.SetURL(esSpec.String()), elastic.SetMaxRetries(defaultESMaxRetries), elastic.SetSniff(false), elastic.SetHttpClient(MakeAWSSigningHTTPClient(nil)))
		return err
	})

	return es
}
