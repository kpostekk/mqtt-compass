package ui

import "github.com/mappu/miqt/qt6"

func expandWithParents(item *qt6.QTreeWidgetItem) {
	if item.Parent() != nil {
		expandWithParents(item.Parent())
		item.Parent().SetExpanded(true)
	}
}

func showWithParents(item *qt6.QTreeWidgetItem) {
	if item.Parent() != nil {
		showWithParents(item.Parent())
	}

	item.SetHidden(false)
}
