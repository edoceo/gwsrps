/**
 * My Clent Connection
 */

package main

import (
	"fmt"
	"time"
	"github.com/gorilla/websocket"
	"github.com/go-redis/redis"
)

type WS_Client struct {
	id string
	ws *websocket.Conn
	pub *redis.Client
	sub *redis.PubSub
	path string // Path in URL == Channel in PubSub
	stat string
}


/**
 * Kill the Client, close the Pump
 */
func (c *WS_Client) kill() {

	if ("dead" == c.stat) {
		return
	}

	c.stat = "dead"

	// Close Sockets
	c.ws.Close()
	c.pub.Close()
	c.sub.Close()

	delete(client_list, c)

	fmt.Println("gwsrps_client-dead")

}


/**
 * Read from Redis send to WebSocket
 */
func (c *WS_Client) pumpRedisToWebSocket() {

	// Wait for channel at most 60s
	waitTick := time.NewTicker(60 * time.Second)

	defer func() {
		fmt.Println("pumpRedisToWebSocket-dead")
		waitTick.Stop()
		c.kill()
	}()

	rch := c.sub.Channel()

	for {

		//fmt.Println("pumpRedisToWebSocket-wait")

		select {
		case msg, ret := <- rch:

			// fmt.Println("pumpRedisToWebSocket-pump!")
			if !ret {
				fmt.Println("pumpRedisToWebSocket-fail", ret)
				return
			}

			// Get a Writer
			w, err078 := c.ws.NextWriter(websocket.TextMessage)
			if err078 != nil {
				return
			}

			// fmt.Println("pumpRedisToWebSocket-send", msg.Payload)
			w.Write([]byte(msg.Payload))
			err085 := w.Close();
			if err085 != nil {
				return
			}

		case <-waitTick.C:
			// Timeout! Restart Loop
			//fmt.Println("pumpRedisToWebSocket-tick!")
			waitTick = time.NewTicker(60 * time.Second)
		}

	}

}


/*
 * Read Message from WebSocket, Publish to Redis
 */
func (c *WS_Client) pumpWebSocketToRedis() {

	defer func() {
		fmt.Println("pumpWebSocketToRedis-dead")
		// waitTick.Stop()
		c.kill()
	}()

	//
	// 		switch (msgType) {
	// 		case websocket.TextMessage:
	// 			fmt.Println("incoming-ws-txt:", string(msgBody))
	// 			// fmt.Println("Text Message")
	// 			break;
	// 		case websocket.BinaryMessage:
	// 			fmt.Println("incoming-ws-bin")
	// 			break;
	// 		}

	for {

		// fmt.Println("pumpWebSocketToRedis-pump")

		_, msgBody, err249 := c.ws.ReadMessage()
		if err249 != nil {
			fmt.Println("pumpWebSocketToRedis-fail", err249)
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// 	fmt.Printf("error: %v", err)
			// }
			return
		}

		// fmt.Println("msgType", msgType)
		// fmt.Println("msgBody", msgBody)

		c.pub.Publish("gwsrps", msgBody)

	}

}
