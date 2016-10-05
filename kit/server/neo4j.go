package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	neo4j "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"net/url"
	"time"
)

const (
	baseNeo4JInitDelay         = 25 * time.Millisecond
	maxNeo4JInitRetryCount     = 10
	defaultNeo4JDriverPoolSize = 10
)

// MustInitNeo4J initializes a Neo4J driver pool, or panics.
func MustInitNeo4J(spec *url.URL) neo4j.DriverPool {
	var driverPool neo4j.DriverPool
	var err error

	utils.MustBackoff(baseNeo4JInitDelay, maxNeo4JInitRetryCount, func() error {
		driverPool, err = neo4j.NewDriverPool(spec.String(), defaultNeo4JDriverPoolSize)
		if err != nil {
			return err
		}

		conn, err := driverPool.OpenPool()
		if err != nil {
			return err
		}
		defer conn.Close()

		return nil
	})

	return driverPool
}
