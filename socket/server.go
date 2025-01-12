package socket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
	sttg "github.com/jamsi-max/arbitrage/settings"
)

type WSConn struct {
	*websocket.Conn
	Topic   string
	Symbols []string
}

type Message struct {
	Type    string   `json:"type"`
	Topic   string   `json:"topic"`
	Symbols []string `json:"symbols"`
}

type MessageSpreads struct {
	Symbol  string                  `json:"symbol"`
	Spreads []orderbook.CrossSpread `json:"spreads"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	crossSpreadch chan map[string][]orderbook.CrossSpread
	lock          sync.RWMutex
	conns         map[string]map[*WSConn]bool
}

func NewServer(crossSpreadch chan map[string][]orderbook.CrossSpread) *Server {
	s := &Server{
		crossSpreadch: crossSpreadch,
		conns:         make(map[string]map[*WSConn]bool),
	}
	for symbol := range sttg.ServerSymbols {
		s.conns[symbol] = map[*WSConn]bool{}
	}
	return s
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleWS)
	go s.writeLoop()
	return http.ListenAndServe(":4000", nil)
}

func (s *Server) unregisterConn(ws *WSConn) {
	s.lock.Lock()
	for _, symbol := range ws.Symbols {
		delete(s.conns[symbol], ws)
	}
	s.lock.Unlock()

	fmt.Printf("unregistered connection %s\n", ws.RemoteAddr())

	ws.Close()
}

func (s *Server) registerConn(ws *WSConn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, symbol := range ws.Symbols {
		if _, ok := sttg.ServerSymbols[symbol]; ok {
			s.conns[symbol][ws] = true
			fmt.Printf("registered connection to symbol %s %s\n", symbol, ws.RemoteAddr())
		}
	}
}

func (s *Server) writeLoop() {
	for data := range s.crossSpreadch {

		for symbol, spreads := range data {
			// fmt.Printf("%+v %+v\n", symbol, spreads) // <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
			for ws := range s.conns[symbol] {
				msg := MessageSpreads{
					Symbol:  symbol,
					Spreads: spreads,
				}
				if err := ws.WriteJSON(msg); err != nil {
					fmt.Println("socket write error:", err)
					s.unregisterConn(ws)
				}
			}
		}
	}
}

func (s *Server) readLoop(ws *websocket.Conn) {
	defer ws.Close()

	for {
		msg := Message{}
		if err := ws.ReadJSON(&msg); err != nil {
			fmt.Println("socket read error:", err)
			break
		}
		if err := s.handleSocketMessage(ws, msg); err != nil {
			fmt.Println("handle msg error:", err)
			break
		}
	}
}

func (s *Server) handleSocketMessage(ws *websocket.Conn, msg Message) error {
	// wsConn := &WSConn{}
	if msg.Topic == "" || msg.Symbols == nil {
		return fmt.Errorf("topic: %v; symbols: %v", msg.Topic, msg.Symbols)
	}
	s.registerConn(&WSConn{ws, msg.Topic, msg.Symbols})
	return nil
}

// func (s *Server) handleSocketMessage(ws *websocket.Conn, msg Message) error {
// 	for _, symbol := range msg.Symbols {
// 		s.registerConn(symbol, ws)
// 	}
// 	return nil
// }

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
	}

	ws.WriteJSON(map[string]string{"version": "0.1"})

	go s.readLoop(ws)

}

func (s *Server) handleBestSpreads(w http.ResponseWriter, r *http.Request) {
	// ws, err := upgrader.Upgrade(w, r, nil)
	// if err != nil {
	// 	log.Println("websocket upgrade error:", err)
	// }

	// s.registerConn(ws)
}
