package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"unsafe"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"mqui/compass/mqtt"
	"mqui/compass/mqtt/topics"
)

type AppUi struct {
	MainWindow              *qt6.QMainWindow
	CurrentConnection       *autopaho.ConnectionManager
	CurrentTreeWidget       *qt6.QTreeWidget
	CurrentFilterInput      *qt6.QLineEdit
	MapTopicPacket          map[string]*topics.IncomingPacket
	MapTopicQTreeEntry      map[string]*qt6.QTreeWidgetItem
	MapUnsafeQTreeItemTopic map[unsafe.Pointer]string
	TopicLock               sync.RWMutex
	CurrentFilterText       string
	InitialDraw             bool
}

func NewAppUi() *AppUi {
	windowMain := qt6.NewQMainWindow2()

	windowMain.SetWindowTitle("MQTT Compass (Not Connected)")

	windowMain.SetMinimumSize(
		qt6.NewQSize2(640, 480),
	)

	appMenu := qt6.NewQMenuBar(windowMain.Window())

	connectionMenu := qt6.NewQMenu(appMenu.QWidget)
	connectionMenu.SetTitle("Connection")
	connectAction := qt6.NewQAction2("Connect")
	connectAction.SetIcon(qt6.QIcon_FromTheme("network-connect"))
	connectionMenu.AddAction(connectAction)
	disconnectAction := qt6.NewQAction2("Disconnect")
	disconnectAction.SetIcon(qt6.QIcon_FromTheme("network-disconnect"))
	connectionMenu.AddAction(disconnectAction)
	appMenu.AddMenu(connectionMenu)

	windowMain.SetMenuBar(appMenu)

	widgetMain := qt6.NewQWidget2()
	layoutMain := qt6.NewQVBoxLayout(widgetMain)
	widgetMain.SetLayout(layoutMain.QLayout)
	windowMain.SetCentralWidget(widgetMain)

	brokerViewer, brokerTreeWidget, brokerFilterField := NewBrokerViewer(widgetMain)

	layoutMain.AddWidget(brokerViewer)

	return &AppUi{
		MainWindow:              windowMain,
		CurrentTreeWidget:       brokerTreeWidget,
		CurrentFilterInput:      brokerFilterField,
		MapTopicQTreeEntry:      make(map[string]*qt6.QTreeWidgetItem),
		MapTopicPacket:          make(map[string]*topics.IncomingPacket),
		MapUnsafeQTreeItemTopic: make(map[unsafe.Pointer]string),
		InitialDraw:             true,
		TopicLock:               sync.RWMutex{},
		CurrentFilterText:       "",
	}
}

func (app *AppUi) SetConnection(ctx *context.Context) {
	packetChan := make(chan *topics.IncomingPacket)

	config, router, error := mqtt.CreateMqttConnectionConfigConfig(mqtt.MiniConfig{
		ServerUrl: os.Getenv("DEFAULT_MQTT_URL"),
		KeepAlive: 60,
		Username:  "",
		Password:  "",
	})

	router.RegisterHandler("#", func(p *paho.Publish) {
		incomingPacket := topics.IncomingPacket(*p.Packet())

		packetChan <- &incomingPacket
	})

	if error != nil {
		panic(error)
	}

	connectionManager, error := mqtt.CreateConnectionManager(*ctx, config)

	if error != nil {
		panic(error)
	}

	connectionManager.AwaitConnection(*ctx)
	log.Default().Println("Connected to broker!")

	connectionManager.Subscribe(*ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: "#",
			},
			{
				Topic: "$SYS/#",
			},
		},
	})

	app.CurrentConnection = connectionManager

	app.CurrentTreeWidget.OnSelectionChanged(func(super func(selected *qt6.QItemSelection, deselected *qt6.QItemSelection), selected, deselected *qt6.QItemSelection) {
		super(selected, deselected)

		app.TopicLock.RLock()

		selectedItemUP := app.CurrentTreeWidget.CurrentItem().UnsafePointer()
		selectedTopic := app.MapUnsafeQTreeItemTopic[selectedItemUP]
		selectedPacket := app.MapTopicPacket[selectedTopic]
		fmt.Println(selectedTopic, selectedPacket)

		app.TopicLock.RUnlock()
	})

	// app.CurrentTreeWidget.SetColumnWidth(0, 240)

	go func() {
		pt := topics.CreateVirtualTopicTree()
		for {
			rcvPacket := <-packetChan
			log.Default().Println("Received packet", rcvPacket.Topic)

			app.TopicLock.Lock()
			updatesTree := pt.AddPacket(rcvPacket)

			mainthread.Wait(func() {
				app.RenderTreeSkeleton(pt)
			})

			mainthread.Wait(func() {
				app.UpdateTree(updatesTree)
			})

			app.TopicLock.Unlock()
		}
	}()

	app.SetupFiltering()

	mainthread.Wait(func() {
		app.CurrentTreeWidget.SetFocus()
	})
}

func NewConnectDialog(parent *qt6.QWidget) *qt6.QDialog {
	dialog := qt6.NewQDialog(parent)
	dialog.SetFixedSize(
		qt6.NewQSize2(320, 240),
	)
	dialog.SetWindowTitle("Connect to MQTT Broker")
	dialog.SetModal(true)

	layout := qt6.NewQFormLayout(dialog.QWidget)
	layout.SetFormAlignment(qt6.AlignCenter)

	brokerAddress := qt6.NewQLineEdit(dialog.QWidget)
	brokerAddress.SetPlaceholderText("mqtt://localhost:1883")
	layout.AddRow3("Broker Address", brokerAddress.QWidget)

	connectButton := qt6.NewQPushButton(dialog.QWidget)
	connectButton.SetText("Connect")
	connectButton.SetIcon(qt6.QIcon_FromTheme("network-connect"))
	layout.AddRowWithWidget(connectButton.QWidget)

	// connectButton.OnClicked(func() {
	// 	dialog.Close()
	// })

	// dialog.OnCloseEvent(func(_ func(_ *qt6.QCloseEvent), _ *qt6.QCloseEvent) {
	// 	qt6.QCoreApplication_Quit()
	// })

	return dialog
}
