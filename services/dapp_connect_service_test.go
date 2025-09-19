package services_test

import (
	"github.com/awnumar/memguard"
	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"duckysigner/internal/kmd/config"
	. "duckysigner/services"
)

var _ = Describe("DappConnectService", func() {
	Describe("DappConnectService.Start()", func() {
		It("starts the server", func() {
			const walletDirName = ".test_dcs_wallets_start"
			kmdService := createKmdServiceForDCS(walletDirName)
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1324",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				KMDService:       kmdService,
			}
			DeferCleanup(func() {
				dcService.Stop()
				createKmdServiceCleanup(walletDirName)
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
		})

		It("can handle attempt to start server when it is already running", func() {
			const walletDirName = ".test_dcs_wallets_start_run"
			kmdService := createKmdServiceForDCS(walletDirName)
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1325",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				KMDService:       kmdService,
			}
			DeferCleanup(func() {
				dcService.Stop()
				createKmdServiceCleanup(walletDirName)
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to start server again while it is running")
			Expect(dcService.Start()).To(Equal(true))
		})
	})

	Describe("DappConnectService.Stop()", func() {
		It("stops the server if it running", func() {
			const walletDirName = ".test_dcs_wallets_stop"
			kmdService := createKmdServiceForDCS(walletDirName)
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1344",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				KMDService:       kmdService,
			}
			DeferCleanup(func() {
				createKmdServiceCleanup(walletDirName)
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop server")
			Expect(dcService.Stop()).To(Equal(false))
		})

		It("can handle attempt to stop server that is not running", func() {
			const walletDirName = ".test_dcs_wallets_stop_run"
			kmdService := createKmdServiceForDCS(walletDirName)
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1345",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				KMDService:       kmdService,
			}
			DeferCleanup(func() {
				createKmdServiceCleanup(walletDirName)
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop server")
			Expect(dcService.Stop()).To(Equal(false))
			By("Attempting to stop server again while it is not running")
			Expect(dcService.Stop()).To(Equal(false))
		})
	})

	Describe("DappConnectService.IsOn()", func() {
		It("shows if the server is currently on", func() {
			const walletDirName = ".test_dcs_wallets_is_on"
			kmdService := createKmdServiceForDCS(walletDirName)
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1364",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				KMDService:       kmdService,
			}
			DeferCleanup(func() {
				createKmdServiceCleanup(walletDirName)
			})

			By("Starting server")
			Expect(dcService.Start()).To(Equal(true))
			By("Running IsOn() to check if the server is running")
			Expect(dcService.IsOn()).To(Equal(true))
			By("Stopping server")
			Expect(dcService.Stop()).To(Equal(false))
			By("Running IsOn() to check if the server is not running")
			Expect(dcService.IsOn()).To(Equal(false))
		})
	})
})

// createKmdService is a helper function that returns a new KMDService that is
// to be used for dApp connect service tests
func createKmdServiceForDCS(walletDirName string) *KMDService {
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
				SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{UnsafeScrypt: true, Disable: true},
				LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
			},
		},
	}
	// Create a wallet
	_, err := kmdService.CreateWallet("DApp Connect Test Wallet", "test password")
	Expect(err).ToNot(HaveOccurred())

	return &kmdService
}
