package grpc

import (
	"github.com/jukylin/esim/log"
	"google.golang.org/grpc"
)

// Collect the grpc.ClientConn instances
type ClientConn struct {
	conns []*grpc.ClientConn

	logger log.Logger
}

type ClientState struct {
	stats []string
}

func NewClientConn(logger log.Logger) *ClientConn {
	connClose := &ClientConn{
		logger: logger,
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
func (cc *ClientConn) CollectConn(conn *grpc.ClientConn) {
	cc.conns = append(cc.conns, conn)
}

//Close unity closes the grpc.ClientConn instances
func (cc *ClientConn) Close() {
	var err error
	for _, conn := range cc.conns {
		err = conn.Close()
		if err != nil {
			cc.logger.Errorf("%s colse err : %s", conn.Target(), err.Error())
		}
	}
}

//State unity show the grpc.ClientConn state
func (cc *ClientConn) State() []string {
	var state []string
	for _, conn := range cc.conns {
		state = append(state, conn.Target()+":"+conn.GetState().String())
	}

	return state
}
