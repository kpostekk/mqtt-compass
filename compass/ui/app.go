package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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
	topicTree := topics.CreateVirtualTopicTree()

	config, router, error := mqtt.CreateMqttConnectionConfigConfig(mqtt.MiniConfig{
		ServerUrl: os.Getenv("DEFAULT_MQTT_URL"),
		KeepAlive: 60,
		Username:  "",
		Password:  "",
	})

	router.RegisterHandler("#", func(p *paho.Publish) {
		incomingPacket := topics.IncomingPacket(*p.Packet())

		app.TopicLock.Lock()
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
				Topic: "+/state/+",
			},
		},
	})

	app.CurrentConnection = connectionManager
	ticker := time.NewTicker(1 * time.Second)

	app.CurrentTreeWidget.OnSelectionChanged(func(super func(selected *qt6.QItemSelection, deselected *qt6.QItemSelection), selected, deselected *qt6.QItemSelection) {
		super(selected, deselected)

		app.TopicLock.Lock()

		selectedItemUP := app.CurrentTreeWidget.CurrentItem().UnsafePointer()
		selectedTopic := app.MapUnsafeQTreeItemTopic[selectedItemUP]
		selectedPacket := app.MapTopicPacket[selectedTopic]
		fmt.Println(selectedTopic, selectedPacket)

		app.TopicLock.Unlock()
	})

	go func() {
		for {
			<-ticker.C
			app.TopicLock.Lock()

			mainthread.Wait(func() {
				app.UpdateTree(topicTree)
				topicTree = topics.CreateVirtualTopicTree()

				if app.InitialDraw {
					app.CurrentTreeWidget.ExpandAll()
					app.CurrentTreeWidget.ResizeColumnToContents(0)
					app.InitialDraw = false
				}
			})

			app.TopicLock.Unlock()
		}
	}()

	app.SetupFiltering()

	mainthread.Wait(func() {
		app.CurrentTreeWidget.SetFocus()
	})

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
			a.MapTopicPacket[contextualPathString] = tree.RelatedPacket
			// go blinkUpdateItem(item)
		}

		for _, child := range tree.Children {
			a.UpdateTree(child)
		}

		return item
	}

	item = qt6.NewQTreeWidgetItem()

	item.SetText(0, tree.ContextualPath[len(tree.ContextualPath)-1])

	if tree.RelatedPacket != nil {
		item.SetText(1, string(tree.RelatedPacket.Payload))
		a.MapTopicPacket[contextualPathString] = tree.RelatedPacket
		log.Default().Println("Created", contextualPathString, item)
	}

	font := item.Font(0)
	if tree.IsVirtual() {
		font.SetItalic(true)
	} else {
		font.SetItalic(false)
	}
	item.SetFont(0, font)

	for _, child := range tree.Children {
		childItem := a.UpdateTree(child)
		item.AddChild(childItem)
	}

	if tree.ContextualPath.IsRoot() {
		a.CurrentTreeWidget.AddTopLevelItem(item)
	}

	a.MapTopicQTreeEntry[contextualPathString] = item
	a.MapUnsafeQTreeItemTopic[item.UnsafePointer()] = contextualPathString

	return item
}

func blinkUpdateItem(item *qt6.QTreeWidgetItem) {
	font := item.Font(0)
	font.SetBold(true)
	mainthread.Wait(func() {
		item.SetFont(0, font)
	})

	time.Sleep(200 * time.Millisecond)

	font.SetBold(false)
	mainthread.Wait(func() {
		item.SetFont(0, font)
	})
}

func (app *AppUi) ShowWithParents(item *qt6.QTreeWidgetItem) {
	if item.Parent() != nil {
		app.ShowWithParents(item.Parent())
	}

	item.SetHidden(false)
}

func (app *AppUi) UpdateFiltersResults() {
	if app.CurrentFilterText == "" {
		for _, item := range app.MapTopicQTreeEntry {
			item.SetHidden(false)
		}

		return
	}

	for _, item := range app.MapTopicQTreeEntry {
		item.SetHidden(true)
	}

	for topic, item := range app.MapTopicQTreeEntry {
		matches := strings.Contains(topic, app.CurrentFilterText)

		if matches {
			app.ShowWithParents(item)
		}
	}
}

func (app *AppUi) SetupFiltering() {
	app.CurrentFilterInput.OnTextChanged(func(text string) {
		app.CurrentFilterText = text

		go func() {
			app.TopicLock.RLock()
			mainthread.Wait(func() {
				app.UpdateFiltersResults()
			})
			app.TopicLock.RUnlock()
		}()
	})
}

func NewBrokerViewer(parent *qt6.QWidget) (*qt6.QWidget, *qt6.QTreeWidget, *qt6.QLineEdit) {
	viewerWidget := qt6.NewQWidget(parent)
	layout := qt6.NewQHBoxLayout(viewerWidget)

	layout.SetSpacing(12)
	viewerWidget.SetLayout(layout.QLayout)

	topicsLayout := qt6.NewQVBoxLayout(viewerWidget)
	topicsLayout.SetSpacing(12)
	layout.AddLayout(topicsLayout.QLayout)

	filterInput := qt6.NewQLineEdit(viewerWidget)
	topicsLayout.AddWidget(filterInput.QWidget)

	topicTreeWidget := qt6.NewQTreeWidget(viewerWidget)
	topicTreeWidget.SetColumnCount(2)
	topicTreeWidget.SetHeaderLabels([]string{"Topic", "Value"})
	topicTreeWidget.SetSortingEnabled(true)
	topicsLayout.AddWidget(topicTreeWidget.QWidget)

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

	return viewerWidget, topicTreeWidget, filterInput
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
