package services_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wailsapp/wails/v3/pkg/application"

	. "duckysigner/services"
)

var _ = Describe("DappConnectService", func() {
	Describe("Start()", func() {
		It("starts the server", func() {
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
				HideServerPort:   true,
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
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
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
				HideServerPort:   true,
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
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
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
		var dcService DappConnectService
		const defaultUserRespTimeout = 2 * time.Second

		BeforeAll(func() {
			dcService = DappConnectService{
				// Make sure to use a port that is not used in another test so the
				// tests can be run in parallel
				ServerAddr:       ":1384",
				LogLevel:         log.ERROR,
				HideServerBanner: true,
				HideServerPort:   true,
				WailsApp: application.New(application.Options{
					LogLevel: slog.LevelError,
				}),
				UserResponseTimeout: defaultUserRespTimeout,
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

				By("Processing response")
				body, err := getResponseBody(resp)
				Expect(err).NotTo(HaveOccurred())

				By("Checking response")
				Expect(string(body)).To(Equal("\"OK\"\n"))
			})
		})

		Describe("POST /session/init", func() {
			It("creates a new wallet connection session", func() {
				// Signals that the request has yielded a response
				var respSignal = make(chan string)
				go func() {
					By("Making a request with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name": "foo"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					By("Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo"}]`))

					By("Wallet user: Approving session connection")
					dcService.WailsApp.EmitEvent("session_init_response", []string{"account 1", "account 2"})
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking response contains new set of Hawk credentials")
				Expect(respBody).To(Equal("\"Hawk credentials\"\n"))
			})

			It("fails when user does not respond", func() {
				// Shorten the timeout for this test
				dcService.UserResponseTimeout = 100 * time.Millisecond
				DeferCleanup(func() {
					dcService.UserResponseTimeout = defaultUserRespTimeout
				})

				// Signals that the request has yielded a response
				var respSignal = make(chan string)
				go func() {
					By("Making a request with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name": "foo"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					By("Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo"}]`))

					By("Wallet user: Not responding...")
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking response contains error message")
				expected, _ := json.Marshal(ApiError{"session_no_response", "User did not respond"})
				Expect(respBody).To(Equal(string(expected) + "\n"))
			})

			It("fails when user rejects the session", func() {
				// Signals that the request has yielded a response
				var respSignal = make(chan string)
				go func() {
					By("Making a request with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name": "foo"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					By("Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo"}]`))

					By("Wallet user: Rejecting session connection")
					dcService.WailsApp.EmitEvent("session_init_response", []string{})
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking response contains new set of Hawk credentials")
				expected, _ := json.Marshal(ApiError{"session_rejected", "Session was rejected"})
				Expect(respBody).To(Equal(string(expected) + "\n"))
			})
		})
	})
})

func getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
