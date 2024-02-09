package handlers

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

// empty websocket to be used for live updates
func WebsocketHandler(c echo.Context) error {
	 websocket.Handler(func(ws *websocket.Conn) {
            defer ws.Close()
            for {
                // do nothing but keep the connection open

                err := websocket.Message.Send(ws, "Hello, Client!")
                if err != nil {
                        return
                        }
                msg := ""
                err = websocket.Message.Receive(ws, &msg)
                if err != nil {
                        return
                        }
                        fmt.Println("Received message:", msg)
        }
        }).ServeHTTP(c.Response(), c.Request())
        return nil
}
