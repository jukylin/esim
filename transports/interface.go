package transports

type Transports interface {

	Start()

	GracefulShutDown()
}
