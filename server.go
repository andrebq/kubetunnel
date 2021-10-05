package kubetunnel

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/andrebq/kubetunnel/internal/protocol"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	Server struct {
		protocol.UnimplementedTunnelServer

		tunnels uint64

		activeTunnels map[uint64]bool
	}
)

func NewServer() *Server {
	return &Server{
		activeTunnels: make(map[uint64]bool),
	}
}

func (s *Server) Run(ctx context.Context, bind string) error {
	server := grpc.NewServer()
	protocol.RegisterTunnelServer(server, s)

	lst, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	return server.Serve(lst)

	// httpServer := http.Server{
	// 	Addr:              bind,
	// 	ReadTimeout:       time.Minute * 5,
	// 	WriteTimeout:      time.Minute * 5,
	// 	ReadHeaderTimeout: time.Minute * 5,
	// 	Handler:           server,
	// }
	// return httpServer.ListenAndServe()
}

func (s *Server) Handshake(ctx context.Context, req *protocol.HandshakeRequest) (*protocol.HandshakeResponse, error) {
	tid := atomic.AddUint64(&s.tunnels, 1)
	s.activeTunnels[tid] = true
	lst, err := net.Listen("tcp", req.GetRemoteBind())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to bind on %v, cause %v", req.GetRemoteBind(), err))
	}
	go s.openTunnel(context.Background(), lst, tid)
	return &protocol.HandshakeResponse{
		RemoteBind: req.GetRemoteBind(),
		TunnelID:   tid,
	}, nil
}

func (s *Server) Mux(stream protocol.Tunnel_MuxServer) error {
	for {
		m, err := stream.Recv()
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
		err = stream.Send(m)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
	}
}

func (s *Server) openTunnel(ctx context.Context, lst net.Listener, tid uint64) {
	for {
		conn, err := lst.Accept()
		if err != nil {
			log.Error().Err(err).Msg("Unable to accept new connection")
			return
		}
		go s.handleConn(ctx, conn, tid)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn, tid uint64) {
	log.Info().Str("local-addr", conn.LocalAddr().String()).Str("remote-addr", conn.RemoteAddr().String()).Msg("Got new connection")
	conn.Close()
}
