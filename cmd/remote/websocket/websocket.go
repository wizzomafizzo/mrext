package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func send(c *websocket.Conn, msg string) error {
	if c == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	err := c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

type connGroup struct {
	mu    sync.Mutex
	conns []*websocket.Conn
}

func (cg *connGroup) Add(c *websocket.Conn) int {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.conns = append(cg.conns, c)
	return len(cg.conns) - 1
}

func (cg *connGroup) Remove(i int) {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.conns = append(cg.conns[:i], cg.conns[i+1:]...)
}

func (cg *connGroup) All() []*websocket.Conn {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	return cg.conns
}

func (cg *connGroup) Clean() {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	fresh := make([]*websocket.Conn, 0)
	for _, c := range cg.conns {
		if c != nil {
			fresh = append(fresh, c)
		}
	}
	cg.conns = fresh
}

func (cg *connGroup) Broadcast(logger *service.Logger, msg string) {
	cg.Clean()
	cg.mu.Lock()
	defer cg.mu.Unlock()
	for _, c := range cg.conns {
		err := send(c, msg)
		if err != nil {
			logger.Error("failed to write to websocket: %s", err)
		}
	}
}

var conns = &connGroup{}

func Handle(
	logger *service.Logger,
	connectPayload func() []string,
	msgHandler func(msg string) string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("failed to upgrade websocket: %s", err)
			return
		}

		id := conns.Add(c)

		defer func(c *websocket.Conn) {
			err := c.Close()
			if err != nil {
				logger.Error("failed to close websocket: %s", err)
			}
			conns.Remove(id)
		}(c)

		for _, msg := range connectPayload() {
			err = send(c, msg)
			if err != nil {
				logger.Error("failed to write to websocket during connect: %s", err)
				return
			}
		}

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					return
				} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				logger.Error("failed to read from websocket: %s", err)
				return
			}

			logger.Info("received message: %s", msg)
			response := msgHandler(string(msg))

			if response == "" {
				continue
			}

			err = send(c, response)
			if err != nil {
				logger.Error("failed to write to websocket: %s", err)
				return
			}
		}
	}
}

func Broadcast(logger *service.Logger, msg string) {
	conns.Broadcast(logger, msg)
}
