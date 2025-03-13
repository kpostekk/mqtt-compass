package mqtt

import (
	"context"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	"mqui/compass/mqtt/topics"
)

type MiniConfig struct {
	ServerUrl string
	KeepAlive uint16
	Username  string
	Password  string
}

func CreateMqttConnectionConfigConfig(miniConfig MiniConfig) (*autopaho.ClientConfig, *paho.StandardRouter, error) {
	urlParsed, err := url.Parse(miniConfig.ServerUrl)

	if err != nil {
		return nil, nil, err
	}

	router := paho.NewStandardRouter()

	config := autopaho.ClientConfig{
		ServerUrls:      []*url.URL{urlParsed},
		KeepAlive:       miniConfig.KeepAlive,
		ConnectUsername: miniConfig.Username,
		ConnectPassword: []byte(miniConfig.Password),
		ClientConfig: paho.ClientConfig{
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					router.Route(pr.Packet.Packet())
					return true, nil
				},
			},
		},
	}

	return &config, router, nil
}

func CreateConnectionManager(ctx context.Context, connectionConfig *autopaho.ClientConfig) (*autopaho.ConnectionManager, error) {
	connectionManager, err := autopaho.NewConnection(ctx, *connectionConfig)

	if err != nil {
		return nil, err
	}

	return connectionManager, nil
}

type ConnectionStore struct {
	Topics topics.VirtualTopicMap
}
