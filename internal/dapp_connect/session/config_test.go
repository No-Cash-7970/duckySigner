package session_test

import (
	"crypto/rand"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"duckysigner/internal/dapp_connect/session"
)

var _ = Describe("Session Configuration", func() {

	Describe("SessionConfig.ToFile()", Ordered, func() {
		var encryptKey [32]byte
		var fileName = ".test_session_config_to_file"

		BeforeAll(func() {
			rand.Read(encryptKey[:]) // Generate file encryption key
			DeferCleanup(sessionConfigCleanup(fileName))
		})

		It("creates a session configuration file with defaults when session configuration is empty", func() {
			var sessionConfig session.SessionConfig
			sessionConfig.ToFile(fileName, encryptKey[:])

			fileContent, err := os.ReadFile(fileName)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContent)).To(ContainSubstring("v4.local"),
				"File contains configuration Paseto")
		})

		It("creates a session configuration file when session configuration is not empty (and file exists)", func() {
			// NOTE: Because this `Describe` container is "Ordered", the session
			// configuration file is assumed to have been created

			oldFileContent, err := os.ReadFile(fileName)
			Expect(err).ToNot(HaveOccurred())

			sessionConfig := session.SessionConfig{
				SessionsFile:        "foo_sessions.parquet",
				ConfirmsFile:        "bar_confirms.parquet",
				DataDir:             "foobar",
				SessionLifetimeSecs: 42,
				ConfirmLifetimeSecs: 8,
			}
			sessionConfig.ToFile(fileName, encryptKey[:])

			newFileContent, err := os.ReadFile(fileName)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(newFileContent)).NotTo(Equal(string(oldFileContent)),
				"File is overwritten to contain new configuration Paseto")
		})
	})

	Describe("ConfigFromFile()", Ordered, func() {
		var encryptKey [32]byte
		var fileName = ".test_session_config_load"

		BeforeAll(func() {
			rand.Read(encryptKey[:]) // Generate file encryption key
			DeferCleanup(sessionConfigCleanup(fileName))
		})

		It("fails when it tries to load session configuration file that does not exist", func() {
			// NOTE: Because this `Describe` container is "Ordered", the session
			// configuration file is assumed to not have been created
			By("Attempting to load configuration file that does not exist")
			_, err := session.ConfigFromFile(fileName, encryptKey[:])
			Expect(err).To(MatchError(os.ErrNotExist))
		})

		It("loads session configuration file", func() {
			By("Creating a session configuration file")
			sessionConfig := session.SessionConfig{
				SessionsFile:        "foo_sessions.parquet",
				ConfirmsFile:        "bar_confirms.parquet",
				DataDir:             "foobar",
				SessionLifetimeSecs: 42,
				ConfirmLifetimeSecs: 8,
			}
			sessionConfig.ToFile(fileName, encryptKey[:])

			By("Loading configuration file")
			loadedConfig, err := session.ConfigFromFile(fileName, encryptKey[:])
			Expect(err).ToNot(HaveOccurred())

			By("Checking loaded configuration")
			Expect(loadedConfig.SessionsFile).To(Equal(sessionConfig.SessionsFile),
				"Loads sessions file setting in configuration")
			Expect(loadedConfig.ConfirmsFile).To(Equal(sessionConfig.ConfirmsFile),
				"Loads confirmation keystore file setting in configuration")
			Expect(loadedConfig.DataDir).To(Equal(sessionConfig.DataDir),
				"Loads data directory setting in configuration")
			Expect(loadedConfig.SessionLifetimeSecs).To(Equal(sessionConfig.SessionLifetimeSecs),
				"Loads session lifetime setting in configuration")
			Expect(loadedConfig.ConfirmLifetimeSecs).To(Equal(sessionConfig.ConfirmLifetimeSecs),
				"Loads confirmation lifetime setting in configuration")
			Expect(loadedConfig.ConfirmCodeCharset).To(Equal(session.DefaultConfirmCodeCharset),
				"Loads (default) confirmation code character set setting in configuration")
			Expect(loadedConfig.ConfirmCodeLen).To(Equal(uint(session.DefaultConfirmCodeLen)),
				"Loads (default) confirmation code length setting in configuration")
		})
	})
})

// sessionCleanup returns a helper function that removes the file with the
// specified name
func sessionConfigCleanup(fileName string) func() {
	return func() {
		// Remove test file
		err := os.RemoveAll(fileName)
		if !os.IsNotExist(err) { // If no "does not exist" error
			Expect(err).NotTo(HaveOccurred())
		}
	}
}
