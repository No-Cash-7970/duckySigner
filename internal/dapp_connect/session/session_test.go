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

	Describe("Session.ID()", func() {
		It("returns the session ID", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			sessionId := sessionKey.PublicKey()

			By("Creating a session using generated session key pair")
			testSession := session.New(sessionKey, nil, time.Time{}, time.Time{}, nil, nil)

			Expect(testSession.ID()).To(Equal(sessionId))
		})
	})

	Describe("Session.Key()", func() {
		It("returns the session secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a session using session key")
			testSession := session.New(sessionKey, nil, time.Time{}, time.Time{}, nil, nil)

			Expect(testSession.Key()).To(Equal(sessionKey))
		})
	})

	Describe("Session.DappId()", func() {
		It("returns the dApp ID", func() {
			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a session using generated dApp ID")
			testSession := session.New(nil, dappId, time.Time{}, time.Time{}, nil, nil)

			Expect(testSession.DappId()).To(Equal(dappId))
		})
	})

	Describe("Session.Expiration()", func() {
		It("returns the expiration date-time", func() {
			testTime := time.Now()
			testSession := session.New(nil, nil, testTime, time.Time{}, nil, nil)
			Expect(testSession.Expiration()).To(Equal(testTime))
		})
	})

	Describe("Session.EstablishedAt()", func() {
		It("returns the establishment date-time", func() {
			testTime := time.Now()
			testSession := session.New(nil, nil, time.Time{}, testTime, nil, nil)
			Expect(testSession.EstablishedAt()).To(Equal(testTime))
		})
	})

	Describe("Session.DappData()", func() {
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
			testSession := session.New(nil, nil, time.Time{}, time.Time{}, &dappData, nil)

			By("Checking dApp data within session")
			Expect(testSession.DappData().Name).To(Equal("Foo Bar"))
			Expect(testSession.DappData().URL).To(Equal("https://example.com"))
			Expect(testSession.DappData().Description).To(Equal("This is an example"))
			Expect(testSession.DappData().Icon).To(Equal(dAppIconUri))
		})
	})

	Describe("Session.Addresses()", func() {
		It("returns the list of connect addresses", func() {
			testSession := session.New(nil, nil, time.Time{}, time.Time{}, nil, []string{
				"RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A",
				"H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A",
				"V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU",
			})
			Expect(testSession.Addresses()).To(HaveLen(3))
		})
	})

	Describe("Session.SharedKey()", func() {
		It("returns the session shared secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			sessionId := sessionKey.PublicKey()

			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a session using generated session key pair and dApp ID")
			testSession := session.New(sessionKey, dappId, time.Time{}, time.Time{}, nil, nil)

			By("Deriving shared key using dApp key and session ID")
			sharedKey, err := dappKey.ECDH(sessionId)
			Expect(err).ToNot(HaveOccurred())

			Expect(testSession.SharedKey()).To(Equal(sharedKey))
		})
	})
})
