package ui

import "github.com/mappu/miqt/qt6"

func NewConnectDialog(parent *qt6.QWidget) (*qt6.QDialog) {
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
