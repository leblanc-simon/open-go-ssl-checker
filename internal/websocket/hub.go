package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"leblanc.io/open-go-ssl-checker/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for simplicity.
		// In production, you should check r.Header.Get("Origin")
		// against a list of allowed origins.
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients          map[*Client]bool
	broadcast        chan []byte // Messages to send (serialized JSON data)
	register         chan *Client
	unregister       chan *Client
	store            *store.Store // To retrieve updated data
	mu               sync.Mutex   // To protect access to `clients`
	refreshRequested chan struct{}
}

func NewHub(s *store.Store) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan []byte),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		store:            s,
		refreshRequested: make(chan struct{}, 1),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true

			log.Println("WebSocket client registered")
			h.mu.Unlock()
			// Send current state on initial connection
			h.sendCurrentState(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("WebSocket client unregistered")
			}
			h.mu.Unlock()

		case message := <-h.broadcast: // This channel will be used by NotifyUpdate
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default: // Do not block if the client's buffer is full
					close(client.send)
					delete(h.clients, client)
					log.Println("WebSocket client unregistered (buffer full or error)")
				}
			}
			h.mu.Unlock()
		}
	}
}

// NotifyUpdate is called when data has changed and needs to be broadcast.
func (h *Hub) NotifyUpdate() {
	log.Println("Update notification received, preparing WebSocket broadcast...")

	summaries, err := h.store.GetLatestChecksSummary()
	if err != nil {
		log.Printf("WebSocket Hub error - GetLatestChecksSummary: %v", err)

		return
	}

	jsonData, err := json.Marshal(summaries)
	if err != nil {
		log.Printf("WebSocket Hub error - Marshal summaries: %v", err)

		return
	}
	h.broadcast <- jsonData // Send JSON data to broadcast channel
}

// RefreshRequests returns a read-only channel that emits a signal
// when any connected client requests a full refresh via the WebSocket.
func (h *Hub) RefreshRequests() <-chan struct{} {
	return h.refreshRequested
}

// ServeWs handles client WebSocket requests.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade error: %v", err)

		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256)}
	h.register <- client

	// Allow receiving messages from the client (if necessary)
	// and handle connection closure.
	go client.writePump(h) // Pass the hub to allow unregistering
	go client.readPump(h)  // Pass the hub to allow unregistering
}

// sendCurrentState sends the current state of data to the specified client.
func (h *Hub) sendCurrentState(client *Client) {
	summaries, err := h.store.GetLatestChecksSummary()
	if err != nil {
		log.Printf("WebSocket Hub error - sendCurrentState - GetLatestChecksSummary: %v", err)

		return
	}

	jsonData, err := json.Marshal(summaries)
	if err != nil {
		log.Printf("WebSocket Hub error - sendCurrentState - Marshal summaries: %v", err)

		return
	}

	// Send directly to the client's 'send' channel to avoid the initial broadcast loop
	// for a single client and ensure it receives the state even if a general broadcast is ongoing.
	select {
	case client.send <- jsonData:
	default:
		log.Printf(
			"Unable to send initial state to client %v (buffer full or closed)",
			client.conn.RemoteAddr(),
		)
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512) // Message size limit
	err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		return
	} // Read timeout
	c.conn.SetPongHandler(
		func(string) error {
			err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			if err != nil {
				return err
			}

			return nil
		},
	)

	for {
		messageType, message, err := c.conn.ReadMessage()
		// Read client messages (if your application needs it)
		// For this application, we do not expect messages from the client.
		// This loop mainly serves to detect connection closure.
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("WebSocket readPump error: %v", err)
			}

			break
		}

		err = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			log.Printf("WebSocket readPump error (SetReadDeadline): %v", err)
			return
		}

		if messageType == websocket.TextMessage {
			if string(message) == "refresh" {
				// Signal a refresh request without blocking
				select {
				case hub.refreshRequested <- struct{}{}:
					log.Println("Received refresh request from client; triggering full checks...")
				default:
					// If a refresh signal is already pending, avoid piling up
					log.Println("Refresh request already pending; ignoring duplicate.")
				}
			}
		}
	}
}

func (c *Client) writePump(hub *Hub) {
	ticker := time.NewTicker(
		45 * time.Second,
	)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				return
			}

			if !ok {
				// The hub closed the c.send channel
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					return
				}

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket writePump error: %v", err)

				return // Important: If writing fails, the client has probably gone away.
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket writePump error (ping): %v", err)

				return
			}
		}
	}
}
