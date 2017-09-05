// deis-logger
// https://github.com/topfreegames/deis-logger
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

//LogTailHandler handler
type LogTailHandler struct {
	App *App
}

// NewLogTailHandler creates a new healthcheck handler
func NewLogTailHandler(a *App) *LogTailHandler {
	m := &LogTailHandler{App: a}
	return m
}

//ServeHTTP method
func (h *LogTailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connectionOpen := true
	notify := w.(http.CloseNotifier).CloseNotify()
	read, write := io.Pipe()

	go func() {
		<-notify
		connectionOpen = false
	}()

	app := mux.Vars(r)["app"]
	pubSub := redis.PubSubConn{Conn: h.App.Redis.Get()}
	channel := "logger-" + app
	pubSub.Subscribe(channel)
	defer pubSub.Unsubscribe(channel)

	go func() {
		defer write.Close()
		for {
			if !connectionOpen {
				pubSub.Close()
				fmt.Print("Done!")
				break
			}
			switch v := pubSub.Receive().(type) {
			case redis.Message:
				message := string(v.Data[:])
				message = strings.TrimPrefix(message, "\"")
				message = strings.TrimSuffix(message, "\\n\"")

				io.Copy(write, strings.NewReader(message))
				io.Copy(write, strings.NewReader("\n"))
			}
		}
	}()
	io.Copy(w, read)
}
