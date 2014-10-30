package ws_helpers

import (
	"github.com/gorilla/websocket"
	"net"
	"sync"
	"encoding/json"
	"log"
)

var ActiveClients = make(map[ClientConn]int)
var ActiveClientsRWMutex sync.RWMutex

type ClientConn struct {
	Websocket *websocket.Conn
	ClientIP  net.Addr
	Id        string
	MessageType int
	LastMessage interface {}
}

func AddClient(cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	ActiveClients[cc] = 0
	ActiveClientsRWMutex.Unlock()
}

func DeleteClient(cc ClientConn) {
	ActiveClientsRWMutex.Lock()
	delete(ActiveClients, cc)
	ActiveClientsRWMutex.Unlock()
}

func BroadcastMessage(msg interface {}) {
	ActiveClientsRWMutex.RLock()
	defer ActiveClientsRWMutex.RUnlock()
	ret, _ := json.Marshal(msg)

	for client, _ := range ActiveClients {
		if err := client.Websocket.WriteMessage(1, ret); err != nil {
			log.Fatalln(client.Id, err)
			return
		}
	}
}

func (client *ClientConn) SendMessage(msg interface {}){
	ret, _ := json.Marshal(msg)
	if err := client.Websocket.WriteMessage(client.MessageType, ret); err != nil {
		return
	}
}

func (client *ClientConn) SendError(error string) {
	client.SendMessage(struct {
		Type    string
		Message string
	}{"error", error})
}
