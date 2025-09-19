package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/handlers"
)

var _ = Describe("POST /session/init", Ordered, func() {
	// Pre-generated keys for dApp connect session
	const (
		dAppId = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		// dAppKey    = "5zYnEKdGIcQSakSTwd21ZEygbX3mQ4vqV8WMZavvBb8="
		sessionId  = "dNoKnxinOqUNKQIbSTn5nk/pTjOtVznlXV5+MaWSH3k="
		sessionKey = "OA7vIBYGze5Vapw/qO3iPr+F9nRnaxsWSVnViTEZ1Ag="
		// dcKey      = "I2y18jGyyNf4KTRrDtWyt09Qw2gppt5KHMJqm+gb9jY="
	)

	BeforeAll(func() {
		setUpDcService(sessionInitPostPort, sessionKey)
	})

	It("responds with session confirmation data", func() {
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)

		go func() {
			defer GinkgoRecover()

			By("Making a request to server with valid dApp data")
			resp, err := http.Post(
				"http://localhost:"+sessionInitPostPort+"/session/init",
				"application/json",
				bytes.NewReader([]byte(`{"dapp_id":"`+dAppId+`"}`)),
			)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with confirmation data")
		var respData handlers.SessionInitPostResp
		json.Unmarshal(respBody, &respData)
		Expect(respData.Id).To(HaveLen(44))
		Expect(respData.Code).To(HaveLen(5))
		Expect(respData.Token).ToNot(BeEmpty())
		Expect(respData.Expiration).To(BeNumerically(">", time.Now().Unix()))
	})

	It("fails when given a dApp ID that is not valid Base64", func() {
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)

		go func() {
			defer GinkgoRecover()

			By("Making a request to server with valid dApp data")
			resp, err := http.Post(
				"http://localhost:"+sessionInitPostPort+"/session/init",
				"application/json",
				bytes.NewReader([]byte(`{"dapp_id":"hello world"}`)),
			)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with confirmation data")
		var respData ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("validation_error"))
	})

	It("fails when given a dApp ID that is not a valid ECDH public key", func() {
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)

		go func() {
			defer GinkgoRecover()

			By("Making a request to server with valid dApp data")
			resp, err := http.Post(
				"http://localhost:"+sessionInitPostPort+"/session/init",
				"application/json",
				bytes.NewReader([]byte(`{"dapp_id":"aGVsbG8gd29ybGQ="}`)),
			)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with confirmation data")
		var respData ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("bad_dapp_id"))
	})

	// It("creates a new dApp connect session", func() {
	// 	// Signal for when the request has yielded a response
	// 	var respSignal = make(chan string)

	// 	go func() {
	// 		defer GinkgoRecover()

	// 		By("Making a request to server with valid dApp data")
	// 		resp, err := http.Post(
	// 			"http://localhost:"+sessionInitPostPort+"/session/init",
	// 			"application/json",
	// 			bytes.NewReader([]byte(`{"name":"foo","dapp_id":"`+dAppId+`"}`)),
	// 		)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Processing response from server")
	// 		body, err := getResponseBody(resp)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// Signal that request has completed
	// 		respSignal <- string(body)
	// 		close(respSignal)
	// 	}()

	// 	// Mock UI/user response to prompt event emitted from server
	// 	dcService.WailsApp.Event.On("session_init_prompt", func(e *application.CustomEvent) {
	// 		defer GinkgoRecover()

	// 		By("UI: Prompting user to approve session connection")
	// 		eventData, err := json.Marshal(e.Data)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_id":"` + dAppId + `"}]`))

	// 		By("Wallet user: Approving session connection")
	// 		dcService.WailsApp.Event.Emit("session_init_response", []string{"account 1", "account 2"})
	// 	})
	// 	DeferCleanup(func() {
	// 		dcService.WailsApp.Event.Off("session_init_prompt")
	// 	})

	// 	// Wait for request to complete before trying to parse & check the response
	// 	respBody := <-respSignal

	// 	By("Checking if server responds with new set of Hawk credentials")
	// 	expectedResp, _ := json.Marshal(HawkCredentials{
	// 		Algorithm: "sha256",
	// 		ID:        dAppId,
	// 		Key:       sessionId,
	// 	})
	// 	Expect(respBody).To(Equal(string(expectedResp) + "\n"))

	// 	// TODO: Somehow check that server and client have same shared key (?)
	// })

	// It("fails when given a dApp ID that is not valid Base64", func() {
	// 	// Signals that the request has yielded a response
	// 	var respSignal = make(chan string)

	// 	go func() {
	// 		defer GinkgoRecover()

	// 		By("Making a request with a dApp ID that is not valid Base64")
	// 		resp, err := http.Post(
	// 			"http://localhost:"+sessionInitPostPort+"/session/init",
	// 			"application/json",
	// 			bytes.NewReader([]byte(`{"name":"foo","dapp_id":"hello world"}`)),
	// 		)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Processing response from UI")
	// 		body, err := getResponseBody(resp)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// Signal that request has completed
	// 		respSignal <- string(body)
	// 		close(respSignal)
	// 	}()

	// 	// Mock UI/user response to prompt event emitted from server
	// 	dcService.WailsApp.Event.On("session_init_prompt", func(e *application.CustomEvent) {
	// 		defer GinkgoRecover()
	// 		Fail("Connection server should not have emitted event for prompting UI/user")
	// 	})
	// 	DeferCleanup(func() {
	// 		dcService.WailsApp.Event.Off("session_init_prompt")
	// 	})

	// 	// Wait for request to complete before trying to check the response
	// 	respBody := <-respSignal

	// 	By("Checking if server responds with error")
	// 	expected, _ := json.Marshal(ApiError{
	// 		Name:    "invalid_dapp_id_b64",
	// 		Message: "DApp ID is not a valid Base64 string",
	// 	})
	// 	Expect(respBody).To(Equal(string(expected) + "\n"))
	// })

	// It("fails when given a dApp ID that is not a valid ECDH public key", func() {
	// 	// Signals that the request has yielded a response
	// 	var respSignal = make(chan string)

	// 	go func() {
	// 		defer GinkgoRecover()

	// 		By("Making a request to server with a dApp ID that is not valid ECDH public key")
	// 		resp, err := http.Post(
	// 			"http://localhost:"+sessionInitPostPort+"/session/init",
	// 			"application/json",
	// 			bytes.NewReader([]byte(`{"name":"foo","dapp_id":"aGVsbG8gd29ybGQ="}`)),
	// 		)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Processing response from server")
	// 		body, err := getResponseBody(resp)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// Signal that request has completed
	// 		respSignal <- string(body)
	// 		close(respSignal)
	// 	}()

	// 	// Mock UI/user response to prompt event emitted from server
	// 	dcService.WailsApp.Event.On("session_init_prompt", func(e *application.CustomEvent) {
	// 		defer GinkgoRecover()
	// 		Fail("Connection server should not have emitted event for prompting UI/user")
	// 	})
	// 	DeferCleanup(func() {
	// 		dcService.WailsApp.Event.Off("session_init_prompt")
	// 	})

	// 	// Wait for request to complete before trying to check the response
	// 	respBody := <-respSignal

	// 	By("Checking if server responds with error")
	// 	expected, _ := json.Marshal(ApiError{
	// 		Name:    "invalid_dapp_id_pk",
	// 		Message: "DApp ID is invalid",
	// 	})
	// 	Expect(respBody).To(Equal(string(expected) + "\n"))
	// })

	// It("fails when user does not respond", func() {
	// 	// Shorten the timeout for this test
	// 	dcService.UserResponseTimeout = 100 * time.Millisecond
	// 	DeferCleanup(func() {
	// 		dcService.UserResponseTimeout = defaultUserRespTimeout
	// 	})

	// 	// Signals that the request has yielded a response
	// 	var respSignal = make(chan string)

	// 	go func() {
	// 		defer GinkgoRecover()

	// 		By("Making a request to server with valid dApp data")
	// 		resp, err := http.Post(
	// 			"http://localhost:"+sessionInitPostPort+"/session/init",
	// 			"application/json",
	// 			bytes.NewReader([]byte(`{"name":"foo","dapp_id":"`+dAppId+`"}`)),
	// 		)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Processing response from server")
	// 		body, err := getResponseBody(resp)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// Signal that request has completed
	// 		respSignal <- string(body)
	// 		close(respSignal)
	// 	}()

	// 	// Mock UI/user response to prompt event emitted from server
	// 	dcService.WailsApp.Event.On("session_init_prompt", func(e *application.CustomEvent) {
	// 		defer GinkgoRecover()

	// 		By("UI: Prompting user to approve session connection")
	// 		eventData, err := json.Marshal(e.Data)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_id":"` + dAppId + `"}]`))

	// 		By("Wallet user: Not responding...")
	// 	})
	// 	DeferCleanup(func() {
	// 		dcService.WailsApp.Event.Off("session_init_prompt")
	// 	})

	// 	// Wait for request to complete before trying to check the response
	// 	respBody := <-respSignal

	// 	By("Checking if server responds with error")
	// 	expected, _ := json.Marshal(ApiError{
	// 		Name:    "session_no_response",
	// 		Message: "User did not respond",
	// 	})
	// 	Expect(respBody).To(Equal(string(expected) + "\n"))
	// })

	// It("fails when user rejects the session", func() {
	// 	// Signals that the request has yielded a response
	// 	var respSignal = make(chan string)

	// 	go func() {
	// 		defer GinkgoRecover()

	// 		By("Making a request to server with valid dApp data")
	// 		resp, err := http.Post(
	// 			"http://localhost:"+sessionInitPostPort+"/session/init",
	// 			"application/json",
	// 			bytes.NewReader([]byte(`{"name":"foo","dapp_id":"`+dAppId+`"}`)),
	// 		)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Processing response from server")
	// 		body, err := getResponseBody(resp)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// Signal that request has completed
	// 		respSignal <- string(body)
	// 		close(respSignal)
	// 	}()

	// 	// Mock UI/user response to prompt event emitted from server
	// 	dcService.WailsApp.Event.On("session_init_prompt", func(e *application.CustomEvent) {
	// 		defer GinkgoRecover()

	// 		By("UI: Prompting user to approve session connection")
	// 		eventData, err := json.Marshal(e.Data)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(string(eventData)).To(Equal(`[{"name":"foo","dapp_id":"` + dAppId + `"}]`))

	// 		By("Wallet user: Rejecting session connection")
	// 		dcService.WailsApp.Event.Emit("session_init_response", []string{})
	// 	})
	// 	DeferCleanup(func() {
	// 		dcService.WailsApp.Event.Off("session_init_prompt")
	// 	})

	// 	// Wait for request to complete before trying to check the response
	// 	respBody := <-respSignal

	// 	By("Checking if server responds with error")
	// 	expected, _ := json.Marshal(ApiError{
	// 		Name:    "session_rejected",
	// 		Message: "Session was rejected",
	// 	})
	// 	Expect(respBody).To(Equal(string(expected) + "\n"))
	// })
})
