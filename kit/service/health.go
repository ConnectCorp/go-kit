package service

// HealthChecker describes a health check operation.
type HealthChecker interface {
	CheckHealth() error
}

// RoutingHealthChecker is a basic HealthChecker that verifies the health check endpoint is reachable.
type RoutingHealthChecker struct {
	// Intentionally empty.
}

// NewRoutingHealthChecker initializes a new RoutingHealthChecker.
func NewRoutingHealthChecker() *RoutingHealthChecker {
	return &RoutingHealthChecker{}
}

// CheckHealth implements the HealthChecker interface.
func (r *RoutingHealthChecker) CheckHealth() error {
	return nil
}

// CompoundHealthChecker combines multiple health checks.
type CompoundHealthChecker struct {
	healthCheckers []HealthChecker
}

// NewCompoundHealthChecker initializes a new CompoundHealthChecker.
func NewCompoundHealthChecker(healthCheckers ...HealthChecker) *CompoundHealthChecker {
	return &CompoundHealthChecker{healthCheckers: healthCheckers}
}

// CheckHealth implements the HealthChecker interface.
func (c *CompoundHealthChecker) CheckHealth() error {
	for _, healthChecker := range c.healthCheckers {
		if err := healthChecker.CheckHealth(); err != nil {
			return err
		}
	}
	return nil
}
