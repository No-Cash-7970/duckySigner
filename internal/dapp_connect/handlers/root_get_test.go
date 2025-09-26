package handlers_test

import (
	"duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/hiyosi/hawk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GET /", Ordered, func() {
	BeforeAll(func() {
		setUpDcService(rootGetPort, "")
	})

	It("works", func() {
		By("Making request to server")
		resp, err := http.Get("http://localhost:" + rootGetPort)
		Expect(err).NotTo(HaveOccurred())

		By("Processing response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())

		By("Checking response from server")
		Expect(string(body)).To(Equal(`"OK"` + "\n"))
	})

	It("works with an authenticated request", func() {
		const uri = "http://localhost:" + rootGetPort + "/" // The trailing slash is needed

		By("By creating a session")
		const dappIdB64 = "c+2pz3JaUkIEMnbi1vuv7RWdGpfyiv6O3xaYbYbieAg="
		dappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
		Expect(err).NotTo(HaveOccurred())
		dappPk, err := curve.NewPublicKey(dappIdBytes)
		Expect(err).NotTo(HaveOccurred())
		sessionManager := session.NewManager(curve, &session.SessionConfig{
			DataDir: kmdService.Session().FilePath,
		})
		session, err := sessionManager.GenerateSession(dappPk, &dapp_connect.DappData{Name: "Foobar"})
		Expect(err).NotTo(HaveOccurred())
		mek, err := kmdService.Session().GetMasterKey()
		Expect(err).NotTo(HaveOccurred())
		err = sessionManager.StoreSession(session, mek)
		Expect(err).NotTo(HaveOccurred())

		By("Creating Hawk request header")
		sessionSharedKey, err := session.SharedKey()
		Expect(err).NotTo(HaveOccurred())
		nonce, err := hawk.Nonce(4) // Generate nonce that is 4 bytes long
		Expect(err).NotTo(HaveOccurred())
		hawkClient := hawk.NewClient(
			&hawk.Credential{
				ID:  base64.StdEncoding.EncodeToString(session.ID().Bytes()),
				Key: base64.StdEncoding.EncodeToString(sessionSharedKey),
				Alg: hawk.SHA256,
			},
			&hawk.Option{TimeStamp: time.Now().Unix(), Nonce: nonce},
		)
		Expect(err).NotTo(HaveOccurred())
		hawkHeader, err := hawkClient.Header("GET", uri)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println(hawkHeader)

		By("Making an authenticated request")
		req, err := http.NewRequest("GET", uri, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", hawkHeader)
		resp, err := (&http.Client{Timeout: 1 * time.Minute}).Do(req)
		Expect(err).NotTo(HaveOccurred())

		By("Processing response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())

		By("Checking response from server")
		Expect(string(body)).To(Equal(`"OK"` + "\n"))
	})
})
