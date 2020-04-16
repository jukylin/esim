package transports

type Transports interface {
	//start server
	Start()

	//register server or router
	Register()

	//graceful shutdown server
	GracefulShutDown()
}
