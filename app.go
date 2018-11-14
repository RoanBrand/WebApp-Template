package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

func main() {
	// An example of a HTTP endpoint at the server
	// which a client (browser) can make a dynamic HTTP call to.
	// This function will parse a number from the request, increment it, and send it back as the response.
	doServerStuff := func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters from request.
		queryParameters := r.URL.Query()

		// Get the number parameter we are looking for.
		numValues, ok := queryParameters["number"]
		if !ok {
			http.Error(w, "No number provided", http.StatusBadRequest)
			return
		}

		// Convert the number from string form into integer.
		num, err := strconv.Atoi(numValues[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Increment the number
		num++

		// Print new number to console and send it back to requester.
		log.Println(num)
		w.Write([]byte(strconv.Itoa(num)))
	}
	http.HandleFunc("/changenumber", doServerStuff) // Run this function when this URL path is hit.

	// Serve static files (files to be run in browser at client html, css, js) to browser when root URL is hit.
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Setup Websockets
	hub := newHub()
	go hub.run()
	http.HandleFunc("/ws", hub.handleWS)

	// Start HTTP server on localhost port 80 and exit when it fails.
	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func handleError(w http.ResponseWriter, errMsg string) {
	log.Println(errMsg)
	http.Error(w, errMsg, http.StatusInternalServerError)
}

// Websockets
type client struct {
	conn *websocket.Conn
	send chan []byte
	hub  *hub
}

func (c *client) reader() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// ignoring message
		// we only want to increase the server copy and send it to all clients
		// if you want to send different things over ws connection, you will need to parse and handle message

		num := atomic.AddInt32(&c.hub.sharedNum, 1)
		c.hub.broadcast <- []byte(strconv.Itoa(int(num)))
	}
}

func (c *client) writer() {
	defer func() {
		c.conn.Close()
	}()
	for {
		msg, ok := <-c.send
		if !ok {
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println(err)
			return
		}
	}
}

type hub struct {
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan []byte

	sharedNum int32
}

func newHub() *hub {
	return &hub{
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte),
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{}

func (h *hub) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleError(w, err.Error())
		return
	}

	newClient := &client{conn: conn, hub: h, send: make(chan []byte, 256)}
	h.register <- newClient

	go newClient.writer()
	go newClient.reader()
	newClient.send <- []byte(strconv.Itoa(int(h.sharedNum))) // must get initial value
}
