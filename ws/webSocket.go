//webSocket.go
package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	// Inbound messages from the clients.
	broadcast chan []byte
	//map with users
	users map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		broadcast: make(chan []byte),

		users: make(map[string]*websocket.Conn),
	}
}

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	userId string

	username string
}

func (c *Client) ClientRead() {

	//fmt.Println("ClientRead()")
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		//fmt.Println(err, "error ocured with ClientRead()")
		return
	}
	//fmt.Println("ClientRead() message sent1")

	select {
	case c.send <- message:
		//fmt.Println("ClientRead() message sent2")
	}

	c.ClientRead()
}

type Msg struct {
	Msg string
	Id  string
}

func (c *Client) ClientWrite() {

	//fmt.Println("ClientWrite()1")
	for {
		select {
		case message := <-c.send:

			msg := &Msg{}
			err := json.Unmarshal(message, msg)
			if err != nil {
				log.Println(err)
			}
			//fmt.Println("ClientWrite() select", msg.Id)
			user, ok := c.hub.users[msg.Id]
			// log.Println("heloo")
			if !ok {
				fmt.Printf("error with key msg: %v", user)
				return
			}

			jsonMap := make(map[string]string)
			jsonMap["msg"] = msg.Msg
			jsonMap["id"] = c.userId
			jsonMap["username"] = c.username
			jsonMsg, err := json.Marshal(jsonMap)
			if err != nil {
				fmt.Println(err)
				return
			}

			if err := user.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
				fmt.Printf("error with write msg: %v", err)
				return
			}
		}
	}
}

func (h *Hub) HubRun() {
	fmt.Println("hubRun()")
	// select {
	// case msg := <-h.broadcast:
	// }
}

func ServeWS(w http.ResponseWriter, r *http.Request, hub *Hub, id string, username string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 1024),
		userId:   id,
		username: username,
	}

	//заносим в карту текущее подключение с ключем = id пользователя
	hub.users[id] = client.conn
	//fmt.Println("id is", id)

	go client.ClientRead()
	go client.ClientWrite()

	//fmt.Println("serveWS")

}
