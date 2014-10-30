package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
//	"gopkg.in/mgo.v2/bson"
	"github.com/martini-contrib/render"
	"math/rand"
	"time"
	wsh "ws_helpers"
	"net/http"
	"log"
	"code.google.com/p/go-uuid/uuid"
)


type Message struct {
	Type    string
	Message string
}


func WSHandler(w http.ResponseWriter, r *http.Request, db *mgo.Database) {

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	ip:= ws.RemoteAddr()
	message := new(Message)
	client := wsh.ClientConn{ws, ip, uuid.New(), 0, message}
	wsh.AddClient(client)
	
	defer wsh.BroadcastMessage(Message{"disconnect", client.Id})
	defer wsh.DeleteClient(client)
	
	for {
		log.Println(len(wsh.ActiveClients), wsh.ActiveClients)
		messageType, msg, err := ws.ReadMessage()
		client.MessageType = messageType
		if err != nil {
			log.Println("Disconnected", client.Id)
			log.Println(err)
			return
		}
		err = json.Unmarshal(msg, &message)
		client.LastMessage = message
		
		switch message.Type {
		case "connect":
			client.SendMessage(Message{"connect", client.Id})
			wsh.BroadcastMessage(Message{"new", client.Id})
			for c, _ := range wsh.ActiveClients{
				if c.Id != client.Id {
					client.SendMessage(Message{"new", c.Id})
				}
			}
		default:
			client.SendError("Unknown command")
		}
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("Started")
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Extensions: []string{".tmpl", ".html"},
	}))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	db := session.DB("duel")
	m.Map(db)

	m.Get("/sock", WSHandler)
	m.Get("/", func(r render.Render) {
			r.HTML(200, "index", nil)
		})

	m.Run()
}

