package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nex-server/internal/auth"
	"nex-server/internal/system"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Manager       *Manager
	Conn          *websocket.Conn
	Send          chan []byte
	Expiry        time.Time
	Authenticated bool
}

type Manager struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Media      *system.MediaController
}

func NewManager() *Manager {
	return &Manager{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Media:      system.NewMediaController(),
	}
}

func (m *Manager) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.Register:
			m.Clients[client] = true
		case client := <-m.Unregister:
			if _, ok := m.Clients[client]; ok {
				client.Conn.Close()
				delete(m.Clients, client)
				close(client.Send)
			}
		case <-ticker.C:
			m.broadcastStats()
			m.checkExpiry()
		}
	}
}

func (m *Manager) broadcastStats() {
	stats, err := system.GetSystemStats(m.Media)
	if err != nil {
		return
	}
	
	msg, _ := json.Marshal(stats)
	
	for client := range m.Clients {
		if !client.Authenticated {
			continue
		}
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(m.Clients, client)
		}
	}
}

func (m *Manager) checkExpiry() {
	now := time.Now()
	for client := range m.Clients {
		timeLeft := client.Expiry.Sub(now)
		
		if timeLeft <= 0 {
			client.Conn.WriteControl(websocket.CloseMessage, 
				websocket.FormatCloseMessage(4004, "Token expired"), 
				time.Now().Add(time.Second))
			m.Unregister <- client
			continue
		}

		if timeLeft < 4*time.Minute && timeLeft > 3*time.Minute+50*time.Second {
			evt := map[string]interface{}{
				"event": "session expiring ",
				"args":  []interface{}{fmt.Sprintf("[%s]: Your Session will expire", time.Now().Format("15:04:05"))},
			}
			data, _ := json.Marshal(evt)
			client.Send <- data
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
	}()
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			break
		}
		
		var msg struct {
			Event string   `json:"event"`
			Args  []string `json:"args"`
		}
		
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.Event == "auth" && len(msg.Args) > 0 {
			claims, err := auth.ValidateToken(msg.Args[0])
			if err != nil {
				c.Conn.WriteControl(websocket.CloseMessage, 
					websocket.FormatCloseMessage(4001, "Authentication failed"), 
					time.Now().Add(time.Second))
				return
			}
			if claims.Type != "websocket" {
				c.Conn.WriteControl(websocket.CloseMessage, 
					websocket.FormatCloseMessage(4001, "Invalid token type"), 
					time.Now().Add(time.Second))
				return
			}
			c.Authenticated = true
		}

		if !c.Authenticated {
			continue
		}

		if msg.Event == "media" && len(msg.Args) > 0 {
			switch msg.Args[0] {
			case "play_pause":
				c.Manager.Media.PlayPause()
			case "next":
				c.Manager.Media.Next()
			case "previous":
				c.Manager.Media.Previous()
			case "set_position":
				if len(msg.Args) > 1 {
					var pos int64
					if _, err := fmt.Sscanf(msg.Args[1], "%d", &pos); err == nil {
						c.Manager.Media.SetPosition(pos)
					}
				}
			}
			c.Manager.broadcastStats()
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func ServeWS(manager *Manager, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	
	client := &Client{
		Manager:       manager, 
		Conn:          conn, 
		Send:          make(chan []byte, 256),
		Expiry:        time.Now().Add(20 * time.Minute),
		Authenticated: false,
	}
	
	client.Manager.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
