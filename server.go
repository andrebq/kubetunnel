package kubetunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type (
	Server struct {
		localListener net.Listener

		proxyConn      chan net.Conn
		listenerClosed chan struct{}
	}
)

var (
	upgrader = websocket.Upgrader{}
)

func NewServer() *Server {
	return &Server{
		proxyConn:      make(chan net.Conn, 1),
		listenerClosed: make(chan struct{}),
	}
}

func (s *Server) Run(ctx context.Context, websocketBind, targetBind string) error {
	select {
	case <-s.listenerClosed:
		return errors.New("closed")
	default:
	}
	lst, err := net.Listen("tcp", targetBind)
	if err != nil {
		return err
	}
	defer lst.Close()
	go func() {
		select {
		case <-ctx.Done():
			lst.Close()
		case <-s.listenerClosed:
			return
		}
	}()
	go func() {
		defer close(s.listenerClosed)
		timeout := time.NewTimer(time.Minute)
		for {
			conn, err := lst.Accept()
			if err != nil {
				log.Error().Err(err).Msg("Listener unable to accept new connections")
			}
			timeout.Reset(time.Minute)
			select {
			case s.proxyConn <- conn:
			case <-timeout.C:
				conn.Close()
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", s.handleWebsocket)
	mux.HandleFunc("/", s.handleIndex)

	httpServer := &http.Server{
		Addr:              websocketBind,
		ReadTimeout:       time.Minute * 10,
		WriteTimeout:      time.Minute * 10,
		ReadHeaderTimeout: time.Minute * 10,
		IdleTimeout:       time.Minute * 30,
		MaxHeaderBytes:    1_000_000,
		Handler:           mux,
	}

	err = httpServer.ListenAndServe()
	return err
}

func (s *Server) handleIndex(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Welcome to kubetunnel")
}

func (s *Server) handleWebsocket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error().Err(err).Msg("Upgrade failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()
	var proxy net.Conn
	select {
	case proxy = <-s.proxyConn:
	case <-s.listenerClosed:
		return
	}
	log := log.With().Str("internal-client", proxy.RemoteAddr().String()).Str("websocket-client", conn.RemoteAddr().String()).Logger()
	log.Info().Msg("Starting websocket tunnel")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer conn.Close()
		var buf [4069]byte
		for {
			n, err := proxy.Read(buf[:])
			if err != nil {
				log.Error().Err(err).Msg("Error reading data from internal client")
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Error().Err(err).Msg("Error sending data to websocket client")
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		defer proxy.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Error().Err(err).Msg("Error reading data from websocket client")
				return
			}
			_, err = proxy.Write(msg)
			if err != nil {
				log.Error().Err(err).Msg("Error writing data to internal client")
				return
			}
		}
	}()
	wg.Wait()
}
