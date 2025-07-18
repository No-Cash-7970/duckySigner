package session_test

import (
	"crypto/ecdh"
	"crypto/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("DApp Connect Confirmation", func() {

	Describe("ID()", func() {
		It("returns the confirmation ID", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			confirmId := confirmKey.PublicKey()

			By("Creating a confirmation using generated confirmation key pair")
			confirm := session.NewConfirmation(confirmKey)

			Expect(confirm.ID()).To(Equal(confirmId))
		})
	})

	Describe("Key()", func() {
		It("returns the confirmation secret key", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a confirmation using confirmation key")
			confirm := session.NewConfirmation(confirmKey)

			Expect(confirm.Key()).To(Equal(confirmKey))
		})
	})

	Describe("SharedKey()", func() {
		It("Returns the confirmation shared secret key", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			confirmId := confirmKey.PublicKey()

			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a confirmation using generated confirmation key pair and dApp ID")
			confirm := session.NewConfirmation(confirmKey)

			By("Deriving shared key using dApp key and confirmation ID")
			sharedKey, err := dappKey.ECDH(confirmId)
			Expect(err).ToNot(HaveOccurred())

			Expect(confirm.SharedKey(dappId)).To(Equal(sharedKey))
		})
	})
})
