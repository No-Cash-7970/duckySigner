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

		By("Checking if server responds with validation error data")
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

		By("Checking if server responds with error data")
		var respData ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("bad_dapp_id"))
	})
})
