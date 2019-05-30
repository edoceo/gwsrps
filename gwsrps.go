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
import "github.com/oklog/ulid"

type WS_Client_List struct {
	client_list map[*WS_Client]bool
}

type WS_Client struct {
	id string
	ws *websocket.Conn
	pump chan []byte
}

func generateULID() string {
	t := time.Unix(1000000, 0)
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	u, _ := ulid.New(ulid.Timestamp(t), entropy)
	return u.String()
	// Output: 0000XSNJG0MQJHBF4QX1EFD6Y3
}


/**
 * Read Message from WebSocket, Publish to Redis
 * @param {[type]} wsc [description]
 * @return {[type]} [description]
 */
func (wsc *WS_Client) readPump() {

}

/**
 * Send Message to WebSocket
 * @param {[type]} wsc [description]
 * @return {[type]} [description]
 */
func (wsc *WS_Client) sendPump() {

}

/*
var (
    gStore      *Store
    gPubSubConn *redis.PubSubConn
    gRedisConn  = func() (redis.Conn, error) {
        return redis.Dial("tcp", ":6379")
    }
)
func init() {
    gStore = &Store{
        Users: make([]*User, 0, 1),
    }
}
*/

var ws_upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func main() {

	cmd, _ := os.Executable()

	dir, _ := filepath.Abs(filepath.Dir(cmd))
	fmt.Println(dir)

	err50 := os.Chdir(dir)
	if err50 != nil {
		panic(err50)
	}

	//cli_args := os.Args[1:]
	//fmt.Println(cli_args)

	redis := flag.String("redis", "127.0.0.1:6357", "Redis Server host:port:db")
	tcpListen := flag.String("listen", ":8080", "Listen Address for TCP")
	tlsListen := flag.String("listen-tls", ":8443", "Listen Address for TLS/SSL")
	tlsPemFile := flag.String("tls-pem-file", "server.pem", "Stuff")
	tlsKeyFile := flag.String("tls-key-file", "server.key", "Stuff")
	origin    := flag.String("origin", "*", "Only allow these origins")
	path      := flag.String("path", "/ws", "What path is the WebSocket in?")

	flag.Parse()


	fmt.Println("Connecting to Redis:", *redis)
	fmt.Println("TCP:", *tcpListen)
	fmt.Println("TLS:", *tlsListen)
	//fmt.Println("CRT:", *tlsFile)
	fmt.Println("Origin:", *origin)

	// Fork to Background


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

		cid := generateULID()
		client := &WS_Client{ id: cid, ws: ws, pump: make(chan []byte, 256) }

		fmt.Println("New Client:", cid)
		// client.hub.register <- client
		fmt.Println(client)

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.
		// go client.readPump()
		// go client.sendPump()

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

	swg.Wait()

}
