package transports

type Transports interface {
	//启动服务
	Start()

	//优雅关闭服务
	GracefulShutDown()
}
