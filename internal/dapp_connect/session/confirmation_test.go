package session_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"time"

	"aidanwoods.dev/go-paseto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("DApp Connect Confirmation", func() {

	Describe("Confirmation.DappId()", func() {
		It("returns the dApp ID", func() {
			By("Generating a dApp key pair (dApp ID & key)")
			dappKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()

			By("Creating a confirmation")
			confirmToken := session.NewConfirmation(dappId, nil, nil, "", time.Time{})

			Expect(confirmToken.DappId()).To(Equal(dappId))
		})
	})

	Describe("Confirmation.SessionKey()", func() {
		It("returns the session secret key", func() {
			By("Generating a session key pair (session ID & key)")
			sessionKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a confirmation")
			confirmToken := session.NewConfirmation(nil, sessionKey, nil, "", time.Time{})

			Expect(confirmToken.SessionKey()).To(Equal(sessionKey))
		})
	})

	Describe("Confirmation.ID()", func() {
		It("returns the confirmation ID", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			confirmId := confirmKey.PublicKey()

			By("Creating a confirmation")
			confirmToken := session.NewConfirmation(nil, nil, confirmKey, "", time.Time{})

			Expect(confirmToken.ID()).To(Equal(confirmId))
		})
	})

	Describe("Confirmation.Key()", func() {
		It("returns the confirmation secret key", func() {
			By("Generating a confirmation key pair (confirmation ID & key)")
			confirmKey, err := ecdh.X25519().GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a confirmation")
			confirmToken := session.NewConfirmation(nil, nil, confirmKey, "", time.Time{})

			Expect(confirmToken.Key()).To(Equal(confirmKey))
		})
	})

	Describe("Confirmation.Code()", func() {
		It("returns the confirmation ID", func() {
			confirmCode := "123456"
			confirmToken := session.NewConfirmation(nil, nil, nil, confirmCode, time.Time{})
			Expect(confirmToken.Code()).To(Equal(confirmCode))
		})
	})

	Describe("Confirmation.Expiration()", func() {
		It("returns the expiration date-time", func() {
			testTime := time.Now()
			confirmToken := session.NewConfirmation(nil, nil, nil, "", testTime)
			Expect(confirmToken.Expiration()).To(Equal(testTime))
		})
	})

	Describe("Confirmation.SharedKey()", func() {
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
			confirm := session.NewConfirmation(dappId, nil, confirmKey, "", time.Time{})

			By("Deriving shared key using dApp key and confirmation ID")
			sharedKey, err := dappKey.ECDH(confirmId)
			Expect(err).ToNot(HaveOccurred())

			Expect(confirm.SharedKey()).To(Equal(sharedKey))
		})
	})

	Describe("Confirmation.GenerateTokenString()", func() {
		It("returns token string", func() {
			curve := ecdh.X25519()
			// Create dApp ID
			dappKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()
			// Create session key
			sessionKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create confirmation key
			confirmKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := "123456"
			exp := time.Now().Add(5 * time.Minute)

			By("Creating an encrypted confirmation token string")
			confirmToken := session.NewConfirmation(dappId, sessionKey, confirmKey, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			Expect(err).ToNot(HaveOccurred())

			By("Decrypting token string")
			pasetoKey, err := paseto.V4SymmetricKeyFromBytes(confirmKey.Bytes())
			Expect(err).ToNot(HaveOccurred())
			parser := paseto.NewParser()
			parsedToken, err := parser.ParseV4Local(pasetoKey, confirmTokenString, nil)
			Expect(err).ToNot(HaveOccurred())

			By("Checking 'claims' in decrypted token")
			// Expiration
			parsedExp, err := parsedToken.GetExpiration()
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedExp).To(BeTemporally("~", exp, time.Second), "Decrypted token contains expiry")
			// DApp ID
			parsedDappId, err := parsedToken.GetString("dapp")
			Expect(err).ToNot(HaveOccurred())
			dappIdB64 := base64.StdEncoding.EncodeToString(dappId.Bytes())
			Expect(parsedDappId).To(Equal(dappIdB64), "Decrypted token contains dApp ID")
			// Confirmation code
			parsedCode, err := parsedToken.GetString("code")
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedCode).To(Equal(confirmCode), "Decrypted token contains confirmation code")
			// Session key
			parsedSKey, err := parsedToken.GetString("skey")
			Expect(err).ToNot(HaveOccurred())
			sessionKeyB64 := base64.StdEncoding.EncodeToString(sessionKey.Bytes())
			Expect(parsedSKey).To(Equal(sessionKeyB64), "Decrypted token contains session key")
		})

		It("fails when dApp ID is missing", func() {
			curve := ecdh.X25519()
			// Create session key
			sessionKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create confirmation key
			confirmKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := "123456"
			exp := time.Now().Add(5 * time.Minute)

			By("Attempting to create an encrypted confirmation token string without a dApp ID")
			confirmToken := session.NewConfirmation(nil, sessionKey, confirmKey, confirmCode, exp)
			_, err = confirmToken.GenerateTokenString()
			Expect(err).To(MatchError(session.MissingConfirmTokenDappIdErrMsg))
		})

		It("fails when session key is missing", func() {
			curve := ecdh.X25519()
			// Create dApp ID
			dappKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()
			// Create confirmation key
			confirmKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := "123456"
			exp := time.Now().Add(5 * time.Minute)

			By("Attempting to create an encrypted confirmation token string without a session key")
			confirmToken := session.NewConfirmation(dappId, nil, confirmKey, confirmCode, exp)
			_, err = confirmToken.GenerateTokenString()
			Expect(err).To(MatchError(session.MissingConfirmTokenSessionKeyErrMsg))
		})

		It("fails when confirmation code is missing", func() {
			curve := ecdh.X25519()
			// Create dApp ID
			dappKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()
			// Create session key
			sessionKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create confirmation key
			confirmKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := ""
			exp := time.Now().Add(5 * time.Minute)

			By("Attempting to create an encrypted confirmation token string without a confirmation code")
			confirmToken := session.NewConfirmation(dappId, sessionKey, confirmKey, confirmCode, exp)
			_, err = confirmToken.GenerateTokenString()
			Expect(err).To(MatchError(session.MissingConfirmTokenCodeErrMsg))
		})

		It("fails when confirmation key is missing", func() {
			curve := ecdh.X25519()
			// Create dApp ID
			dappKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()
			// Create session key
			sessionKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := "123456"
			exp := time.Now().Add(5 * time.Minute)

			By("Attempting to create an encrypted confirmation token string without a confirmation key")
			confirmToken := session.NewConfirmation(dappId, sessionKey, nil, confirmCode, exp)
			_, err = confirmToken.GenerateTokenString()
			Expect(err).To(MatchError(session.MissingConfirmTokenConfirmKeyErrMsg))
		})
	})

	Describe("DecryptToken()", func() {
		It("decrypts given confirmation token string", func() {
			curve := ecdh.X25519()
			// Create dApp ID
			dappKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			dappId := dappKey.PublicKey()
			// Create session key
			sessionKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create confirmation key
			confirmKey, err := curve.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			// Create other confirmation token data
			confirmCode := "123456"
			exp := time.Now().Add(5 * time.Minute)

			By("Creating a confirmation encrypted token string")
			token := paseto.NewToken()
			token.SetExpiration(exp)
			token.SetString(session.DappIdClaimName, base64.StdEncoding.EncodeToString(dappId.Bytes()))
			token.SetString(session.ConfirmCodeClaimName, confirmCode)
			token.SetString(session.SessionKeyClaimName, base64.StdEncoding.EncodeToString(sessionKey.Bytes()))
			pasetoKey, err := paseto.V4SymmetricKeyFromBytes(confirmKey.Bytes())
			Expect(err).ToNot(HaveOccurred())
			encryptedTokenString := token.V4Encrypt(pasetoKey, nil)

			By("Decrypting the encrypted confirmation token")
			decryptedToken, err := session.DecryptToken(encryptedTokenString, confirmKey, curve)
			Expect(err).ToNot(HaveOccurred())

			By("Checking 'claims' in decrypted token")
			// Expiration
			tokenExp := decryptedToken.Expiration()
			Expect(tokenExp).To(BeTemporally("~", exp, time.Second), "Decrypted token contains expiry")
			// DApp ID
			tokenDappId := decryptedToken.DappId()
			Expect(tokenDappId).To(Equal(dappId), "Decrypted token contains dApp ID")
			// Confirmation code
			tokenCode := decryptedToken.Code()
			Expect(tokenCode).To(Equal(confirmCode), "Decrypted token contains confirmation code")
			// Session key
			tokenSKey := decryptedToken.SessionKey()
			Expect(tokenSKey).To(Equal(sessionKey), "Decrypted token contains session key")
		})
	})
})
