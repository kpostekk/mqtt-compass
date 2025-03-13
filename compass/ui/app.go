package ui

import (
	"context"
	"log"
	"sync"
	"time"

	// "time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	// "github.com/mappu/miqt/qt6/mainthread"

	"mqui/compass/mqtt"
	"mqui/compass/mqtt/topics"
)

type AppUi struct {
	MainWindow         *qt6.QMainWindow
	CurrentConnection  *autopaho.ConnectionManager
	CurrentTreeWidget  *qt6.QTreeWidget
	MapTopicPacket     map[string]*topics.IncomingPacket
	MapTopicQTreeEntry map[string]*qt6.QTreeWidgetItem
	TopicLock          sync.RWMutex
	InitialDraw        bool
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

	brokerViewer, brokerTreeWidget := NewBrokerViewer(widgetMain)

	layoutMain.AddWidget(brokerViewer)

	return &AppUi{
		MainWindow:         windowMain,
		CurrentTreeWidget:  brokerTreeWidget,
		MapTopicQTreeEntry: make(map[string]*qt6.QTreeWidgetItem),
		MapTopicPacket:     make(map[string]*topics.IncomingPacket),
		InitialDraw:        true,
		TopicLock:          sync.RWMutex{},
	}
}

func (app *AppUi) SetConnection(ctx *context.Context) {
	topicTree := topics.CreateVirtualTopicTree()

	config, router, error := mqtt.CreateMqttConnectionConfigConfig(mqtt.MiniConfig{
		ServerUrl: "mqtt://127.0.0.1:31883",
		KeepAlive: 60,
		Username:  "",
		Password:  "",
	})

	router.RegisterHandler("#", func(p *paho.Publish) {
		app.TopicLock.Lock()

		incomingPacket := topics.IncomingPacket(*p.Packet())
		topicTree.AddPacket(&incomingPacket)

		app.TopicLock.Unlock()
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
		},
	})

	app.CurrentConnection = connectionManager
	ticker := time.NewTicker(1 * time.Second)

	app.CurrentTreeWidget.OnSelectionChanged(func(super func(selected *qt6.QItemSelection, deselected *qt6.QItemSelection), selected, deselected *qt6.QItemSelection) {
		super(selected, deselected)

		selectedItem := app.CurrentTreeWidget.CurrentItem()

		go func() {
			app.TopicLock.Lock()
			log.Println("Selected item", selectedItem)
			for item, packet := range app.MapQTreeItemPacket {
				log.Println("Candidate", item, item.Text(0) == packet.Topic, packet.Topic)
			}

			if packet, ok := app.MapQTreeItemPacket[selectedItem]; ok {
				log.Println("Selected item has packet", packet.Topic)
			}

			app.TopicLock.Unlock()
		}()
	})

	go func() {
		for range ticker.C {
			mainthread.Wait(func() {
				app.TopicLock.Lock()
				app.UpdateTree(topicTree)
				topicTree = topics.CreateVirtualTopicTree()
				app.TopicLock.Unlock()

				if app.InitialDraw {
					app.CurrentTreeWidget.ExpandAll()
					app.CurrentTreeWidget.ResizeColumnToContents(0)
					app.InitialDraw = false
				}
			})
		}
	}()
}

func (a *AppUi) UpdateTree(tree *topics.VirtualTopicTree) *qt6.QTreeWidgetItem {
	if tree.IsRoot() {
		for _, child := range tree.Children {
			a.UpdateTree(child)
		}

		return nil
	}

	// check if tree is already in the map
	contextualPathString := tree.ContextualPath.String()
	item, itemExists := a.MapTopicQTreeEntry[contextualPathString]

	if itemExists {
		// log.Default().Println("Found!", contextualPathString, a.MapTopicQTreeEntry[contextualPathString])

		if tree.RelatedPacket != nil {
			item.SetText(1, string(tree.RelatedPacket.Payload))
		}

		for _, child := range tree.Children {
			a.UpdateTree(child)
		}

		return item
	}

	item = qt6.NewQTreeWidgetItem()
	a.MapTopicQTreeEntry[contextualPathString] = item

	item.SetText(0, contextualPathString)

	if tree.RelatedPacket != nil {
		item.SetText(1, string(tree.RelatedPacket.Payload))
		a.MapQTreeItemPacket[item] = tree.RelatedPacket
		log.Default().Println("Created", contextualPathString, item)
	}

	for _, child := range tree.Children {
		childItem := a.UpdateTree(child)
		item.AddChild(childItem)
	}

	if tree.ContextualPath.IsRoot() {
		a.CurrentTreeWidget.AddTopLevelItem(item)
	}

	return item
}

func NewBrokerViewer(parent *qt6.QWidget) (*qt6.QWidget, *qt6.QTreeWidget) {
	viewerWidget := qt6.NewQWidget(parent)
	layout := qt6.NewQHBoxLayout(viewerWidget)

	layout.SetSpacing(12)
	viewerWidget.SetLayout(layout.QLayout)

	topicTreeWidget := qt6.NewQTreeWidget(viewerWidget)
	topicTreeWidget.SetColumnCount(2)
	topicTreeWidget.SetHeaderLabels([]string{"Topic", "Value"})
	topicTreeWidget.SetSortingEnabled(true)
	layout.AddWidget(topicTreeWidget.QWidget)

	detailsWidget := qt6.NewQWidget(viewerWidget)
	detailsLayout := qt6.NewQVBoxLayout(detailsWidget)
	detailsLayout.SetSpacing(12)
	detailsWidget.SetFixedWidth(240)
	layout.AddWidget(detailsWidget)

	labelNoTopic := qt6.NewQLabel(detailsWidget)
	labelNoTopic.SetText("No topic selected")
	labelNoTopic.Font().SetItalic(true)
	labelNoTopic.SetAlignment(qt6.AlignCenter)
	detailsLayout.AddWidget(labelNoTopic.QWidget)

	subscriptionsButton := qt6.NewQPushButton(detailsWidget)

	subscriptionsButton.SetText("Subscriptions")
	subscriptionsButton.SetIcon(qt6.QIcon_FromTheme("list-add"))
	detailsLayout.AddWidget(subscriptionsButton.QWidget)

	return viewerWidget, topicTreeWidget
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
