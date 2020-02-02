package grpc

import "google.golang.org/grpc"

// Collect the grpc.ClientConn instances
type ClientConn struct{
	conns []*grpc.ClientConn
}


type ClientState struct{
	stats []string
}


func NewClientConn() *ClientConn {
	connClose := &ClientConn{}
	return connClose
}


// clientConn := NewClientConn()
// client := grpc.NewClient(clientOptions)
// ctx := context.Background()
// conn := client.DialContext(ctx, ":50051")
// clientConn.CollectConn(conn)
func (this *ClientConn) CollectConn(conn *grpc.ClientConn)  {
	this.conns = append(this.conns, conn)
}


//Close unity closes the grpc.ClientConn instances
func (this *ClientConn) Close()  {
	for _, conn := range this.conns {
		conn.Close()
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