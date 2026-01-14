package handlers_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	algoTypes "github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/hiyosi/hawk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/handlers"
	"duckysigner/internal/dapp_connect/session"
)

const txnSignPostUri = "http://localhost:" + transactionSignPostPort + "/transaction/sign"

var _ = Describe("POST /transaction/sign", Ordered, func() {
	// Pre-generated keys for dApp connect session
	const (
		dappIdB64 = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		// dAppKeyB64 = "5zYnEKdGIcQSakSTwd21ZEygbX3mQ4vqV8WMZavvBb8="
		// sessionIdB64  = "dNoKnxinOqUNKQIbSTn5nk/pTjOtVznlXV5+MaWSH3k="
		sessionKeyB64 = "OA7vIBYGze5Vapw/qO3iPr+F9nRnaxsWSVnViTEZ1Ag="
		// dcKeyB64      = "I2y18jGyyNf4KTRrDtWyt09Qw2gppt5KHMJqm+gb9jY="
	)
	var testSession *session.Session
	var acctAddr string
	var testTxn algoTypes.Transaction
	var encodedTestTxn []byte

	BeforeAll(func() {
		var err error

		By("Setting up dApp connect server")
		setUpDcService(transactionSignPostPort, sessionKeyB64)

		By("Generating an account in the wallet")
		acctAddr, err = kmdService.Session().GenerateAccount()
		Expect(err).NotTo(HaveOccurred())

		By("Creating a session")
		dappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
		Expect(err).NotTo(HaveOccurred())
		dappPk, err := curve.NewPublicKey(dappIdBytes)
		Expect(err).NotTo(HaveOccurred())
		sessionManager := session.NewManager(curve, &session.SessionConfig{
			DataDir: kmdService.Session().FilePath,
		})
		testSession, err = sessionManager.GenerateSession(dappPk, &dc.DappData{Name: "Foobar"}, nil)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		err = sessionManager.StoreSession(testSession, mek)
		Expect(err).NotTo(HaveOccurred())

		By("Creating a test transaction")
		testTxn, err = transaction.MakePaymentTxn(acctAddr, acctAddr, 1000000, nil, "", algoTypes.SuggestedParams{
			Fee:             1000,
			GenesisID:       "testnet-v1.0",
			GenesisHash:     []byte("SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="),
			FirstRoundValid: 10000,
			LastRoundValid:  11000,
		})
		Expect(err).NotTo(HaveOccurred())
		encodedTestTxn = msgpack.Encode(testTxn)
	})

	AfterEach(func() {
		dcService.WailsApp.Event.Reset()
	})

	It("responds with signed transaction", func() {
		var reqBody = `{"transaction":"` + base64.StdEncoding.EncodeToString(encodedTestTxn) + `"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		dcService.WailsApp.Event.On(handlers.TxnSignPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve transaction")
			Expect(fmt.Sprint(e.Data)).To(Equal(`{"data":` + reqBody + `}`))
			By("Wallet user: Approving transaction")
			dcService.WailsApp.Event.Emit(handlers.TxnSignRespEventName, `{"approved":true}`)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with signed transaction data")
		var respData handlers.TransactionSignPostResp
		json.Unmarshal(respBody, &respData)
		Expect(respData.SignedTxn).ToNot(BeEmpty())
	})

	It("fails if request is not authenticated", func() {
		var reqBody = `{"transaction":"` + base64.StdEncoding.EncodeToString(encodedTestTxn) + `"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()

			By("Making an UNAUTHENTICATED request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
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

		By("Checking if server responds with error")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("auth_request_failed"))
	})

	It("fails if no transaction is given", func() {
		var reqBody = `{"transaction":""}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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

	It("fails if invalid transaction is given", func() {
		var reqBody = `{"transaction":"hello world"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		Expect(respData.Name).To(Equal("invalid_txn"))
	})

	It("fails if invalid signer is given", func() {
		var reqBody = `{"transaction":"` +
			base64.StdEncoding.EncodeToString(encodedTestTxn) +
			`","signer":"hello"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		Expect(respData.Name).To(Equal("invalid_signer"))
	})

	It("fails if signer account not in wallet", func() {
		var reqBody = `{"transaction":"` +
			base64.StdEncoding.EncodeToString(encodedTestTxn) +
			`","signer":"EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		Expect(respData.Name).To(Equal("invalid_signer"))
	})

	It("fails if transaction's sender is not in wallet (when signer is not specified)", func() {
		By("Creating a test transaction with a sender that (probably) does not exist in wallet")
		txn, err := transaction.MakePaymentTxn(
			"EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4",
			"EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4",
			1000000, nil, "",
			algoTypes.SuggestedParams{
				Fee:             1000,
				GenesisID:       "testnet-v1.0",
				GenesisHash:     []byte("SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="),
				FirstRoundValid: 10000,
				LastRoundValid:  11000,
			},
		)
		Expect(err).NotTo(HaveOccurred())
		encodedTxn := msgpack.Encode(txn)

		var reqBody = `{"transaction":"` + base64.StdEncoding.EncodeToString(encodedTxn) + `"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		Expect(respData.Name).To(Equal("invalid_sender"))
	})

	It("fails when user does not respond", func() {
		var reqBody = `{"transaction":"` + base64.StdEncoding.EncodeToString(encodedTestTxn) + `"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		dcService.WailsApp.Event.On(handlers.TxnSignPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve transaction")
			Expect(fmt.Sprint(e.Data)).To(Equal(`{"data":` + reqBody + `}`))
			By("Wallet user: Not responding...")
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error data")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("txn_sign_timeout"))
	})

	It("fails when user rejects signing the transaction", func() {
		var reqBody = `{"transaction":"` + base64.StdEncoding.EncodeToString(encodedTestTxn) + `"}`
		// Signal for when the request has yielded a response
		var respSignal = make(chan []byte)
		go func() {
			defer GinkgoRecover()
			hawkHeader := CreateTransactionSignPostReqHawkHeader(testSession, reqBody)

			By("Making an authenticated request to server with valid data")
			req, err := http.NewRequest("POST", txnSignPostUri, bytes.NewReader([]byte(reqBody)))
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
		dcService.WailsApp.Event.On(handlers.TxnSignPromptEventName, func(e *application.CustomEvent) {
			defer GinkgoRecover()
			By("UI: Prompting user to approve transaction")
			Expect(fmt.Sprint(e.Data)).To(Equal(`{"data":` + reqBody + `}`))
			By("Wallet user: Approving transaction")
			dcService.WailsApp.Event.Emit(handlers.TxnSignRespEventName, `{"approved":false}`)
		})

		// Wait for request to complete before trying to parse & check the response
		respBody := <-respSignal

		By("Checking if server responds with error data")
		var respData dc.ApiError
		json.Unmarshal(respBody, &respData)
		Expect(respData.Name).To(Equal("txn_sign_rejected"))
	})
})

// CreateTransactionSignPostReqHawkHeader creates a new Hawk authentication
// header for a `/transaction/sign` request
func CreateTransactionSignPostReqHawkHeader(sess *session.Session, reqBody string) (hawkHeader string) {
	By("Creating Hawk request header")
	sessionSharedKey, err := sess.SharedKey()
	Expect(err).NotTo(HaveOccurred())
	nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
	Expect(err).NotTo(HaveOccurred())
	hawkClient := hawk.NewClient(
		&hawk.Credential{
			ID:  base64.StdEncoding.EncodeToString(sess.ID().Bytes()),
			Key: base64.StdEncoding.EncodeToString(sessionSharedKey),
			Alg: hawk.SHA256,
		},
		&hawk.Option{
			TimeStamp:   time.Now().Unix(),
			Nonce:       nonce,
			Payload:     reqBody,
			ContentType: "application/json",
		},
	)
	Expect(err).NotTo(HaveOccurred())
	hawkHeader, err = hawkClient.Header("POST", txnSignPostUri)
	Expect(err).NotTo(HaveOccurred())
	return
}
