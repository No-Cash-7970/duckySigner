package services_test

import (
	"duckysigner/internal/kmd/config"
	. "duckysigner/services"
	"encoding/base64"
	"os"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/awnumar/memguard"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KmdService", func() {
	Context("using SQLite", func() {
		// XXX: I shouldn't have to tell you this, but don't use the following
		// mnemonics for anything other than testing purposes
		const testWalletMnemonic = "increase document mandate absorb chapter valve apple amazing pipe hope impact travel away comfort two desk business robust brand sudden vintage scheme valve above inmate"
		const testStandaloneAcctAddr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
		const testStandaloneAcctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

		// NOTE: When creating or renaming a wallet, make sure the wallet name
		// is not a name used for wallet in a different test. Two or more tests
		// using the same wallet name causes errors when running tests in
		// parallel.

		It("can manage wallets", func() {
			By("Initializing KMD")
			const walletDirName = ".test_wallet_mngmt"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(func() {
				kmdService.CleanUp()
				createKmdServiceCleanup(walletDirName)
			})

			By("Listing 0 wallets")
			walletsInfo, err := kmdService.ListWallets()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletsInfo).To(HaveLen(0), "There should be no wallets to list")

			By("Creating a new wallet")
			newWalletInfoA, err := kmdService.CreateWallet("wallet_A", "passwordA")
			Expect(err).NotTo(HaveOccurred())
			Expect(newWalletInfoA.DriverName).To(Equal("sqlite"))
			Expect(newWalletInfoA.Name).To(BeEquivalentTo("wallet_A"), "1st wallet should have been created")

			By("Not giving error if a given wallet password is correct")
			err = kmdService.CheckWalletPassword(string(newWalletInfoA.ID), "passwordA")
			Expect(err).NotTo(HaveOccurred())

			By("Giving error if a given wallet password is wrong")
			err = kmdService.CheckWalletPassword(string(newWalletInfoA.ID), "notpasswordA")
			Expect(err).To(HaveOccurred())

			By("Creating another new wallet")
			newWalletInfoB, err := kmdService.CreateWallet("wallet_B", "passwordB")
			Expect(err).NotTo(HaveOccurred())
			Expect(newWalletInfoB.DriverName).To(Equal("sqlite"))
			Expect(newWalletInfoB.Name).To(BeEquivalentTo("wallet_B"), "2nd wallet should have been created")

			By("Listing multiple wallets")
			walletsInfo, err = kmdService.ListWallets()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletsInfo).To(HaveLen(2), "All wallets should be listed")

			By("Retrieving information of a wallet with the given ID")
			retrievedWalletInfoA, err := kmdService.GetWalletInfo(string(newWalletInfoA.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedWalletInfoA.DriverName).To(Equal("sqlite"))
			Expect(retrievedWalletInfoA.Name).To(BeEquivalentTo("wallet_A"), "Should have the information of the correct wallet")

			By("Renaming a wallet")
			err = kmdService.RenameWallet(string(newWalletInfoB.ID), "B_wallet", "passwordB")
			Expect(err).NotTo(HaveOccurred())

			By("Checking if renamed wallet has new name")
			retrievedWalletInfoB, err := kmdService.GetWalletInfo(string(newWalletInfoB.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedWalletInfoB.Name).To(BeEquivalentTo("B_wallet"), "Renamed wallet should have new name")

			// By("Removing a wallet")
			// err = kmdService.RemoveWallet(string(retrievedWalletInfoA.ID), "passwordA")
			// Expect(err).NotTo(HaveOccurred())
			// // Should not in list after removal
			// walletsInfo, err = kmdService.ListWallets()
			// Expect(err).NotTo(HaveOccurred())
			// Expect(walletsInfo).To(HaveLen(1), "Removed wallet should not be in list")
			// // Should not be retrievable after removal
			// _, err = kmdService.GetWalletInfo(string(retrievedWalletInfoA.ID))
			// Expect(err).To(HaveOccurred(), "Removed wallet should not be retrievable")
		})

		It("can manage wallet sessions", func() {
			By("Initializing KMD")
			const walletDirName = ".test_sqlite_wallet_session"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(func() {
				kmdService.CleanUp()
				createKmdServiceCleanup(walletDirName)
			})

			// Import a wallet with a known master derivation key (MDK) to make the generated accounts predictable
			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Session Mgmt Test Wallet", "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Session Mgmt Test Wallet"), "Wallet should have been imported")

			By("Starting a new wallet session with the imported wallet")
			err = kmdService.StartSession(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Retrieving information of the wallet in the current session")
			retrievedWalletInfo, err := kmdService.Session().GetWalletInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(retrievedWalletInfo.Name).To(BeEquivalentTo("Session Mgmt Test Wallet"), "Should have the information of the correct wallet")

			By("Checking if session is for wallet with given (incorrect) ID")
			sessionIsForWallet, err := kmdService.SessionIsForWallet("7ae575eca54410806f0c1e99861417fd")
			Expect(err).NotTo(HaveOccurred())
			Expect(sessionIsForWallet).To(BeFalse())

			By("Checking if session is for wallet with given (correct) ID")
			sessionIsForWallet2, err := kmdService.SessionIsForWallet(string(importedWalletInfo.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(sessionIsForWallet2).To(BeTrue())

			By("Checking if session is still valid")
			Expect(kmdService.Session().Check()).To(Succeed())

			By("Getting the expiration date and time")
			exp := kmdService.Session().Expiration()
			Expect(exp).To(BeTemporally(">", time.Now()))

			By("Renewing session")
			Expect(kmdService.RenewSession()).To(Succeed())
			Expect(kmdService.Session().Expiration()).To(BeTemporally(">", exp), "The new expiration date should be later than the old one")

			By("Ending session")
			kmdService.EndSession()
			session := kmdService.Session() // Check session is removed
			Expect(session).To(BeNil())

			By("Starting another wallet session with a lifetime of 0 seconds")
			kmdService.Config.SessionLifetimeSecs = 0
			err = kmdService.StartSession(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Checking if the session is invalid because it expired")
			<-time.After(1 * time.Millisecond) // Wait a tiny bit
			Expect(kmdService.Session().Check()).To(MatchError("wallet session is expired"))

			By("Failing to renew session after it expired")
			Expect(kmdService.RenewSession()).To(MatchError("wallet session is expired"))
		})

		It("can import and export wallet mnemonics", func() {
			By("Initializing KMD")
			const walletDirName = ".test_sqlite_import_export"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(func() {
				kmdService.CleanUp()
				createKmdServiceCleanup(walletDirName)
			})

			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Import-Export Test Wallet", "password for imported wallet")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Import-Export Test Wallet"), "Wallet should have been imported")

			By("Listing all wallets")
			walletsInfo, err := kmdService.ListWallets()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletsInfo).To(HaveLen(1), "There should be 1 wallet in the list after importing a wallet")

			By("Starting a new wallet session with the imported wallet")
			err = kmdService.StartSession(string(importedWalletInfo.ID), "password for imported wallet")
			Expect(err).NotTo(HaveOccurred())

			By("Exporting wallet that was imported")
			exportedMnemonic, err := kmdService.Session().ExportWallet("password for imported wallet")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedMnemonic).To(Equal(testWalletMnemonic), "Exported mnemonic should be the same as imported mnemonic")
		})

		It("can manage accounts in wallet", func() {
			By("Initializing KMD")
			const walletDirName = ".test_sqlite_wallet_accts"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(func() {
				kmdService.CleanUp()
				createKmdServiceCleanup(walletDirName)
			})

			// Import a wallet with a known master derivation key (MDK) to make the generated accounts predictable
			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Acct Mgmt Test Wallet", "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Acct Mgmt Test Wallet"), "Wallet should have been imported")

			By("Starting a new wallet session with the imported wallet")
			err = kmdService.StartSession(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Listing 0 accounts in session wallet")
			accts, err := kmdService.Session().ListAccounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).To(HaveLen(0), "There should be no accounts in the wallet")

			By("Generating an account in the session wallet using its key")
			walletAcctAddrA, err := kmdService.Session().GenerateAccount()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletAcctAddrA).To(Equal("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A"))

			By("Generating another account in the session wallet")
			walletAcctAddrB, err := kmdService.Session().GenerateAccount()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletAcctAddrB).To(Equal("V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU"))

			By("Importing an account into the session wallet")
			importedAcctAddr, err := kmdService.Session().ImportAccount(testStandaloneAcctMnemonic)
			Expect(err).NotTo(HaveOccurred())
			Expect(importedAcctAddr).To(Equal(testStandaloneAcctAddr))

			By("By listing multiple accounts within the session wallet")
			accts, err = kmdService.Session().ListAccounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).To(HaveLen(3), "All accounts in wallet should be listed")

			By("Exporting an account in the session wallet")
			exportedWalletMnemonic, err := kmdService.Session().ExportAccount(walletAcctAddrA, "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedWalletMnemonic).To(
				Equal("zero weekend library concert youth ancient bus report style mixed mansion wrong purchase bench satisfy clock need wave math inflict aisle ignore buddy above decide"),
				"Wallet account should have been exported",
			)

			By("Exporting an imported account stored in the session wallet")
			exportedImportedMnemonic, err := kmdService.Session().ExportAccount(testStandaloneAcctAddr, "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedImportedMnemonic).To(Equal(testStandaloneAcctMnemonic), "Imported account should have been exported")

			By("Removing a generated account from the session wallet")
			err = kmdService.Session().RemoveAccount("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A")
			Expect(err).NotTo(HaveOccurred())

			By("Checking removed account is not in session wallet")
			accts, err = kmdService.Session().ListAccounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).NotTo(ContainElement("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A"), "The removed wallet account should not be in the account list")

			By("Removing an imported account from the session wallet")
			err = kmdService.Session().RemoveAccount(testStandaloneAcctAddr)
			Expect(err).NotTo(HaveOccurred())

			By("Checking removed imported account is not in wallet")
			accts, err = kmdService.Session().ListAccounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).NotTo(ContainElement(testStandaloneAcctAddr), "The removed imported account should not be in the account list")
		})

		It("can sign a transaction", func() {
			const knownSignedTxnB64 = "gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ=="

			By("Initializing KMD")
			const walletDirName = ".test_sqlite_sign_txn"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(func() {
				kmdService.CleanUp()
				createKmdServiceCleanup(walletDirName)
			})

			// Import a wallet with a known master derivation key (MDK) to make the generated accounts predictable
			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Sign Txn Test Wallet", "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Sign Txn Test Wallet"), "Wallet should have been imported")

			By("Starting a new wallet session with the imported wallet")
			err = kmdService.StartSession(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Importing an account")
			importedAcctAddr, err := kmdService.Session().ImportAccount(testStandaloneAcctMnemonic)
			Expect(err).NotTo(HaveOccurred())
			Expect(importedAcctAddr).To(Equal(testStandaloneAcctAddr))

			By("Signing a transaction with an account when address is given")
			// Convert Base64 encoded signed transaction into a SignedTxn struct
			knownSignedTxn := types.SignedTxn{}
			err = knownSignedTxn.FromBase64String(knownSignedTxnB64)
			Expect(err).NotTo(HaveOccurred())
			// Convert core transaction (without signature) to Base64 string
			unsignedTxnB64 := base64.StdEncoding.EncodeToString(msgpack.Encode(knownSignedTxn.Txn))
			// Sign transaction
			outputSignedTxn, err := kmdService.Session().SignTransaction(unsignedTxnB64, testStandaloneAcctAddr)
			Expect(err).NotTo(HaveOccurred())
			Expect(outputSignedTxn).To(Equal(knownSignedTxnB64))

			By("Signing a transaction with an account when address is NOT given")
			// NOTE: When an empty address is given, the transaction sender should be used as the signing address
			outputSignedTxn2, err := kmdService.Session().SignTransaction(unsignedTxnB64, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(outputSignedTxn2).To(Equal(knownSignedTxnB64))
		})
	})
})

// createKmdService is a helper function that a new KMDService that is
// configured to use the given walletDirName as the wallet directory
func createKmdService(walletDirName string) KMDService {
	// Clean up by will wipe sensitive data if process is terminated suddenly
	memguard.CatchInterrupt()
	// Create the service
	return KMDService{
		Config: config.KMDConfig{
			SessionLifetimeSecs: 3600,
			DriverConfig: config.DriverConfig{
				SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{
					WalletsDir:   walletDirName,
					UnsafeScrypt: true, // For testing purposes only
					ScryptParams: config.ScryptParams{
						ScryptN: 2,
						ScryptR: 1,
						ScryptP: 1,
					},
				},
				LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
			},
		},
	}
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
