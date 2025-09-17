package main

import (
	"embed"
	"encoding/base64"
	"log"
	"os"
	"time"

	"github.com/awnumar/memguard"
	"github.com/wailsapp/wails/v3/pkg/application"

	"duckysigner/internal/kmd/config"
	"duckysigner/services"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/build
var assets embed.FS

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	// Starts an interrupt handler that will clean up by will wipe sensitive
	// data in memory before terminating suddenly
	memguard.CatchInterrupt()

	// Create KMD service
	kmdService := &services.KMDService{
		Config: config.KMDConfig{
			SessionLifetimeSecs: uint64((1 * time.Hour).Seconds()),
			DriverConfig: config.DriverConfig{
				ParquetWalletDriverConfig: config.ParquetWalletDriverConfig{
					ScryptParams: config.ScryptParams{
						ScryptN: 65536,
						ScryptR: 1,
						ScryptP: 32,
					},
				},
				SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{Disable: true, UnsafeScrypt: true},
				LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
			},
		},
	}
	// Clean up KMD when application terminates and we're returning from this
	// main function
	defer kmdService.CleanUp()

	// Create dApp connect service
	dcService := &services.DappConnectService{
		HideServerBanner:    true,
		UserResponseTimeout: 5 * time.Minute,
	}
	// Clean up dApp connect service when application terminates and we're
	// returning from this main function
	defer dcService.CleanUp()

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Ducky Signer",
		Description: "Experimental desktop wallet for Algorand",
		Services: []application.Service{
			// application.NewService(&services.GreetService{}),
			application.NewService(kmdService),
			application.NewService(dcService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	dcService.WailsApp = app

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Ducky Signer",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		// BackgroundColour: application.NewRGB(27, 38, 54),
		URL:    "/",
		Width:  1024,
		Height: 768,
		// StartState: application.WindowStateMaximised,
	})

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	app.Event.On("saveFile", func(e *application.CustomEvent) {
		dataBytes, _ := base64.StdEncoding.DecodeString(e.Data.(string))
		fileLoc, _ := application.SaveFileDialog().
			AddFilter("Algorand Transaction", "*.txn.msgpack").
			HideExtension(true).
			PromptForSingleSelection()
		os.WriteFile(fileLoc, dataBytes, 0666)
	})

	// Run the application. This blocks until the application has been exited.
	// If an error occurred while running the application, log it and exit.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
