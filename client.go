package kubetunnel

import (
	"context"
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type (
	Client struct {
	}
)

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Run(ctx context.Context, serverEndpoint, localEndpoint string) error {
	openConnections := make(map[*websocket.Conn]struct{})
	defer func() {
		for conn := range openConnections {
			conn.Close()
		}
	}()
	popConn := make(chan *websocket.Conn, 10)
LOOP:
	for {
		popComplete := false
		for !popComplete {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case c := <-popConn:
				c.Close()
				delete(openConnections, c)
			default:
				popComplete = true
			}
		}
		conn, resp, err := websocket.DefaultDialer.DialContext(ctx, serverEndpoint, nil)
		if err != nil {
			return err
		}
		// TODO: properly handle a failed handshake
		_ = resp
		// TODO: by waiting until the loop exists, we are effectively creating a memory leak
		// deal with this eventually
		openConnections[conn] = struct{}{}
		firstPacket := make(chan struct{})
		go c.proxyConn(ctx, conn, popConn, localEndpoint, firstPacket)
		select {
		case <-firstPacket:
			continue
		case <-ctx.Done():
			break LOOP
		}
	}
	return ctx.Err()
}

func (c *Client) proxyConn(ctx context.Context, conn *websocket.Conn, done chan<- *websocket.Conn, localEndpoint string, firstPacket chan<- struct{}) {
	log := log.With().Str("local-endpoint", localEndpoint).Str("websocket-local", conn.LocalAddr().String()).Logger()
	defer func() { done <- conn }()
	defer func() {
		close(firstPacket)
	}()

	// wait for the first packet from the remote client
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Error().Err(err).Msg("Unable to obtain the first message from the websocket connection")
		return
	}

	localConn, err := net.Dial("tcp", localEndpoint)
	if err != nil {
		log.Error().Err(err).Msg("Unable to open connection to local endpoint")
		return
	}
	defer localConn.Close()

	_, err = localConn.Write(msg)
	if err != nil {
		log.Error().Err(err).Msg("Unable to send the first packet to the local endpoint")
		return
	}
	firstPacket <- struct{}{}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer conn.Close()
		var buf [4096]byte
		for {
			n, err := localConn.Read(buf[:])
			if err != nil {
				log.Error().Err(err).Msg("Unable to read data from local connection")
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Error().Err(err).Msg("Unable to send data to websocket connection")
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer localConn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Error().Err(err).Msg("Unable to read message from websocket connection")
				return
			}
			_, err = localConn.Write(msg)
			if err != nil {
				log.Error().Err(err).Msg("Unable to write message to local connection")
				return
			}
		}
	}()

	wg.Wait()
}
