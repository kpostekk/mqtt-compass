package compass

import (
	"context"
	"log"
	"mqui/compass/ui"
	"os"

	"github.com/joho/godotenv"
	"github.com/mappu/miqt/qt6"
)

const URL_REGEX = `^mqtts?://[0-9a-zA-Z\.\-]+(:[0-9]+)?$`

func StartUi() {
	godotenv.Load()
	ctx := context.Background()

	log.Default().Println("Hello, World!")
	qt6.NewQApplication(os.Args)

	appUi := ui.NewAppUi()
	appUi.MainWindow.Show()
	go appUi.SetConnection(&ctx)

	qt6.QApplication_Exec()
}
