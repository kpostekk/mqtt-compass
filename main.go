package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"os"

	"github.com/eclipse/paho.golang/paho"
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

const URL_REGEX = `^mqtts?://[0-9a-zA-Z\.\-]+(:[0-9]+)?$`

func main() {
	log.Default().Println("Hello, World!")
	qt6.NewQApplication(os.Args)

	windowMain := qt6.NewQMainWindow2()
	windowMain.SetWindowTitle("MQTT Compass")
	windowMain.SetMinimumSize(
		qt6.NewQSize2(640, 480),
	)
	windowMain.Show()

	widgetMain := qt6.NewQWidget2()
	layoutMain := qt6.NewQVBoxLayout(widgetMain)
	// layoutMain.SetDirection(qt6.QBoxLayout__Down)
	widgetMain.SetLayout(layoutMain.QLayout)

	windowMain.SetCentralWidget(widgetMain)

	filterInput := qt6.NewQLineEdit(widgetMain)
	filterInput.SetPlaceholderText("Filter")
	// filterInput.SetDisabled(true)
	layoutMain.AddWidget(filterInput.QWidget)

	tree := qt6.NewQTreeWidget(widgetMain)
	cols := []string{"Topic", "Value"}
	tree.SetColumnCount(len(cols))
	tree.SetHeaderLabels(cols)
	tree.SetSortingEnabled(true)
	tree.SortByColumn(0, qt6.AscendingOrder)
	tree.SetContextMenuPolicy(qt6.CustomContextMenu)
	tree.OnCustomContextMenuRequested(func(pos *qt6.QPoint) {
		menu := qt6.NewQMenu(tree.QWidget)
		clipboard := qt6.UnsafeNewQClipboard(menu.UnsafePointer())
		treeItem := tree.ItemAt(pos)

		if treeItem == nil {
			return
		}

		copyValueAction := qt6.NewQAction3(qt6.QIcon_FromTheme("edit-copy"), "Copy Value")
		copyTopicFullAction := qt6.NewQAction3(qt6.QIcon_FromTheme("edit-copy"), "Copy Topic (Full)")
		copyTopicPartAction := qt6.NewQAction3(qt6.QIcon_FromTheme("edit-copy"), "Copy Value (Part)")

		copyValueAction.OnTriggered(func() {
			value := treeItem.Text(1)
			clipboard.SetText(value)
		})

		copyTopicFullAction.OnTriggered(func() {
			panic("Not implemented")
		})

		copyTopicPartAction.OnTriggered(func() {
			topicPart := treeItem.Text(0)
			clipboard.SetText(topicPart)
		})

		menu.AddAction(copyValueAction)
		menu.AddAction(copyTopicFullAction)
		menu.AddAction(copyTopicPartAction)

		menu.AddSeparator()

		inspectAction := qt6.NewQAction3(qt6.QIcon_FromTheme("document-properties"), "Inspect")
		inspectAction.OnTriggered(func() {
			fmt.Println("Inspect triggered")
		})

		menu.AddAction(inspectAction)

		posViewportMapped := tree.Viewport().MapToGlobal(pos.ToPointF()).ToPoint()

		menu.ExecWithPos(posViewportMapped)
	})
	layoutMain.AddWidget(tree.QWidget)

	packetsChannel := make(chan paho.Publish)

	brokerUrl := "mqtt://127.0.0.1:1883"
	// brokerUrl := ""

	IsValidMqttUrl := func(url string) bool {
		return regexp.MustCompile(URL_REGEX).MatchString(url)
	}

	connectWindow := qt6.NewQDialog(windowMain.QWidget)
	connectWindowLayout := qt6.NewQVBoxLayout(connectWindow.QWidget)
	connectWindowLayout.SetContentsMargins(12, 12, 12, 12)
	connectWindow.SetLayout(connectWindowLayout.QLayout)
	connectWindow.SetFocusPolicy(qt6.StrongFocus)
	connectWindow.SetWindowModality(qt6.ApplicationModal)
	connectWindow.SetFixedSize(qt6.NewQSize2(480, 120))

	urlInput := qt6.NewQLineEdit(connectWindow.QWidget)
	urlInput.SetPlaceholderText("Broker URL")
	urlInput.SetText(brokerUrl)
	connectWindowLayout.AddWidget(urlInput.QWidget)

	connectButton := qt6.NewQPushButton(urlInput.QWidget)
	connectButton.SetEnabled(false)
	connectButton.SetText("Connect")
	connectButton.SetIcon(
		qt6.QIcon_FromTheme("network-connect"),
	)
	connectWindowLayout.AddWidget(connectButton.QWidget)

	connectWindow.Show()

	topicMap := make(TopicMap)
	topicMapMutex := sync.RWMutex{}
	topicPointers := make(map[string]*qt6.QTreeWidgetItem)
	ticker := time.NewTicker(2_000 * time.Millisecond)

	RebuildTree := func(tm *TopicMap) {
		topicTree := tm.IntoTree()
		items, pointers := CreateQTreeContents(topicTree)

		tree.Clear()
		tree.AddTopLevelItems(items)
		tree.ExpandAll()
		tree.ResizeColumnToContents(0)
		tree.SortByColumn(0, qt6.AscendingOrder)
		topicPointers = pointers
	}

	UpdateTree := func(tm *TopicMap) {
		updatedValues := 0

		for k, v := range *tm {
			if item := topicPointers[k]; item != nil {
				item.SetText(1, string(v))
				updatedValues++
			} else {
				log.Default().Printf("Topic %s not found in tree!\n", k)
				RebuildTree(tm)
				return
			}
		}

		log.Default().Printf("Updated %d values in tree\n", updatedValues)
	}

	FilterTree := func(filter string) {
		if filter == "" {
			for _, v := range topicPointers {
				v.SetHidden(false)
			}
			return
		}

		for k, v := range topicPointers {
			if strings.Contains(k, filter) {
				v.SetHidden(false)
			} else {
				if IsParentOfAChildWithTopicPart(v, filter) {
					v.SetHidden(false)
					continue
				}

				v.SetHidden(true)
			}
		}

	}

	urlInput.OnTextChanged(func(newText string) {
		if IsValidMqttUrl(newText) {
			connectButton.SetEnabled(true)
		} else {
			connectButton.SetEnabled(false)
		}
	})

	connectButton.OnClicked(func() {
		brokerUrl = urlInput.Text()
		connectWindow.Hide()

		connection, router, ctx := connectToBroker(&brokerUrl)

		router.RegisterHandler("#", func(p *paho.Publish) {
			packetsChannel <- *p
		})

		go func() {
			connection.AwaitConnection(ctx)
			connection.Subscribe(ctx, &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{
						Topic: "#",
						QoS:   0,
					},
				},
			})

			time.Sleep(200 * time.Millisecond)

			mainthread.Wait(func() {
				topicMapMutex.RLock()
				RebuildTree(&topicMap)
				topicMapMutex.RUnlock()
			})

			for {
				select {
				case <-ctx.Done():
					return
				case p := <-packetsChannel:
					topicMapMutex.Lock()
					topicMap[p.Topic] = p.Payload
					topicMapMutex.Unlock()
				}
			}
		}()
	})

	go func() {
		for {
			select {
			case <-ticker.C:
				mainthread.Wait(func() {
					topicMapMutex.RLock()
					UpdateTree(&topicMap)
					topicMapMutex.RUnlock()
				})
			}
		}
	}()

	filterInput.OnTextChanged(func(newText string) {
		topicMapMutex.RLock()
		FilterTree(newText)
		topicMapMutex.RUnlock()
	})

	widgetMain.Show()

	qt6.QApplication_Exec()
}
