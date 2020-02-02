package grpc

import (
	"google.golang.org/grpc"
	"github.com/jukylin/esim/log"
)

// Collect the grpc.ClientConn instances
type ClientConn struct{
	conns []*grpc.ClientConn

	logger log.Logger
}


type ClientState struct{
	stats []string
}


func NewClientConn(logger log.Logger) *ClientConn {
	connClose := &ClientConn{
		logger:logger,
	}
	return connClose
}


// Examples：
// 		client := grpc.NewClient(clientOptions)
// 		ctx := context.Background()
// 		conn := client.DialContext(ctx, ":50051")
//
// 		clientConn := NewClientConn()
// 		clientConn.CollectConn(conn)
func (this *ClientConn) CollectConn(conn *grpc.ClientConn)  {
	this.conns = append(this.conns, conn)
}


//Close unity closes the grpc.ClientConn instances
func (this *ClientConn) Close()  {
	var err error
	for _, conn := range this.conns {
		err = conn.Close()
		if err != nil {
			this.logger.Errorf("%s colse err : %s", conn.Target(), err.Error())
		}
	}
}

//State unity show the grpc.ClientConn state
func (this *ClientConn) State() []string {
	var state []string
	for _, conn := range this.conns {
		state = append(state, conn.Target() + ":" + conn.GetState().String())
	}

	return state
}