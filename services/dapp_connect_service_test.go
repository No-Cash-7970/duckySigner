package services_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "duckysigner/services"
)

var _ = Describe("DappConnectService", func() {
	Describe("Start()", func() {
		It("starts the server", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1324",
				LogLevel:         log.WARN,
				HideServerBanner: true,
			}
			DeferCleanup(func() {
				dcService.Stop()
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
		})

		It("can handle attempt to start server when it is already running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so
				// the tests can be run in parallel
				ServerAddr:       ":1325",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
			}
			DeferCleanup(func() {
				dcService.Stop()
			})

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to start server again while it is running")
			Expect(dcService.Start()).To(Equal(true))
		})
	})

	Describe("Stop()", func() {
		It("stops the server if it running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1344",
				LogLevel:         log.WARN,
				HideServerBanner: true,
			}

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop server")
			Expect(dcService.Stop()).To(Equal(false))
		})

		It("can handle attempt to stop server that is not running", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1345",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
			}

			By("Attempting to start server")
			Expect(dcService.Start()).To(Equal(true))
			By("Attempting to stop server")
			Expect(dcService.Stop()).To(Equal(false))
			By("Attempting to stop server again while it is not running")
			Expect(dcService.Stop()).To(Equal(false))
		})
	})

	Describe("IsOn()", func() {
		It("shows if the server is currently on", func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1364",
				LogLevel:         log.WARN,
				HideServerBanner: true,
			}

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

	Describe("Server routes", Ordered, func() {
		BeforeAll(func() {
			dcService := DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1384",
				LogLevel:         log.WARN,
				HideServerBanner: true,
			}
			By("Starting server")
			dcService.Start()
			DeferCleanup(func() {
				By("Stopping server")
				dcService.Stop()
			})
		})

		Describe("GET /", func() {
			It("works", func() {
				By("Making request")
				resp, err := http.Get("http://localhost:1384/")
				Expect(err).NotTo(HaveOccurred())
				body, err := getResponseBody(resp)
				Expect(err).NotTo(HaveOccurred())

				By("Checking response")
				expected, _ := json.Marshal("OK")
				Expect(string(body)).To(Equal(string(expected) + "\n"))
			})
		})
	})
})

func getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
