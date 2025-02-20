package main

import (
	"context"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func connectToBroker(urlString *string) (*autopaho.ConnectionManager, *paho.StandardRouter, context.Context) {
	ctx := context.Background()

	urlObj, _ := url.Parse(*urlString)

	router := paho.NewStandardRouter()

	config := autopaho.ClientConfig{
		ServerUrls: []*url.URL{urlObj},
		KeepAlive:  20,
		ClientConfig: paho.ClientConfig{
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					router.Route(pr.Packet.Packet())
					return true, nil
				}},
		},
	}

	connection, _ := autopaho.NewConnection(ctx, config)

	// connection.AwaitConnection(ctx)

	// connection.Subscribe(ctx, &paho.Subscribe{
	// 	Subscriptions: []paho.SubscribeOptions{
	// 		{
	// 			Topic: "ledatel_pr116/+/state/+",
	// 			QoS:   0,
	// 		},
	// 	},
	// })

	// create string string map
	// topicTree := make(TopicTree)

	// router.RegisterHandler("#", func(p *paho.Publish) {
	// 	// log.Default().Printf("Received message (%s): %s\n", p.Topic)
	// 	// fmt.Printf("Received message (%s): %s", p.Topic, string(p.Payload))

	// 	l := topicTree.Add(p.Topic, p.Payload)
	// 	fmt.Print("\n---\n")
	// 	fmt.Println(l)
	// })

	return connection, router, ctx

	// <-ctx.Done()
}
