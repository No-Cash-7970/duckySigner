package handlers_test

import (
	"encoding/base64"
	"net/http"
	"time"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"

	"github.com/hiyosi/hawk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The trailing slash is needed
const SessionEndGetUri = "http://localhost:" + sessionEndGetPort + "/"

var _ = Describe("GET /session/end", Ordered, func() {
	// Pre-generated keys for dApp connect session
	const (
		dappIdB64 = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		// dAppKeyB64 = "5zYnEKdGIcQSakSTwd21ZEygbX3mQ4vqV8WMZavvBb8="
		// sessionIdB64  = "dNoKnxinOqUNKQIbSTn5nk/pTjOtVznlXV5+MaWSH3k="
		sessionKeyB64 = "OA7vIBYGze5Vapw/qO3iPr+F9nRnaxsWSVnViTEZ1Ag="
		// dcKeyB64      = "I2y18jGyyNf4KTRrDtWyt09Qw2gppt5KHMJqm+gb9jY="
	)
	var sessionManager *session.Manager
	var testSession *session.Session
	var hawkHeader string

	BeforeAll(func() {
		By("Setting up dApp connect server")
		setUpDcService(sessionEndGetPort, sessionKeyB64)

		By("Creating a session")
		dappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
		Expect(err).NotTo(HaveOccurred())
		dappPk, err := curve.NewPublicKey(dappIdBytes)
		Expect(err).NotTo(HaveOccurred())
		sessionManager = session.NewManager(curve, &session.SessionConfig{
			DataDir: kmdService.Session().FilePath,
		})
		testSession, err = sessionManager.GenerateSession(dappPk, &dc.DappData{Name: "Foobar"}, nil)
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		err = sessionManager.StoreSession(testSession, mek)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with success when ending valid session", func() {
		By("Creating Hawk request header")
		sessionSharedKey, err := testSession.SharedKey()
		Expect(err).NotTo(HaveOccurred())
		nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
		Expect(err).NotTo(HaveOccurred())
		hawkClient := hawk.NewClient(
			&hawk.Credential{
				ID:  base64.StdEncoding.EncodeToString(testSession.ID().Bytes()),
				Key: base64.StdEncoding.EncodeToString(sessionSharedKey),
				Alg: hawk.SHA256,
			},
			&hawk.Option{TimeStamp: time.Now().Unix(), Nonce: nonce},
		)
		Expect(err).NotTo(HaveOccurred())
		hawkHeader, err := hawkClient.Header("GET", SessionEndGetUri)
		Expect(err).NotTo(HaveOccurred())

		By("Making an authenticated request")
		req, err := http.NewRequest("GET", SessionEndGetUri, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", hawkHeader)
		resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
		Expect(err).NotTo(HaveOccurred())

		By("Processing and checking response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(Equal(`"OK"` + "\n"))

		By("Checking if session exists")
		sessionId := base64.StdEncoding.EncodeToString(testSession.ID().Bytes())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		nilSession, err := sessionManager.GetSession(sessionId, mek)
		Expect(nilSession).To(BeNil())
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with success with ending session that does not exist", func() {
		By("Making an authenticated request")
		req, err := http.NewRequest("GET", SessionEndGetUri, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", hawkHeader)
		resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
		Expect(err).NotTo(HaveOccurred())

		By("Processing and checking response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(Equal(`"OK"` + "\n"))
	})

	It("fails if request is not authenticated", func() {
		By("Making an unauthenticated request")
		req, err := http.NewRequest("GET", SessionEndGetUri, nil)
		Expect(err).NotTo(HaveOccurred())
		resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
		Expect(err).NotTo(HaveOccurred())

		By("Checking response status code")
		Expect(resp.StatusCode).To(Equal(401))
	})
})
