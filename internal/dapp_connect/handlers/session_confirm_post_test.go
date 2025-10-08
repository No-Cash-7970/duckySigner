package handlers_test

import (
	"bytes"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hiyosi/hawk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/handlers"
	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("POST /session/confirm", Ordered, func() {
	// Pre-generated keys for dApp connect session
	const (
		dappIdB64 = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		// dAppKeyB64 = "5zYnEKdGIcQSakSTwd21ZEygbX3mQ4vqV8WMZavvBb8="
		// sessionIdB64  = "dNoKnxinOqUNKQIbSTn5nk/pTjOtVznlXV5+MaWSH3k="
		sessionKeyB64 = "OA7vIBYGze5Vapw/qO3iPr+F9nRnaxsWSVnViTEZ1Ag="
		// dcKeyB64      = "I2y18jGyyNf4KTRrDtWyt09Qw2gppt5KHMJqm+gb9jY="
	)
	var dappPk *ecdh.PublicKey
	var sessionManager *session.Manager
	var testConfirm *session.Confirmation
	const uri = "http://localhost:" + sessionConfirmPostPort + "/session/confirm"

	BeforeAll(func() {
		dappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
		Expect(err).NotTo(HaveOccurred())
		dappPk, err = curve.NewPublicKey(dappIdBytes)
		Expect(err).NotTo(HaveOccurred())

		setUpDcService(sessionConfirmPostPort, sessionKeyB64)

		sessionManager = session.NewManager(curve, &session.SessionConfig{
			DataDir: kmdService.Session().FilePath,
		})
	})

	It("confirms session and responds with session data", func() {
		By("Creating and storing a session confirmation")
		var err error
		testConfirm, err = sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(testConfirm.Key(), mek)
		token, err := testConfirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := testConfirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(testConfirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Mock UI/user response to prompt event emitted from server
		dcService.WailsApp.Event.On(handlers.SessionConfirmPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve session connection")
			Expect(fmt.Sprint(e.Data)).To(Equal(`[{"dapp":{"name":"foo"}}]`))
			By("Wallet user: Approving session connection")
			dcService.WailsApp.Event.Emit(
				"session_confirm_response",
				`{"code":"`+testConfirm.Code()+`","addrs":["account 1","account 2"]}`,
			)
		})
		DeferCleanup(func() {
			dcService.WailsApp.Event.Off(handlers.SessionConfirmPromptEventName)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with session data")
		var respData handlers.SessionConfirmPostResp
		json.Unmarshal(respBody, &respData)
		Expect(respData.Id).To(HaveLen(44), "Session ID is within response")
		Expect(respData.Expiration).To(BeNumerically(">", time.Now().Unix()),
			"Expiry is within response")
	})

	It("fails when attempting to confirm a session again", func() {
		// NOTE: Because this `Describe` container is "Ordered", a particular
		// session is assumed to have been established

		token, err := testConfirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := testConfirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(testConfirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Mock UI/user response to prompt event emitted from server
		dcService.WailsApp.Event.On(handlers.SessionConfirmPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve session connection")
			Expect(fmt.Sprint(e.Data)).To(Equal(`[{"dapp":{"name":"foo"}}]`))
			By("Wallet user: Approving session connection")
			dcService.WailsApp.Event.Emit(
				"session_confirm_response",
				`{"code":"`+testConfirm.Code()+`","addrs":["account 1","account 2"]}`,
			)
		})
		DeferCleanup(func() {
			dcService.WailsApp.Event.Off(handlers.SessionConfirmPromptEventName)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("session_create_fail"))
	})

	It("fails when no confirmation token is given", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)

		var reqBody = `{"token":"","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
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
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("validation_error"))
	})

	It("fails when given an invalid confirmation token", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)

		var reqBody = `{"token":"invalid token","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
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
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("confirm_auth_request_failed"))
	})

	It("fails when confirmation token has expired", func() {
		By("Creating and storing a session confirmation")
		generatedConfirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		// Create a confirmation that is expired
		confirm := session.NewConfirmation(
			dappPk,
			generatedConfirm.SessionKey(),
			generatedConfirm.Key(),
			generatedConfirm.Code(),
			time.Now().Add(-1*time.Minute), // Expired a minute ago
		)
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)
		token, err := confirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
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
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("confirm_expired"))
	})

	It("fails when no dApp name is given", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)
		token, err := confirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":""}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
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
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("validation_error"))
	})

	It("fails when user does not respond", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)
		token, err := confirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Mock UI/user response to prompt event emitted from server
		dcService.WailsApp.Event.On(handlers.SessionConfirmPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve session connection")
			Expect(fmt.Sprint(e.Data)).To(Equal(`[{"dapp":{"name":"foo"}}]`))
			By("Wallet user: Not responding...")
		})
		DeferCleanup(func() {
			dcService.WailsApp.Event.Off(handlers.SessionConfirmPromptEventName)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error data")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("confirm_timeout"))
	})

	It("fails when user rejects the session", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)
		token, err := confirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Mock UI/user response to prompt event emitted from server
		dcService.WailsApp.Event.On(handlers.SessionConfirmPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve session connection")
			Expect(fmt.Sprint(e.Data)).To(Equal(`[{"dapp":{"name":"foo"}}]`))
			By("Wallet user: Approving session connection")
			dcService.WailsApp.Event.Emit("session_confirm_response", `{"code":"","addrs":[]}`)
		})
		DeferCleanup(func() {
			dcService.WailsApp.Event.Off(handlers.SessionConfirmPromptEventName)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error data")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("session_rejected"))
	})

	It("fails when user does not provide correct confirmation code", func() {
		By("Creating and storing a session confirmation")
		confirm, err := sessionManager.GenerateConfirmation(dappPk)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		sessionManager.StoreConfirmKey(confirm.Key(), mek)
		token, err := confirm.GenerateTokenString()
		Expect(err).NotTo(HaveOccurred())

		var reqBody = `{"token":"` + token + `","dapp":{"name":"foo"}}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			By("Creating Hawk request header")
			confirmSharedKey, err := confirm.SharedKey()
			Expect(err).NotTo(HaveOccurred())
			nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
			Expect(err).NotTo(HaveOccurred())
			hawkClient := hawk.NewClient(
				&hawk.Credential{
					ID:  base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
					Key: base64.StdEncoding.EncodeToString(confirmSharedKey),
					Alg: hawk.SHA256,
				},
				&hawk.Option{
					TimeStamp:   time.Now().Unix(),
					Payload:     reqBody,
					ContentType: "application/json",
					Nonce:       nonce,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			hawkHeader, err := hawkClient.Header("POST", uri)
			Expect(err).NotTo(HaveOccurred())

			By("Making an authenticated request to server with valid dApp data")
			req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", hawkHeader)
			resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("Processing response from server")
			body, err := getResponseBody(resp)
			Expect(err).NotTo(HaveOccurred())
			// Signal that request has completed
			respSignal <- body
			close(respSignal)
		}()

		// Mock UI/user response to prompt event emitted from server
		dcService.WailsApp.Event.On(handlers.SessionConfirmPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve session connection")
			Expect(fmt.Sprint(e.Data)).To(Equal(`[{"dapp":{"name":"foo"}}]`))
			By("Wallet user: Approving session connection")
			dcService.WailsApp.Event.Emit(
				"session_confirm_response",
				`{"code":"0000","addrs":["account 1","account 2"]}`,
			)
		})
		DeferCleanup(func() {
			dcService.WailsApp.Event.Off(handlers.SessionConfirmPromptEventName)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error data")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("wrong_confirm_code"))
	})
})
