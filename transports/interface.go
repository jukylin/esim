package transports

type Transports interface {
	// start server
	Start()

	// graceful shutdown server
	GracefulShutDown()
}
