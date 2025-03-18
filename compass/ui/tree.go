package ui

import (
	"log"
	"mqui/compass/mqtt/topics"

	"github.com/mappu/miqt/qt6"
)

func NewBrokerViewer(parent *qt6.QWidget) (*qt6.QWidget, *qt6.QTreeWidget, *qt6.QLineEdit, *qt6.QLabel, *qt6.QVBoxLayout) {
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

	previewWidget := qt6.NewQWidget(viewerWidget)
	previewLayout := qt6.NewQVBoxLayout(previewWidget)
	previewLayout.SetSpacing(12)
	previewWidget.SetFixedWidth(240)
	layout.AddWidget(previewWidget)

	labelNoTopic := qt6.NewQLabel(previewWidget)
	labelNoTopic.SetText("No topic selected")
	labelNoTopic.Font().SetItalic(true)
	labelNoTopic.SetAlignment(qt6.AlignCenter)
	previewLayout.AddWidget(labelNoTopic.QWidget)

	subscriptionsButton := qt6.NewQPushButton(previewWidget)

	subscriptionsButton.SetText("Subscriptions")
	subscriptionsButton.SetIcon(qt6.QIcon_FromTheme("list-add"))
	previewLayout.AddWidget(subscriptionsButton.QWidget)

	return viewerWidget, topicTreeWidget, filterInput, labelNoTopic, previewLayout
}

func (app *AppUi) UpdateTreeValues(tree *topics.VirtualTopicTree) {
	if tree.RelatedPacket != nil {
		log.Default().Println("Updating entry", tree.ContextualPath.String(), "->", string(tree.RelatedPacket.Payload))
		app.MapTopicQTreeEntry[tree.ContextualPath.String()].SetText(1, string(tree.RelatedPacket.Payload))
		app.MapTopicPacket[tree.ContextualPath.String()] = tree.RelatedPacket
	}
}

func (app *AppUi) RenderTreeSkeleton(tree *topics.VirtualTopicTree) (*qt6.QTreeWidgetItem, bool) {
	if tree.IsRoot() {
		for _, child := range tree.Children {
			app.RenderTreeSkeleton(child)
		}

		return nil, false
	}

	item, itemExists := app.MapTopicQTreeEntry[tree.ContextualPath.String()]

	if !itemExists {
		item = qt6.NewQTreeWidgetItem()
		item.SetExpanded(true)
		item.SetText(0, tree.ContextualPath.Suffix())
		app.MapTopicQTreeEntry[tree.ContextualPath.String()] = item
		app.MapUnsafeQTreeItemTopic[item.UnsafePointer()] = tree.ContextualPath.String()

		log.Default().Println("Created entry", tree.ContextualPath.String(), item)
	}

	for _, child := range tree.Children {
		childItem, _ := app.RenderTreeSkeleton(child)
		item.AddChild(childItem)
	}

	if tree.ContextualPath.IsRoot() {
		app.CurrentTreeWidget.AddTopLevelItem(item)
	}

	return item, !itemExists
}
