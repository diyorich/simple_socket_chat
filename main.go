package main

import (
	"github.com/centrifugal/centrifuge"
	"log"
	"net/http"
)

func auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cred := &centrifuge.Credentials{
			UserID: "",
		}
		newCtx := centrifuge.SetCredentials(ctx, cred)
		r = r.WithContext(newCtx)
		h.ServeHTTP(w, r)
	})
}

func main() {

	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		log.Fatal(err)
	}
	node.OnConnect(func(client *centrifuge.Client) {
		transportName := client.Transport().Name()
		transportProto := client.Transport().Protocol()

		log.Printf("client connected via %s (%s)", transportName, transportProto)

		client.OnSubscribe(func(event centrifuge.SubscribeEvent, callback centrifuge.SubscribeCallback) {
			log.Printf("client subscribes on channel %s", event.Channel)
			callback(centrifuge.SubscribeReply{}, nil)
		})

		client.OnDisconnect(func(event centrifuge.DisconnectEvent) {
			log.Printf("client disconnected")
		})

		client.OnPublish(func(event centrifuge.PublishEvent, callback centrifuge.PublishCallback) {
			node.Publish("chat", event.Data)
		})
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}

	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{})
	http.Handle("/connection/websocket", auth(wsHandler))

	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Printf("Starting server, go to http://localhost:8000")

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
