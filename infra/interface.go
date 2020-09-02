package infra

//
type Infra interface {
	// close resource when app stop
	Close()

	HealthCheck() []error
}
