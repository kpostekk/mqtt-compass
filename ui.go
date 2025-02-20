package main

import (
	"strings"

	"github.com/mappu/miqt/qt6"
)

func CreateQTreeContents(t *TopicTree) ([]*qt6.QTreeWidgetItem, map[string]*qt6.QTreeWidgetItem) {
	itemList := make([]*qt6.QTreeWidgetItem, 0)
	treeItemPointers := make(map[string]*qt6.QTreeWidgetItem)

	for k, v := range *t {
		newItem := qt6.NewQTreeWidgetItem()
		treeItemPointers[v.TopicCanonical()] = newItem
		newItem.SetText(0, k)
		newItem.SetText(1, string(v.Value))
		newItem.SetToolTip(0, v.TopicCanonical())
		// newItem.TreeWidget().SetContextMenuPolicy(qt6.CustomContextMenu)
		// newItem.TreeWidget().OnCustomContextMenuRequested(func(pos *qt6.QPoint) {
		// 	fmt.Println(pos)
		// 	// menu := qt6.NewQMenu2()
		// 	// me
		// })

		itemList = append(itemList, newItem)

		subItems, subPointers := CreateQTreeContents(&v.Children)
		newItem.AddChildren(subItems)

		for k, v := range subPointers {
			treeItemPointers[k] = v
		}
	}

 	return itemList, treeItemPointers
}

func ExpandParents(item *qt6.QTreeWidgetItem) {
	if item.Parent() != nil {
		item.Parent().SetExpanded(true)
		ExpandParents(item.Parent())
	}
}

func IsParentOfAChildWithTopicPart(maybeParent *qt6.QTreeWidgetItem, filter string) bool {
	isMatchingContent := strings.Contains(maybeParent.Text(0), filter)

	if maybeParent.ChildCount() == 0 {
		return isMatchingContent
	}

	for childIndex := 0; childIndex < maybeParent.ChildCount(); childIndex++ {
		if IsParentOfAChildWithTopicPart(maybeParent.Child(childIndex), filter) {
			return true
		}
	}

	return isMatchingContent
}

// func populateQTreeItem(qTreeItem *qt6.QTreeWidgetItem, tt *TopicTree) {
// 	for k, v := range *tt {
// 		newItem := qt6.NewQTreeWidgetItem()
// 		newItem.SetText(0, k)
// 		newItem.SetText(1, string(v.Value))
// 		qTreeItem.AddChild(newItem)
// 		populateQTreeItem(newItem, &v.Children)
// 	}
// }
