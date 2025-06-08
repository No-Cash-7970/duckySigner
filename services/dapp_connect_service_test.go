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

	. "duckysigner/internal/dapp_connect"
	"duckysigner/internal/testing/mocks"
	. "duckysigner/services"

	"github.com/wailsapp/wails/v3/pkg/application"
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

	Describe("Connection server routes", Ordered, func() {
		var dcService DappConnectService
		const defaultUserRespTimeout = 2 * time.Second
		// Pre-generated keys for wallet connection session
		const dAppId = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		const dAppKey = "5zYnEKdGIcQSakSTwd21ZEygbX3mQ4vqV8WMZavvBb8="
		const sessionId = "dNoKnxinOqUNKQIbSTn5nk/pTjOtVznlXV5+MaWSH3k="
		const sessionKey = "OA7vIBYGze5Vapw/qO3iPr+F9nRnaxsWSVnViTEZ1Ag="
		const wcKey = "I2y18jGyyNf4KTRrDtWyt09Qw2gppt5KHMJqm+gb9jY="

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
				ECDHCurve:           &mocks.EcdhCurveMock{GeneratedPrivateKey: sessionKey},
			}
			By("Starting connection server")
			dcService.Start()
			DeferCleanup(func() {
				By("Stopping connection server")
				dcService.Stop()
			})
		})

		Describe("GET /", func() {
			It("works", func() {
				By("Making request to connection server")
				resp, err := http.Get("http://localhost:1384/")
				Expect(err).NotTo(HaveOccurred())

				By("Processing response from connection server")
				body, err := getResponseBody(resp)
				Expect(err).NotTo(HaveOccurred())

				By("Checking response from connection server")
				Expect(string(body)).To(Equal(`"OK"` + "\n"))
			})
		})

		Describe("POST /session/init", func() {
			It("creates a new wallet connection session", func() {
				// Signal for when the request has yielded a response
				var respSignal = make(chan string)

				go func() {
					defer GinkgoRecover()

					By("Making a request to connection server with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name":"foo","dapp_session_pk":"`+dAppId+`"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response from connection server")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				// Mock UI/user response to prompt event emitted from server
				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					defer GinkgoRecover()

					By("UI: Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_session_pk":"` + dAppId + `"}]`))

					By("Wallet user: Approving session connection")
					dcService.WailsApp.EmitEvent("session_init_response", []string{"account 1", "account 2"})
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to parse & check the response
				respBody := <-respSignal

				By("Checking if connection server responds with new set of Hawk credentials")
				expectedResp, _ := json.Marshal(HawkCredentials{
					Algorithm: "sha256",
					ID:        dAppId,
					Key:       sessionId,
				})
				Expect(respBody).To(Equal(string(expectedResp) + "\n"))

				// TODO: Somehow check that server and client have same shared key (?)
			})

			It("fails when given a dApp ID that is not valid Base64", func() {
				// Signals that the request has yielded a response
				var respSignal = make(chan string)

				go func() {
					defer GinkgoRecover()

					By("Making a request with a dApp ID that is not valid Base64")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name":"foo","dapp_session_pk":"hello world"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response from UI")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				// Mock UI/user response to prompt event emitted from server
				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					defer GinkgoRecover()
					Fail("Connection server should not have emitted event for prompting UI/user")
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking if connection server responds with error")
				expected, _ := json.Marshal(ApiError{
					Name:    "invalid_dapp_id_b64",
					Message: "DApp ID is not a valid Base64 string",
				})
				Expect(respBody).To(Equal(string(expected) + "\n"))
			})

			It("fails when given a dApp ID that is not a valid ECDH public key", func() {
				// Signals that the request has yielded a response
				var respSignal = make(chan string)

				go func() {
					defer GinkgoRecover()

					By("Making a request to connection server with a dApp ID that is not valid ECDH public key")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name":"foo","dapp_session_pk":"aGVsbG8gd29ybGQ="}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response from connection server")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				// Mock UI/user response to prompt event emitted from server
				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					defer GinkgoRecover()
					Fail("Connection server should not have emitted event for prompting UI/user")
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking if connection server responds with error")
				expected, _ := json.Marshal(ApiError{
					Name:    "invalid_dapp_id_pk",
					Message: "DApp ID is invalid",
				})
				Expect(respBody).To(Equal(string(expected) + "\n"))
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
					defer GinkgoRecover()

					By("Making a request to connection server with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name":"foo","dapp_session_pk":"`+dAppId+`"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response from connection server")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				// Mock UI/user response to prompt event emitted from server
				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					defer GinkgoRecover()

					By("UI: Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_session_pk":"` + dAppId + `"}]`))

					By("Wallet user: Not responding...")
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking if connection server responds with error")
				expected, _ := json.Marshal(ApiError{
					Name:    "session_no_response",
					Message: "User did not respond",
				})
				Expect(respBody).To(Equal(string(expected) + "\n"))
			})

			It("fails when user rejects the session", func() {
				// Signals that the request has yielded a response
				var respSignal = make(chan string)

				go func() {
					defer GinkgoRecover()

					By("Making a request to connection server with valid dApp data")
					resp, err := http.Post(
						"http://localhost:1384/session/init",
						"application/json",
						bytes.NewReader([]byte(`{"name":"foo","dapp_session_pk":"`+dAppId+`"}`)),
					)
					Expect(err).NotTo(HaveOccurred())

					By("Processing response from connection server")
					body, err := getResponseBody(resp)
					Expect(err).NotTo(HaveOccurred())
					// Signal that request has completed
					respSignal <- string(body)
					close(respSignal)
				}()

				// Mock UI/user response to prompt event emitted from server
				dcService.WailsApp.OnEvent("session_init_prompt", func(e *application.CustomEvent) {
					defer GinkgoRecover()

					By("UI: Prompting user to approve session connection")
					eventData, err := json.Marshal(e.Data)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_session_pk":"` + dAppId + `"}]`))

					By("Wallet user: Rejecting session connection")
					dcService.WailsApp.EmitEvent("session_init_response", []string{})
				})
				DeferCleanup(func() {
					dcService.WailsApp.OffEvent("session_init_prompt")
				})

				// Wait for request to complete before trying to check the response
				respBody := <-respSignal

				By("Checking if connection server responds with error")
				expected, _ := json.Marshal(ApiError{
					Name:    "session_rejected",
					Message: "Session was rejected",
				})
				Expect(respBody).To(Equal(string(expected) + "\n"))
			})
		})
	})
})

func getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
