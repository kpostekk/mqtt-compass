package ui

import (
	"log"
	"mqui/compass/mqtt/topics"

	"github.com/mappu/miqt/qt6"
)

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

func (app *AppUi) UpdateTree(tree *topics.VirtualTopicTree) {
	if tree.RelatedPacket != nil {
		log.Default().Println("Updating entry", tree.ContextualPath.String())
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