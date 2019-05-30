/**
 * Go/WebSocket + Redis/PubSub
 */

package main

import "os"
import "flag"
import "fmt"
import "time"
import "math/rand"
import "net/http"
import "path/filepath"
import "sync"

import "github.com/gorilla/websocket"
import "github.com/go-redis/redis"
import "github.com/oklog/ulid"

// type WS_Client_List struct {
// 	client_list map[*WS_Client]int
// }

type WS_Client struct {
	id string
	ws *websocket.Conn
	pub *redis.Client
	sub *redis.PubSub
	pump chan []byte
}

/**
 * Read Message from WebSocket, Publish to Redis
 * @param {[type]} wsc [description]
 * @return {[type]} [description]
 */
func (c *WS_Client) incomingPump() {

	defer func() {
		// c.hub.unregister <- c
		fmt.Println("incomingPump - defer")
		c.ws.Close()
		// c.pub.Close()
		c.sub.Close()
	}()

	pongWait := 60 * time.Second

	c.ws.SetReadLimit(32768)

	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		msgType, msgBody, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}

		switch (msgType) {
		case websocket.TextMessage:
			fmt.Println("incoming-ws:", string(msgBody))
			// fmt.Println("Text Message")
			break;
		case websocket.BinaryMessage:
			fmt.Println("Data Message")
			break;
		}

		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// c.hub.broadcast <- message
		c.pub.Publish("gwsrps", msgBody)
	}

}

/**
 * Send Message to WebSocket
 * @param {[type]} wsc [description]
 * @return {[type]} [description]
 */
func (c *WS_Client) outgoingPump() {

	sendWait := 8 * time.Second
	pingWait := 60 * time.Second * 9 / 10

	tick := time.NewTicker(pingWait)

	defer func() {
		fmt.Println("outgoingPump - defer")
		// c.hub.unregister <- c
		tick.Stop()
		c.ws.Close()
		//c.pub.Close()
		c.sub.Close()
	}()


	for {
		fmt.Println("Line 098")
		select {
		case msgBody, ok := <- c.pump:

			fmt.Println("outgoing-gws:", string(msgBody))

			c.ws.SetWriteDeadline(time.Now().Add(sendWait))

			if !ok {
				// The hub closed the channel.
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msgBody)

			// // Add queued chat messages to the current websocket message.
			// n := len(c.pump)
			// for i := 0; i < n; i++ {
			// 	w.Write(newline)
			// 	w.Write(<-c.pump)
			// }

			// Close Error, Kill It
			err112 := w.Close()
			if err112 != nil {
				return
			}

			break

		case <-tick.C:

			c.ws.SetWriteDeadline(time.Now().Add(sendWait))
			err127 := c.ws.WriteMessage(websocket.PingMessage, nil)
			if err127 != nil {
				return
			}

		}
	}

}

// var gPubSubConn *redis.PubSubConn
//     gPubSubConn = &redis.PubSubConn{Conn: conn}

func generateULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	u, _ := ulid.New(ulid.Timestamp(t), entropy)
	return u.String()
	// Output: 0000XSNJG0MQJHBF4QX1EFD6Y3
}


var ws_upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var client_list = make(map[*WS_Client]int64)


func main() {

	cmd, _ := os.Executable()

	dir, _ := filepath.Abs(filepath.Dir(cmd))
	fmt.Println(dir)

	err50 := os.Chdir(dir)
	if err50 != nil {
		panic(err50)
	}


	// Read Options
	redisConn := flag.String("redis", "127.0.0.1:6379", "Redis Server host:port:db")
	tcpListen := flag.String("listen", ":8080", "Listen Address for TCP")
	tlsListen := flag.String("listen-tls", ":8443", "Listen Address for TLS/SSL")
	tlsPemFile := flag.String("tls-pem-file", "server.pem", "Stuff")
	tlsKeyFile := flag.String("tls-key-file", "server.key", "Stuff")
	origin    := flag.String("origin", "*", "Only allow these origins")
	path      := flag.String("path", "/ws", "What path is the WebSocket in?")

	flag.Parse()


	fmt.Println("Connecting to Redis:", *redisConn)
	fmt.Println("TCP:", *tcpListen)
	fmt.Println("TLS:", *tlsListen)
	//fmt.Println("CRT:", *tlsFile)
	fmt.Println("Origin:", *origin)

	// Static Files
	http.Handle("/", http.FileServer(http.Dir(dir)))


	// HTTP Handler
	http.HandleFunc(*path, func(w http.ResponseWriter, r *http.Request) {

		// fmt.Println("HTTP Connection, Should Upgrade")

		ws_upgrader.CheckOrigin = func(r *http.Request) bool {
			good, _ := filepath.Match(*origin, r.Host)
			return good
		}

		ws, err := ws_upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		// defer ws.Close()

		cid := generateULID()
		client := &WS_Client{ id: cid, ws: ws }

		client_list[ client ] = time.Now().Unix()

		// client.hub.register <- client
		// fmt.Println(client)

		// Connect to Redis
		rdbc := redis.NewClient(&redis.Options{
			Addr: *redisConn,
			Password: "", // no password set
			DB: 0,  // use default DB
		})

		client.pub = rdbc

		client.sub = rdbc.Subscribe("gwsrps")
		_, err142 := client.sub.Receive()
		if (err142 != nil) {
			panic(err142)
		}

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.

		// Read the PubSub and Forward to Pump
		go client.incomingPump()
		go client.outgoingPump()

		go func() {
			for {
				var msg, _ = client.sub.ReceiveMessage()
				fmt.Println("incoming-ps:", msg)
				client.pump <- []byte(msg.Payload)
			}
		}()

		fmt.Println("New Client:", cid, "Clients:", len(client_list))

	})

	// Start Servers in a WaitGroup

	swg := &sync.WaitGroup{}

	// TCP Listener
	swg.Add(1)
	go func() {
		err := http.ListenAndServeTLS(*tlsListen, *tlsPemFile, *tlsKeyFile, nil)
		if err != nil {
			panic(err)
		}
		swg.Done()
	}()

	// TLS Listener
	swg.Add(1)
	go func() {
		err := http.ListenAndServe(*tcpListen, nil)
		if err != nil {
			panic(err)
		}
		swg.Done()
	}()

	// Fork to Background


	swg.Wait()

}
