package kubetunnel

import (
	"context"
	"sync"

	"github.com/andrebq/kubetunnel/internal/protocol"
	"google.golang.org/grpc"
)

type (
	Client struct {
		tunnelCli  protocol.TunnelClient
		tunnelConn *grpc.ClientConn
		tunnelID   uint64

		localEndpoint  string
		remoteEndpoint string

		inPkts  chan *protocol.Packet
		outPkts chan *protocol.Packet
	}
)

func NewClient(ctx context.Context, remoteEndpoint, localEndpoint, tunnelServer string) (*Client, error) {
	conn, err := grpc.DialContext(ctx, tunnelServer, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := &Client{
		tunnelCli:      protocol.NewTunnelClient(conn),
		tunnelConn:     conn,
		localEndpoint:  localEndpoint,
		remoteEndpoint: remoteEndpoint,

		inPkts:  make(chan *protocol.Packet, 1000),
		outPkts: make(chan *protocol.Packet, 1000),
	}

	err = c.handshake(ctx)
	if err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

func (c *Client) Run(ctx context.Context) error {
	cli, err := c.tunnelCli.Mux(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case v := <-c.outPkts:
				err := cli.Send(v)
				if err != nil {
					return
				}
			case <-ctx.Done():
				cli.CloseSend()
			}
		}
	}()
	go func() {
		defer wg.Done()
		for {
			p, err := cli.Recv()
			if err != nil {
				return
			}
			select {
			case c.inPkts <- p:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()

	return ctx.Err()
}

func (c *Client) Close() {
	c.tunnelConn.Close()
}

func (c *Client) handshake(ctx context.Context) error {
	res, err := c.tunnelCli.Handshake(ctx, &protocol.HandshakeRequest{
		RemoteBind: c.remoteEndpoint,
	})
	if err != nil {
		return err
	}
	c.tunnelID = res.TunnelID
	return err
}
