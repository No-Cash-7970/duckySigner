package services_test

import (
	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "duckysigner/services"
)

var _ = Describe("DappConnectService", func() {
	Describe("Start()", func() {
		It("starts the connection server", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1324",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
			}
			DeferCleanup(func() {
				dcService.Stop()
			})

			By("Attempting to start connection server")
			Expect(dcService.Start()).To(Equal(true))
		})

		It("can handle attempt to start connection server when it is already running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1325",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
			}
			DeferCleanup(func() {
				dcService.Stop()
			})

			By("Attempting to start connection server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to start connection server again while it is running")
			Expect(dcService.Start()).To(Equal(true))
		})
	})

	Describe("Stop()", func() {
		It("stops the connection server if it running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1344",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
			}

			By("Attempting to start connection server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop connection server")
			Expect(dcService.Stop()).To(Equal(false))
		})

		It("can handle attempt to stop connection server that is not running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1345",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
			}

			By("Attempting to start connection server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop connection server")
			Expect(dcService.Stop()).To(Equal(false))
			By("Attempting to stop connection server again while it is not running")
			Expect(dcService.Stop()).To(Equal(false))
		})
	})

	Describe("IsOn()", func() {
		It("shows if the connection server is currently on", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1364",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
			}

			By("Starting connection server")
			Expect(dcService.Start()).To(Equal(true))
			By("Running IsOn() to check if the connection server is running")
			Expect(dcService.IsOn()).To(Equal(true))
			By("Stopping connection server")
			Expect(dcService.Stop()).To(Equal(false))
			By("Running IsOn() to check if the connection server is not running")
			Expect(dcService.IsOn()).To(Equal(false))
		})
	})
})
