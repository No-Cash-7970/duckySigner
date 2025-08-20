package session

import "os"

const (
	// dataDirPermission is the OS file permissions used for the data directory
	// when it is created
	dataDirPermissions = 0700
	// tempFileSuffix is the suffix used to create a temporary file. The
	// temporary file name is the original file name + this suffix.
	tempFileSuffix = ".new"
)

// removeTempFile attempts to remove the temporary file used when modifying the
// file with the given file name.
func removeTempFile(originalFilename string) error {
	// Check if temporary file exists
	_, err := os.Stat(originalFilename + tempFileSuffix)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		// Could not access file some reason other than that it does not exist
		// (e.g. permissions, drive failure)
		return err
	}

	// Remove the original file
	err = os.Remove(originalFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Rename the temporary file
	err = os.Rename(originalFilename+tempFileSuffix, originalFilename)
	if err != nil {
		return err
	}

	return nil
}
