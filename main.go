package main

import (
	"log"
	"os"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
)

var io *socketio.Server

func main() {

	engine := gin.New()

	engine.Use(static.Serve("/", static.LocalFile("./public", true)))
	engine.Use(static.Serve("/socket.io/", static.LocalFile("./node_modules/socket.io-client/dist", false)))
	engine.Any("/socket.io/", func(ctx *gin.Context) {
		io.ServeHTTP(ctx.Writer, ctx.Request)
	})

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":3000"
	}

	log.Fatal(engine.Run(PORT))
}

func init() {
	var err error

	io, err = socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	var (
		numUsers int32
		session  = make(map[string]string)
	)

	io.On("connection", func(socket socketio.Socket) {
		var addedUser = false

		socket.Join("chat")

		socket.On("new message", func(msg string) {
			socket.BroadcastTo("chat", "new message", map[string]interface{}{
				"username": session[socket.Id()],
				"message":  msg,
			})
		})

		socket.On("add user", func(username string) {
			if addedUser {
				return
			}

			session[socket.Id()] = username
			numUsers++
			addedUser = true

			socket.Emit("login", map[string]interface{}{
				"numUsers": numUsers,
			})

			socket.BroadcastTo("chat", "user joined", map[string]interface{}{
				"username": username,
				"numUsers": numUsers,
			})
		})

		socket.On("typing", func() {
			socket.BroadcastTo("chat", "typing", map[string]interface{}{
				"username": session[socket.Id()],
			})
		})

		socket.On("stop typing", func() {
			socket.BroadcastTo("chat", "stop typing", map[string]interface{}{
				"username": session[socket.Id()],
			})
		})

		socket.On("disconnection", func() {
			if !addedUser {
				return
			}

			delete(session, socket.Id())

			numUsers--
			socket.BroadcastTo("chat", "user left", map[string]interface{}{
				"username": session[socket.Id()],
				"numUsers": numUsers,
			})
		})
	})

	io.On("error", func(socket socketio.Socket) {
		log.Println("error in io", err)
	})
}
