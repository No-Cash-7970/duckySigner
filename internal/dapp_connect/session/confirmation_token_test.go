package session_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("DApp Connect Confirmation Token", func() {

	Describe("DappId()", func() {
		It("returns the dApp ID", func() {
			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a confirmation token")
			confirmToken := session.NewConfirmationToken(dappId, nil, nil, "", time.Time{})

			Expect(confirmToken.DappId()).To(Equal(dappId))
		})
	})

	Describe("SessionKey()", func() {
		It("returns the session secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a confirmation token")
			confirmToken := session.NewConfirmationToken(nil, sessionKey, nil, "", time.Time{})

			Expect(confirmToken.SessionKey()).To(Equal(sessionKey))
		})
	})

	Describe("ConfirmationId()", func() {
		It("returns the confirmation ID", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			confirmId := confirmKey.PublicKey()

			By("Creating a confirmation token")
			confirmToken := session.NewConfirmationToken(nil, nil, confirmKey, "", time.Time{})

			Expect(confirmToken.ConfirmationId()).To(Equal(confirmId))
		})
	})

	Describe("ConfirmationKey()", func() {
		It("returns the confirmation secret key", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a confirmation token")
			confirmToken := session.NewConfirmationToken(nil, nil, confirmKey, "", time.Time{})

			Expect(confirmToken.ConfirmationKey()).To(Equal(confirmKey))
		})
	})

	Describe("Code()", func() {
		It("returns the confirmation ID", func() {
			confirmCode := "123456"
			confirmToken := session.NewConfirmationToken(nil, nil, nil, confirmCode, time.Time{})
			Expect(confirmToken.Code()).To(Equal(confirmCode))
		})
	})

	Describe("Expiration()", func() {
		It("returns the expiration date-time", func() {
			testTime := time.Now()
			confirmToken := session.NewConfirmationToken(nil, nil, nil, "", testTime)
			Expect(confirmToken.Expiration()).To(Equal(testTime))
		})
	})

	PDescribe("ToEncryptedString()", func() {
		It("returns token string", func() {
			// TODO: Complete this
		})

		It("fails when dApp ID is missing", func() {
			// TODO: Complete this
		})

		It("fails when session key is missing", func() {
			// TODO: Complete this
		})

		It("fails when confirmation key is missing", func() {
			// TODO: Complete this
		})

		It("fails when confirmation code is missing", func() {
			// TODO: Complete this
		})
	})

	PDescribe("DecryptConfirmationToken()", func() {
		It("decrypts given confirmation token string", func() {
			// TODO: Complete this
		})
	})
})
