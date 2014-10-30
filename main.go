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
	"ws_helpers"
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
	client := ws_helpers.ClientConn{ws, ip, uuid.New(), 0, message}
	ws_helpers.AddClient(client)
	
	defer ws_helpers.BroadcastMessage(struct {
		Type    string
		Message string
	}{"disconnect", client.Id})
	defer ws_helpers.DeleteClient(client)
	
	for {
		log.Println(len(ws_helpers.ActiveClients), ws_helpers.ActiveClients)
		messageType, msg, err := ws.ReadMessage()
		client.MessageType = messageType
		log.Println(messageType)
		if err != nil {
			log.Println("bye")
			log.Println(err)
			return
		}
		err = json.Unmarshal(msg, &message)
		client.LastMessage = message
		
		switch message.Type {
		case "connect":
			client.SendMessage(struct {
				Type    string
				Message string
			}{"connect", client.Id})
			ws_helpers.BroadcastMessage(struct {
					Type    string
					Message string
			}{"new", client.Id})
			for c, _ := range ws_helpers.ActiveClients{
				client.SendMessage(struct {
						Type    string
						Message string
					}{"new", c.Id})
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

