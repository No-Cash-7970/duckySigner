package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorand/go-algorand/logging"
	"github.com/wailsapp/wails/v3/pkg/application"

	"duckysigner/kmd/config"
	"duckysigner/kmd/wallet"
	"duckysigner/kmd/wallet/driver"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Ducky Signer",
		Description: "Experimental desktop wallet for Algorand",
		Bind: []any{
			&GreetService{},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title: "Ducky Signer",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		// BackgroundColour: application.NewRGB(27, 38, 54),
		URL:      "/",
		Width:    1024,
		Height:   768,
		Centered: true,
		// StartState: application.WindowStateMaximised,
	})

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Events.Emit(&application.WailsEvent{
				Name: "time",
				Data: now,
			})
			time.Sleep(time.Second)
		}
	}()

	// ====================================================================== //

	// Create a pw
	pw := []byte("password") // DANGEROUS INSECURE PASSWORD

	// Load or generate config
	log.Println("Loading KMD config...")
	kmdConfig, _ := config.LoadKMDConfig("")
	kmdConfigJson, _ := json.MarshalIndent(kmdConfig, "", "  ")
	log.Println(string(kmdConfigJson))
	log2 := logging.NewLogger()
	log2.SetLevel(logging.Info)

	// Initialize and fetch the wallet driver
	log.Println("Initializing SQLite driver...")
	driver.InitWalletDrivers(kmdConfig, log2)
	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		log.Fatal(err)
	}

	// Create new wallet
	walletID, _ := wallet.GenerateWalletID()
	// walletID := []byte("8f919d5883bc375c3dfb59c9865b87c6")
	log.Println("Wallet ID:", string(walletID))
	log.Println("Creating new wallet...")
	err2 := sqliteDriver.CreateWallet([]byte("test wallet"), walletID, pw, types.MasterDerivationKey{})
	if err2 != nil {
		log.Fatal(err2)
	}

	// Fetch newly created newWallet
	log.Println("Fetching newly created wallet...")
	newWallet, err := sqliteDriver.FetchWallet(walletID)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Display data about newly created wallet
	walletMetadata, _ := newWallet.Metadata()
	walletMetadataJson, _ := json.MarshalIndent(walletMetadata, "", "  ")
	log.Println(string(walletMetadataJson))

	// DANGEROUS: Load encryption password & MDK into memory for certain operations like retrieving the MDK or generating an account
	// Need to use mlock, memory enclave or something to protect the secrets in memory
	newWallet.Init(pw)

	// DANGEROUS: Display master derivation key
	mdk, err := newWallet.ExportMasterDerivationKey(pw)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("MDK:", base64.StdEncoding.EncodeToString(mdk[:]))
	mdkMnemonic, err := mnemonic.FromMasterDerivationKey(mdk)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("MDK mnemonic:", mdkMnemonic)

	// Generate account using wallet MDK
	log.Println("Generating new account using MDK...")
	newMdkPk, err := newWallet.GenerateKey(false)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("New MDK account address:", types.Address(newMdkPk).String())

	// DANGEROUS: Show mnemonic for new MDK account
	newMdkSk, err := newWallet.ExportKey(newMdkPk, pw)
	if err != nil {
		log.Fatal(err)
		return
	}
	mdkSkMnemonic, err := mnemonic.FromPrivateKey(newMdkSk)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("New MDK account mnemonic:", mdkSkMnemonic)

	// Generate standalone account
	log.Println("Generating new standalone account...")
	standaloneAcct := crypto.GenerateAccount()
	log.Println("New standalone account address:", standaloneAcct.Address.String())

	// DANGEROUS: Show mnemonic for new standalone account
	standaloneSkMnemonic, err := mnemonic.FromPrivateKey(standaloneAcct.PrivateKey)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("New standalone account mnemonic:", standaloneSkMnemonic)

	// Import generated standalone account
	log.Println("Importing new standalone account...")
	_, err3 := newWallet.ImportKey(standaloneAcct.PrivateKey)
	if err3 != nil {
		log.Fatal(err3)
		return
	}
	log.Println(standaloneAcct.Address, "has been imported.")

	// List accounts (1st time)
	pks, err := newWallet.ListKeys()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Accounts in wallet (1):")
	for _, pk := range pks {
		log.Println(types.Address(pk).String())
	}

	// Remove new MDK account
	log.Println("Attempting to remove MDK account...")
	err2 = newWallet.DeleteKey(newMdkPk, pw)
	if err2 != nil {
		log.Fatal(err2)
		return
	}
	log.Println(types.Address(newMdkPk).String(), "has been removed.")

	// List accounts (2nd time)
	pks, err = newWallet.ListKeys()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Accounts in wallet (2):")
	if len(pks) == 0 {
		log.Println("<None>")
	} else {
		for _, pk := range pks {
			log.Println(types.Address(pk).String())
		}
	}

	// Remove new standalone account
	log.Println("Attempting to remove standalone account...")
	err2 = newWallet.DeleteKey(types.Digest(standaloneAcct.PublicKey), pw)
	if err2 != nil {
		log.Fatal(err2)
		return
	}
	log.Println(standaloneAcct.Address, "has been removed.")

	// List accounts (3rd time)
	pks, err = newWallet.ListKeys()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Accounts in wallet (3):")
	if len(pks) == 0 {
		log.Println("<None>")
	} else {
		for _, pk := range pks {
			log.Println(types.Address(pk).String())
		}
	}
	// Rename wallet
	log.Println("Renaming wallet...")
	err = sqliteDriver.RenameWallet([]byte("insecure test"), walletID, pw)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Display data about newly created wallet (again)
	walletMetadata, _ = newWallet.Metadata()
	walletMetadataJson, _ = json.MarshalIndent(walletMetadata, "", "  ")
	log.Println(string(walletMetadataJson))

	// ====================================================================== //

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
