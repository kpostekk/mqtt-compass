package compass

import (
	"context"
	"fmt"
	"log"
	"mqui/compass/mqtt"
	"mqui/compass/mqtt/topics"
	"mqui/compass/ui"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"os"

	"github.com/eclipse/paho.golang/paho"
	"github.com/mappu/miqt/qt6"
	"github.com/joho/godotenv"
)

const URL_REGEX = `^mqtts?://[0-9a-zA-Z\.\-]+(:[0-9]+)?$`

func StartUi() {
	godotenv.Load()
	ctx, _ := context.WithCancel(context.Background())

	log.Default().Println("Hello, World!")
	qt6.NewQApplication(os.Args)

	appUi := ui.NewAppUi()
	appUi.MainWindow.Show()
	appUi.SetConnection(&ctx)

	qt6.QApplication_Exec()
}

func StartHeadless() {
	ctx, cancel := context.WithCancel(context.Background())

	topicMap := make(topics.VirtualTopicMap)
	topicMutex := sync.RWMutex{}

	config, router, error := mqtt.CreateMqttConnectionConfigConfig(mqtt.MiniConfig{
		ServerUrl: os.Getenv("DEFAULT_MQTT_URL"),
		KeepAlive: 60,
		Username:  "",
		Password:  "",
	})

	router.RegisterHandler("#", func(p *paho.Publish) {
		// log.Default().Printf("Received packet on topic %s\n", p.Topic)
		topicMutex.Lock()
		topicMap[p.Topic] = topics.IncomingPacket(*p.Packet())
		topicMutex.Unlock()
	})

	if error != nil {
		panic(error)
	}

	connectionManager, error := mqtt.CreateConnectionManager(ctx, config)

	if error != nil {
		panic(error)
	}

	connectionManager.AwaitConnection(ctx)
	log.Default().Println("Connected to broker!")

	connectionManager.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: "#",
			},
		},
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<-sigs
		connectionManager.Disconnect(ctx)
		cancel()
	}()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		tree := topics.CreateVirtualTopicTree()

		for _ = range ticker.C {
			topicMutex.RLock()
			for _, packet := range topicMap {
				tree.AddPacket(&packet)
			}
			topicMutex.RUnlock()

			fmt.Println(tree)
		}
	}()

	<-ctx.Done()
}
