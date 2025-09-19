package handlers_test

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/awnumar/memguard"
	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/wailsapp/wails/v3/pkg/application"

	"duckysigner/internal/kmd/config"
	"duckysigner/internal/testing/mocks"
	. "duckysigner/services"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DApp Connect Endpoints Suite")
}

func getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

var dcService DappConnectService

const defaultUserRespTimeout = 2 * time.Second
const rootGetPort = "1384"
const sessionInitPostPort = "1385"

func setUpDcService(port string, mockSessionKey string) {
	walletDirName := ".test_dc_handlers_" + port
	kmdService := createKmdService(walletDirName)
	dcService = DappConnectService{
		// Make sure to use a port that is not used in another test so the
		// tests can be run in parallel
		ServerAddr:       ":" + port,
		LogLevel:         log.ERROR,
		HideServerBanner: true,
		HideServerPort:   true,
		WailsApp: application.New(application.Options{
			LogLevel: slog.LevelError,
		}),
		UserResponseTimeout: defaultUserRespTimeout,
		ECDHCurve:           &mocks.EcdhCurveMock{GeneratedPrivateKey: mockSessionKey},
		KMDService:          kmdService,
	}
	By("Starting dApp connect server")
	dcService.Start()
	DeferCleanup(func() {
		By("Stopping dApp connect server")
		dcService.Stop()
		By("Cleaning up `" + walletDirName + "` wallet")
		createKmdServiceCleanup(walletDirName)
		kmdService.CleanUp()
	})
}

// createKmdService is a helper function that returns a new KMDService
func createKmdService(walletDirName string) *KMDService {
	// Clean up by will wipe sensitive data if process is terminated suddenly
	memguard.CatchInterrupt()

	// Create the service
	kmdService := KMDService{
		Config: config.KMDConfig{
			SessionLifetimeSecs: 3600,
			DriverConfig: config.DriverConfig{
				ParquetWalletDriverConfig: config.ParquetWalletDriverConfig{
					WalletsDir:   walletDirName,
					UnsafeScrypt: true, // For testing purposes only
					ScryptParams: config.ScryptParams{
						ScryptN: 2,
						ScryptR: 1,
						ScryptP: 1,
					},
				},
				SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{
					UnsafeScrypt: true,
					Disable:      true,
					WalletsDir:   walletDirName,
				},
				LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
			},
		},
	}

	const walletPassword = "test password"

	// Create a wallet
	walletMetadata, err := kmdService.CreateWallet("DApp Connect Test Wallet", walletPassword)
	Expect(err).ToNot(HaveOccurred())

	// Start wallet session
	err = kmdService.StartSession(string(walletMetadata.ID), walletPassword)
	Expect(err).ToNot(HaveOccurred())

	return &kmdService
}

// createKmdServiceCleanup is a helper function that cleans up the directory
// with specified by walletDirName, which was created for an instance of the
// KMDService
func createKmdServiceCleanup(walletDirName string) {
	// Remove test wallet directory
	err := os.RemoveAll(walletDirName)
	if !os.IsNotExist(err) { // If no "directory does not exist" error
		Expect(err).NotTo(HaveOccurred())
	}
}
