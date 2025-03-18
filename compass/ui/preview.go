package ui

import (
	"mqui/compass/mqtt/topics"

	"github.com/mappu/miqt/qt6"
)

func CreatePreviewWidget(packet *topics.IncomingPacket) *qt6.QTableWidget {
	widget := qt6.NewQTableWidget2()
	widget.SetColumnCount(2)
	widget.SetRowCount(2)
	widget.SetHorizontalHeaderLabels([]string{"Key", "Value"})
	widget.VerticalHeader().SetVisible(false)

	widget.SetItem(0, 0, qt6.NewQTableWidgetItem2("Topic"))
	widget.SetItem(0, 1, qt6.NewQTableWidgetItem2(packet.Topic))

	widget.SetItem(1, 0, qt6.NewQTableWidgetItem2("Payload"))
	widget.SetItem(1, 1, qt6.NewQTableWidgetItem2(string(packet.Payload)))

	widget.ResizeColumnsToContents()

	return widget
}
