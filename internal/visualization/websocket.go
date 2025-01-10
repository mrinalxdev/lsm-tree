package visualization

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mrinalxdev/lsm-tree/internal/store"
)

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    lsm        *store.LSMTree
    mu         sync.Mutex
}

type Message struct {
    Type      string    `json:"type"`
    Key       string    `json:"key"`
    Value     string    `json:"value,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

func NewHub(lsm *store.LSMTree) *Hub {
    return &Hub{
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        clients:    make(map[*Client]bool),
        lsm:        lsm,
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.Lock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.Unlock()
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }

        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("error unmarshaling message: %v", err)
            continue
        }

        // Handle different message types
        switch msg.Type {
        case "set":
            if err := c.hub.lsm.Set(msg.Key, msg.Value); err != nil {
                log.Printf("error setting value: %v", err)
            }
        case "get":
            value, err := c.hub.lsm.Get(msg.Key)
            if err != nil {
                log.Printf("error getting value: %v", err)
                continue
            }
            msg.Value = value
        case "delete":
            if err := c.hub.lsm.Delete(msg.Key); err != nil {
                log.Printf("error deleting value: %v", err)
            }
        }

        // Broadcast the message to all clients
        response, err := json.Marshal(msg)
        if err != nil {
            log.Printf("error marshaling response: %v", err)
            continue
        }
        c.hub.broadcast <- response
    }
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            if err := w.Close(); err != nil {
                return
            }
        }
    }
}

func ServeWs(hub *Hub, conn *websocket.Conn) {
    client := &Client{
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 256),
    }
    client.hub.register <- client

    go client.writePump()
    go client.readPump()
}
