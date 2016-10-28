package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	neo4j "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"net/url"
	"time"
)

const (
	baseNeo4JInitDelay     = 25 * time.Millisecond
	maxNeo4JInitRetryCount = 10
)

// Neo4JConnProvider is a connection provider for a neo4j database.
type Neo4JConnProvider struct {
	spec   string
	driver neo4j.Driver
}

// GetConn returns a new connection to neo4j.
func (n *Neo4JConnProvider) GetConn() (neo4j.Conn, error) {
	return n.driver.OpenNeo(n.spec)
}

// MustInitNeo4J initializes a Neo4J driver pool, or panics.
func MustInitNeo4J(spec *url.URL) *Neo4JConnProvider {
	p := &Neo4JConnProvider{
		spec:   spec.String(),
		driver: neo4j.NewDriver(),
	}

	utils.MustBackoff(baseNeo4JInitDelay, maxNeo4JInitRetryCount, func() error {
		conn, err := p.GetConn()
		if err != nil {
			return err
		}
		defer conn.Close()
		return nil
	})

	return p
}
