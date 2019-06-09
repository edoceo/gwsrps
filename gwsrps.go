/**
 * Go/WebSocket + Redis/PubSub
 */

package main

import "os"
import "flag"
import "fmt"
import "strings"
import "time"
import "math/rand"
import "net/http"
import "path/filepath"
import "sync"

import "github.com/gorilla/websocket"
import "github.com/go-redis/redis"
import "github.com/oklog/ulid"


// I like these as IDs
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


/**
 * Main
 */
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
	//path      := flag.String("path", "/ws", "What path is the WebSocket in?")

	flag.Parse()


	fmt.Println("Connecting to Redis:", *redisConn)
	fmt.Println("TCP:", *tcpListen)
	fmt.Println("TLS:", *tlsListen)
	//fmt.Println("CRT:", *tlsFile)
	fmt.Println("Origin:", *origin)

	// Static Files
	// http.Handle("/", http.FileServer(http.Dir(dir)))


	// HTTP Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		is_ws := false
		for _, header := range r.Header["Upgrade"] {
			if header == "websocket" {
				is_ws = true
				break
			}
		}

		if (!is_ws) {
			// Serve Static File
			http.ServeFile(w, r, "index.html")
			return
		}

		// Verify Origin
		ws_upgrader.CheckOrigin = func(r *http.Request) bool {
			// Filepath Match? Should we use RegEx?
			good, _ := filepath.Match(*origin, r.Host)
			return good
		}

		// Upgrade
		ws, err := ws_upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		// defer ws.Close()


		// Check *path and use that as the subscribe channel
		path := strings.Trim(r.URL.Path, "/");
		fmt.Println("Requst Page", path)


		// Start our Client Session Object
		client := &WS_Client{
			id: generateULID(),
			ws: ws,
			path: path,
			stat: "live",
		}

		client_list[ client ] = time.Now().Unix()

		// Connect to Redis
		rdbc := redis.NewClient(&redis.Options{
			Addr: *redisConn,
			Password: "", // no password set
			DB: 0,  // use default DB
		})

		client.pub = rdbc

		client.sub = rdbc.Subscribe("gwsrps")

		// Read the PubSub and Forward to Pump
		go client.pumpWebSocketToRedis();
		go client.pumpRedisToWebSocket()

		fmt.Println("New Client:", client.id, "Clients:", len(client_list))

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
