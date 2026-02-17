package driver_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KMD Driver Suite")
}

// createKmdServiceCleanup is a helper function that cleans up the directory
// with specified by walletDirName, which was created for an instance of the
// KMDService
func createKmdServiceCleanup(walletDirName string) {
	// Remove test wallet directory
	err := os.RemoveAll(walletDirName)
	if !os.IsNotExist(err) { // If no "directory does not exist" error
		Expect(err).NotTo(HaveOccurred())
	}
}
