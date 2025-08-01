package session_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("DApp Connect Session", func() {

	Describe("ID()", func() {
		It("returns the session ID", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			sessionId := sessionKey.PublicKey()

			By("Creating a session using generated session key pair")
			session := session.New(sessionKey, nil, time.Time{}, time.Time{}, nil)

			Expect(session.ID()).To(Equal(sessionId))
		})
	})

	Describe("Key()", func() {
		It("returns the session secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a session using session key")
			session := session.New(sessionKey, nil, time.Time{}, time.Time{}, nil)

			Expect(session.Key()).To(Equal(sessionKey))
		})
	})

	Describe("DappId()", func() {
		It("returns the dApp ID", func() {
			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a session using generated dApp ID")
			session := session.New(nil, dappId, time.Time{}, time.Time{}, nil)

			Expect(session.DappId()).To(Equal(dappId))
		})
	})

	Describe("Expiration()", func() {
		It("returns the expiration date-time", func() {
			testTime := time.Now()
			session := session.New(nil, nil, testTime, time.Time{}, nil)
			Expect(session.Expiration()).To(Equal(testTime))
		})
	})

	Describe("EstablishedAt()", func() {
		It("returns the establishment date-time", func() {
			testTime := time.Now()
			session := session.New(nil, nil, time.Time{}, testTime, nil)
			Expect(session.EstablishedAt()).To(Equal(testTime))
		})
	})

	Describe("DappData()", func() {
		It("returns the dApp data", func() {
			By("Creating some dApp data")
			dAppIconUri := "data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 50 50'%3e%3cpath d='M22 38V51L32 32l19-19v12C44 26 43 10 38 0 52 15 49 39 22 38z'/%3e%3c/svg%3e"
			dappData := dc.DappData{
				Name:        "Foo Bar",
				URL:         "https://example.com",
				Description: "This is an example",
				Icon:        dAppIconUri,
			}

			By("Creating a session using dApp data")
			session := session.New(nil, nil, time.Time{}, time.Time{}, &dappData)

			By("Checking dApp data within session")
			Expect(session.DappData().Name).To(Equal("Foo Bar"))
			Expect(session.DappData().URL).To(Equal("https://example.com"))
			Expect(session.DappData().Description).To(Equal("This is an example"))
			Expect(session.DappData().Icon).To(Equal(dAppIconUri))
		})
	})

	Describe("SharedKey()", func() {
		It("Returns the session shared secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			sessionId := sessionKey.PublicKey()

			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a session using generated session key pair and dApp ID")
			session := session.New(sessionKey, dappId, time.Time{}, time.Time{}, nil)

			By("Deriving shared key using dApp key and session ID")
			sharedKey, err := dappKey.ECDH(sessionId)
			Expect(err).ToNot(HaveOccurred())

			Expect(session.SharedKey()).To(Equal(sharedKey))
		})
	})
})
