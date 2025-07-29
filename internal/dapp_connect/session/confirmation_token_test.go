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

	Describe("GenerateTokenString()", func() {

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

			By("Using GenerateTokenString() to create a confirmation encrypted token string")
			confirmToken := session.NewConfirmationToken(dappId, sessionKey, confirmKey, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			GinkgoWriter.Println(confirmTokenString)
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

			By("Attempting to use GenerateTokenString() to create a confirmation encrypted token string")
			confirmToken := session.NewConfirmationToken(nil, sessionKey, confirmKey, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			GinkgoWriter.Println(confirmTokenString)
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

			By("Attempting to use GenerateTokenString() to create a confirmation encrypted token string")
			confirmToken := session.NewConfirmationToken(dappId, nil, confirmKey, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			GinkgoWriter.Println(confirmTokenString)
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

			By("Attempting to use GenerateTokenString() to create a confirmation encrypted token string")
			confirmToken := session.NewConfirmationToken(dappId, sessionKey, confirmKey, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			GinkgoWriter.Println(confirmTokenString)
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

			By("Attempting to use GenerateTokenString() to create a confirmation encrypted token string")
			confirmToken := session.NewConfirmationToken(dappId, sessionKey, nil, confirmCode, exp)
			confirmTokenString, err := confirmToken.GenerateTokenString()
			GinkgoWriter.Println(confirmTokenString)
			Expect(err).To(MatchError(session.MissingConfirmTokenConfirmKeyErrMsg))
		})
	})

	Describe("DecryptConfirmationToken()", func() {
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

			By("Using DecryptConfirmationToken() to decrypt an encrypted confirmation token")
			decryptedToken, err := session.DecryptConfirmationToken(encryptedTokenString, confirmKey, curve)
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
