package services_test

import (
	"duckysigner/kmd/config"
	. "duckysigner/services"
	"os"

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

		It("can manage wallets", func() {
			By("Initializing KMD")
			const walletDirName = ".test_wallet_mngmt"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(createKmdServiceCleanup(walletDirName))
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

		It("can import and export wallet mnemonics", func() {
			By("Initializing KMD")
			const walletDirName = ".test_sqlite_import_export"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(createKmdServiceCleanup(walletDirName))

			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Imported Wallet", "password for imported wallet")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Imported Wallet"), "Wallet should have been imported")

			By("Listing all wallets")
			walletsInfo, err := kmdService.ListWallets()
			Expect(err).NotTo(HaveOccurred())
			Expect(walletsInfo).To(HaveLen(1), "There should be 1 wallet in the list after importing a wallet")

			By("Exporting wallet that was imported")
			exportedMnemonic, err := kmdService.ExportWalletMnemonic(string(importedWalletInfo.ID), "password for imported wallet")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedMnemonic).To(Equal(testWalletMnemonic), "Exported mnemonic should be the same as imported mnemonic")
		})

		It("can manage accounts in wallet", func() {
			By("Initializing KMD")
			const walletDirName = ".test_sqlite_wallet_accts"
			kmdService := createKmdService(walletDirName)
			DeferCleanup(createKmdServiceCleanup(walletDirName))

			// Import a wallet with a known master derivation key (MDK) to make the generated accounts predictable
			By("Importing a wallet")
			importedWalletInfo, err := kmdService.ImportWalletMnemonic(testWalletMnemonic, "Test Wallet", "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedWalletInfo.DriverName).To(Equal("sqlite"))
			Expect(importedWalletInfo.Name).To(BeEquivalentTo("Test Wallet"), "Wallet should have been imported")

			By("Listing 0 accounts in wallet")
			accts, err := kmdService.ListAccountsInWallet(string(importedWalletInfo.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).To(HaveLen(0), "There should be no accounts in the wallet")

			By("Generating an account using wallet key (wallet account)")
			walletAcctAddrA, err := kmdService.GenerateWalletAccount(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(walletAcctAddrA).To(Equal("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A"))

			By("Generating another wallet account")
			walletAcctAddrB, err := kmdService.GenerateWalletAccount(string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(walletAcctAddrB).To(Equal("V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU"))

			By("Importing an account")
			importedAcctAddr, err := kmdService.ImportAccountIntoWallet(testStandaloneAcctMnemonic, string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(importedAcctAddr).To(Equal(testStandaloneAcctAddr))

			By("By listing multiple accounts within wallet")
			accts, err = kmdService.ListAccountsInWallet(string(importedWalletInfo.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).To(HaveLen(3), "All accounts in wallet should be listed")

			By("Exporting a wallet account")
			exportedWalletMnemonic, err := kmdService.ExportAccountInWallet(walletAcctAddrA, string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedWalletMnemonic).To(
				Equal("zero weekend library concert youth ancient bus report style mixed mansion wrong purchase bench satisfy clock need wave math inflict aisle ignore buddy above decide"),
				"Wallet account should have been exported",
			)

			By("Exporting an imported account stored in wallet")
			exportedImportedMnemonic, err := kmdService.ExportAccountInWallet(testStandaloneAcctAddr, string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())
			Expect(exportedImportedMnemonic).To(Equal(testStandaloneAcctMnemonic), "Imported account should have been exported")

			By("Removing a wallet account")
			err = kmdService.RemoveAccountFromWallet("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A", string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Checking removed wallet account is not in wallet")
			accts, err = kmdService.ListAccountsInWallet(string(importedWalletInfo.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).NotTo(ContainElement("H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A"), "The removed wallet account should not be in the account list")

			By("Removing an imported account")
			err = kmdService.RemoveAccountFromWallet(testStandaloneAcctAddr, string(importedWalletInfo.ID), "bad password")
			Expect(err).NotTo(HaveOccurred())

			By("Checking removed imported account is not in wallet")
			accts, err = kmdService.ListAccountsInWallet(string(importedWalletInfo.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(accts).NotTo(ContainElement(testStandaloneAcctAddr), "The removed imported account should not be in the account list")
		})
	})
})

// createKmdService is a helper function that a new KMDService that is
// configured to use the given walletDirName as the wallet directory
func createKmdService(walletDirName string) KMDService {
	// Create the service
	return KMDService{
		Config: config.KMDConfig{
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

// createKmdServiceCleanup is a helper function that creates function that
// cleans up the directory with specified by walletDirName, which was created
// for an instance of the KMDService
func createKmdServiceCleanup(walletDirName string) func() {
	return func() {
		// Remove test wallet directory
		err := os.RemoveAll(walletDirName)
		if !os.IsNotExist(err) { // If no "directory does not exist" error
			Expect(err).NotTo(HaveOccurred())
		}
	}
}
