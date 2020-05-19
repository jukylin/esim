package factory

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

const (
	testStructName = "Test"
)

// Model is the interface that we're exposing as a plugin.
type Model interface {
	Sort() string

	InitField() string
}

// Here is an implementation that talks over RPC
type ModelRPC struct{ client *rpc.Client }

func (g *ModelRPC) Sort() string {
	var resp string
	err := g.client.Call("Plugin.Sort", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (g *ModelRPC) InitField() string {
	var resp string
	err := g.client.Call("Plugin.InitField", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

// Here is the RPC server that ModelRPC talks to, conforming to
// the requirements of net/rpc
type ModelRPCServer struct {
	// This is the real implementation
	Impl Model
}

func (s *ModelRPCServer) Sort(args interface{}, resp *string) error {
	*resp = s.Impl.Sort()
	return nil
}

func (s *ModelRPCServer) InitField(args interface{}, resp *string) error {
	*resp = s.Impl.InitField()
	return nil
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a ModelRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return ModelRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type ModelPlugin struct {
	// Impl Injection
	Impl Model
}

func (p *ModelPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ModelRPCServer{Impl: p.Impl}, nil
}

func (ModelPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ModelRPC{client: c}, nil
}
