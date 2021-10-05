package wsbus

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type (
	bus struct {
		sync.RWMutex
		conns map[*websocket.Conn]struct{}
	}
)

func (b *bus) propagate(source *websocket.Conn, msgType int, data []byte) {
	b.RLock()
	defer b.RUnlock()
	for k := range b.conns {
		if k == source {
			continue
		}
		err := k.WriteMessage(msgType, data)
		if err != nil {
			log.Error().Err(err).Str("failed-client", k.RemoteAddr().String()).Msg("Unable to send message to client")
			k.Close()
			delete(b.conns, k)
		}
	}
}

func (b *bus) register(conn *websocket.Conn) {
	b.Lock()
	defer b.Unlock()
	if b.conns == nil {
		b.conns = make(map[*websocket.Conn]struct{})
	}
	b.conns[conn] = struct{}{}
}

func (b *bus) unregister(conn *websocket.Conn) {
	b.Lock()
	defer b.Unlock()
	defer conn.Close()
	if b.conns == nil {
		return
	}
	delete(b.conns, conn)
}

func Run(bind, directory string) error {
	up := websocket.Upgrader{}
	bus := &bus{}
	http.HandleFunc("/bus", func(w http.ResponseWriter, req *http.Request) {
		conn, err := up.Upgrade(w, req, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		bus.register(conn)
		defer bus.unregister(conn)
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				bus.unregister(conn)
				log.Error().Err(err).Str("failed-client", conn.RemoteAddr().String()).Msg("Unable to read message from client")
				return
			}
			bus.propagate(conn, mt, data)
		}
	})
	http.Handle("/", http.FileServer(http.Dir(directory)))

	return http.ListenAndServe(bind, nil)
}
