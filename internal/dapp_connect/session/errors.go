package session

const (
	// SessionExistsErrMsg is the error message text for when a session already
	// exists
	SessionExistsErrMsg = "session already exists"
	// NoSessionGivenErrMsg is the error message text for when no session is
	// provided
	NoSessionGivenErrMsg = "no session was given"
	// NoSessionKeyGivenErrMsg is the error message text for when no session key
	// is given
	NoSessionKeyGivenErrMsg = "no session key was given"
	// NoDappIdGivenErrMsg is the error message text for when no dApp ID is
	// given
	NoDappIdGivenErrMsg = "no dApp ID was given"
	// RemoveSessionNotExistErrMsg is the error message text for when there is
	// an attempt to remove a session that is not stored
	RemoveSessionNotStoredErrMsg = "cannot remove session that is not stored"
	// WrongConfirmCodeErrMsg is the error message text for when the given
	// confirmation code does not match the code that is with in the given
	// confirmation token
	WrongConfirmCodeErrMsg = "wrong confirmation code"
	// NoConfirmTokenGivenErrMsg is the error message text for when the given
	// confirmation token is empty (i.e. no token was given)
	NoConfirmTokenGivenErrMsg = "no confirmation token given"
	// ConfirmKeyExistsErrMsg is the error message text for when a confirmation
	// key already exists
	ConfirmKeyExistsErrMsg = "confirmation key already exists"
	// NoConfirmKeyGivenErrMsg is the error message text for when no
	// confirmation key is provided
	NoConfirmKeyGivenErrMsg = "no confirmation key was given"
	// RemoveConfirmKeyNotExistErrMsg is the error message text for when there is
	// an attempt to remove a confirmation key that is not stored
	RemoveConfirmKeyNotStoredErrMsg = "cannot remove confirmation key that is not stored"

	// MissingConfirmTokenDappIdErrMsg is the error message for when the dApp ID
	// is missing within the confirmation token
	MissingConfirmTokenDappIdErrMsg = "missing dApp ID in confirmation token"
	// MissingConfirmTokenSessionKeyErrMsg is the error message for when the
	// session key is missing within the confirmation token
	MissingConfirmTokenSessionKeyErrMsg = "missing session key in confirmation token"
	// MissingConfirmTokenConfirmKeyErrMsg is the error message for when the
	// confirmation key is missing within the confirmation token
	MissingConfirmTokenConfirmKeyErrMsg = "missing confirmation key in confirmation token"
	// MissingConfirmTokenCodeErrMsg is the error message for when the
	// confirmation code is missing within the confirmation token
	MissingConfirmTokenCodeErrMsg = "missing confirmation code in confirmation token"
)
