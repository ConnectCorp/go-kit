package server

import (
	"net/url"
	neo4j "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/ConnectCorp/go-kit/kit/utils"
	"time"
)

const (
	baseNeo4JInitDelay     = 25 * time.Millisecond
	maxNeo4JInitRetryCount = 10
	defaultNeo4JDriverPoolSize = 10
)

// MustInitNeo4J initializes a Neo4J driver pool, or panics.
func MustInitNeo4J(spec *url.URL) neo4j.DriverPool {
	var driverPool neo4j.DriverPool
	var err error

	utils.MustBackoff(baseNeo4JInitDelay, maxNeo4JInitRetryCount, func() error {
		driverPool, err = neo4j.NewDriverPool(spec.String(), defaultNeo4JDriverPoolSize)
		return err
	})

	return driverPool
}