package driver_test

import (
	"database/sql"
	"encoding/json"
	"os"

	algoTypes "github.com/algorand/go-algorand-sdk/v2/types"
	logging "github.com/sirupsen/logrus"

	"duckysigner/internal/kmd/config"
	"duckysigner/internal/kmd/wallet/driver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Parquet Wallet Driver", func() {

	Describe("ParquetWalletDriver", func() {

		Describe("ListWalletMetadatas()", Ordered, func() {
			const walletDirName = ".test_pq_list_metas"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns no wallet metadatas and does not generate a `metadatas.parquet` file if there are no wallets", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet
				Expect(parquetDriver.ListWalletMetadatas()).To(BeEmpty(), "Returns no wallet metadatas")
				_, err := os.Stat(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).To(MatchError(os.ErrNotExist), "No `metadatas.parquet` file was generated")
			})

			It("returns the wallet metadatas of the wallets in the wallet directory if `metadatas.parquet` file exists", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating 2 wallets (which generates a metadatas file)")
				walletId1 := "000"
				walletId2 := "001"
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId1),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())
				err = parquetDriver.CreateWallet(
					[]byte("Bar"),
					[]byte(walletId2),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Listing wallets")
				Expect(parquetDriver.ListWalletMetadatas()).To(HaveLen(2), "Returns 2 wallet metadatas")

				By("Checking `metadatas.parquet` file")
				db, err := sql.Open("duckdb", "")
				Expect(err).ToNot(HaveOccurred())
				defer db.Close()

				rows, err := db.Query(
					"SELECT wallet_id FROM read_parquet('" + walletDirName + "/" + driver.ParquetMetadatasFile + "')",
				)
				Expect(err).ToNot(HaveOccurred())

				var metaWalletId string

				// Get first wallet metadata
				rows.Next()
				err = rows.Scan(&metaWalletId)
				Expect(err).ToNot(HaveOccurred())
				Expect(metaWalletId).To(Equal(walletId1), "The first wallet is in the metadatas")
				// Get second wallet metadata
				rows.Next()
				err = rows.Scan(&metaWalletId)
				Expect(err).ToNot(HaveOccurred())
				Expect(metaWalletId).To(Equal(walletId2), "The second wallet is in the metadatas")
				rows.Close()
			})

			It("generates a `metadatas.parquet` file containing the metadata of each wallet if the file does not exist", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that multiple wallets have been created

				By("Removing the generated `metadatas.parquet` file")
				err := os.Remove(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).ToNot(HaveOccurred())

				By("Listing wallets")
				Expect(parquetDriver.ListWalletMetadatas()).To(HaveLen(2), "Returns 2 wallet metadatas")

				By("Checking if a new `metadatas.parquet` file was generated")
				_, err = os.Stat(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).ToNot(HaveOccurred(), "A new `metadatas.parquet` file was generated")
			})

			It("skips listing directories that are not wallets", func() {
				By("Creating a new directory within the wallets directory")
				err := os.Mkdir(walletDirName+"/not_a_wallet", 0700)
				Expect(err).ToNot(HaveOccurred())

				By("Listing wallets")
				Expect(parquetDriver.ListWalletMetadatas()).To(HaveLen(2),
					"Returns 2 wallet metadatas")
			})
		})

		Describe("CreateWallet()", Ordered, func() {
			const walletDirName = ".test_pq_create_wallet"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if wallet name is too long", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet
				err := parquetDriver.CreateWallet(
					[]byte("Foooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo"),
					[]byte("abc123"),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err.Error()).To(ContainSubstring("wallet name too long"))
			})

			It("fails if wallet ID is too long", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte("000000000000000000000000000000000000000000000000000000000000000000000"),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err.Error()).To(ContainSubstring("wallet id too long"))
			})

			It("creates a new wallet with the given name and ID if there are NO existing wallets", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				walletId := "0000000000"
				walletName := "Foo"

				By("Creating a new wallet")
				err := parquetDriver.CreateWallet(
					[]byte(walletName),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Checking if wallet metadata file exists")
				metadataFileContents, err := os.ReadFile(walletDirName + "/" + walletId + "/" + driver.ParquetWalletMetadataFile)
				Expect(err).ToNot(HaveOccurred())
				metadata := &driver.ParquetWalletMetadata{}
				err = json.Unmarshal(metadataFileContents, metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.DriverName).To(Equal("parquet"))
				Expect(metadata.DriverVersion).To(Equal(1))
				Expect(metadata.WalletId).To(Equal(walletId))
				Expect(metadata.WalletName).To(Equal(walletName))
				Expect(metadata.MEPEncrypted).ToNot(BeEmpty())
				Expect(metadata.MDKEncrypted).ToNot(BeEmpty())
				Expect(metadata.MaxKeyIdxEncrypted).ToNot(BeEmpty())

				By("Checking if wallet is added to metadatas file")
				db, err := sql.Open("duckdb", "")
				Expect(err).ToNot(HaveOccurred())
				defer db.Close()

				row := db.QueryRow(
					"FROM read_parquet('"+walletDirName+"/"+driver.ParquetMetadatasFile+"') WHERE wallet_id = ?",
					walletId,
				)
				var (
					metaDriverName         string
					metaDriverVersion      int
					metaWalletId           string
					metaWalletName         string
					metaMepEncrypted       []byte
					metaMdkEncrypted       []byte
					metaMaxKeyIdxEncrypted []byte
				)
				err = row.Scan(
					&metaDriverName,
					&metaDriverVersion,
					&metaWalletId,
					&metaWalletName,
					&metaMepEncrypted,
					&metaMdkEncrypted,
					&metaMaxKeyIdxEncrypted,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(metaDriverName).To(Equal("parquet"))
				Expect(metaDriverVersion).To(Equal(1))
				Expect(metaWalletId).To(Equal(walletId))
				Expect(metaWalletName).To(Equal(walletName))
				Expect(metaMepEncrypted).ToNot(BeEmpty())
				Expect(metaMdkEncrypted).ToNot(BeEmpty())
				Expect(metaMaxKeyIdxEncrypted).ToNot(BeEmpty())
			})

			It("fails if directory for wallet already exists", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet with a certain ID has been created

				walletId := "0000000000"

				By("Attempting to create a new wallet using same wallet ID")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).To(MatchError(os.ErrExist))
			})

			It("creates a new wallet with the given name and ID if it does not exist and there are other wallets", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet with a certain ID has been created

				walletId := "1111111111"
				walletName := "Bar"

				By("Creating a new wallet")
				err := parquetDriver.CreateWallet(
					[]byte(walletName),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Checking if wallet metadata file exists")
				metadataFileContents, err := os.ReadFile(walletDirName + "/" + walletId + "/" + driver.ParquetWalletMetadataFile)
				Expect(err).ToNot(HaveOccurred())
				metadata := &driver.ParquetWalletMetadata{}
				err = json.Unmarshal(metadataFileContents, metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.DriverName).To(Equal("parquet"))
				Expect(metadata.DriverVersion).To(Equal(1))
				Expect(metadata.WalletId).To(Equal(walletId))
				Expect(metadata.WalletName).To(Equal(walletName))
				Expect(metadata.MEPEncrypted).ToNot(BeEmpty())
				Expect(metadata.MDKEncrypted).ToNot(BeEmpty())
				Expect(metadata.MaxKeyIdxEncrypted).ToNot(BeEmpty())

				By("Checking if wallet is added to metadatas file")
				db, err := sql.Open("duckdb", "")
				Expect(err).ToNot(HaveOccurred())
				defer db.Close()

				row := db.QueryRow(
					"FROM read_parquet('"+walletDirName+"/"+driver.ParquetMetadatasFile+"') WHERE wallet_id = ?",
					walletId,
				)
				var (
					metaDriverName         string
					metaDriverVersion      int
					metaWalletId           string
					metaWalletName         string
					metaMepEncrypted       []byte
					metaMdkEncrypted       []byte
					metaMaxKeyIdxEncrypted []byte
				)
				err = row.Scan(
					&metaDriverName,
					&metaDriverVersion,
					&metaWalletId,
					&metaWalletName,
					&metaMepEncrypted,
					&metaMdkEncrypted,
					&metaMaxKeyIdxEncrypted,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(metaDriverName).To(Equal("parquet"))
				Expect(metaDriverVersion).To(Equal(1))
				Expect(metaWalletId).To(Equal(walletId))
				Expect(metaWalletName).To(Equal(walletName))
				Expect(metaMepEncrypted).ToNot(BeEmpty())
				Expect(metaMdkEncrypted).ToNot(BeEmpty())
				Expect(metaMaxKeyIdxEncrypted).ToNot(BeEmpty())
			})
		})

		Describe("RenameWallet()", Ordered, func() {
			const walletDirName = ".test_pq_rename_wallet"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails when there is no metadatas file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet
				By("Attempting to rename wallet with the given ID")
				err := parquetDriver.RenameWallet(
					[]byte("my new name"),
					[]byte("000"),
					[]byte("password"),
				)
				Expect(err).To(MatchError("wallet not found"))
			})

			It("renames the wallet to the new name", func() {
				By("Creating a wallet")
				walletId := "000"
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Renaming the wallet")
				newWalletName := "my new name"
				err = parquetDriver.RenameWallet(
					[]byte(newWalletName),
					[]byte("000"),
					[]byte("password"),
				)
				Expect(err).ToNot(HaveOccurred())

				By("Checking if the wallet has new name in the metadatas file")
				db, err := sql.Open("duckdb", "")
				Expect(err).ToNot(HaveOccurred())
				defer db.Close()

				row := db.QueryRow(
					"SELECT wallet_name FROM read_parquet('"+walletDirName+"/"+driver.ParquetMetadatasFile+"') WHERE wallet_id = ?",
					walletId,
				)
				var retrievedWalletName string
				err = row.Scan(&retrievedWalletName)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedWalletName).To(Equal(newWalletName),
					"The wallet is renamed in the wallet's metadata file")

				By("Checking if the wallet's metadata file has the new name")
				metadataFileContents, err := os.ReadFile(walletDirName + "/" + walletId + "/" + driver.ParquetWalletMetadataFile)
				Expect(err).ToNot(HaveOccurred())

				metadata := driver.ParquetWalletMetadata{}
				err = json.Unmarshal(metadataFileContents, &metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.WalletName).To(Equal(newWalletName),
					"The wallet is renamed in the metadatas file")
			})

			It("does not fail if the given password is incorrect", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created
				By("Attempting to rename wallet with the given ID with an incorrect password")
				err := parquetDriver.RenameWallet(
					[]byte("another new name"),
					[]byte("000"),
					[]byte("not the password"),
				)
				Expect(err).ToNot(HaveOccurred(),
					"Renaming the wallet succeeds despite having the wrong password")
			})

			It("fails when a wallet with the given ID does not exist in the metadatas file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Remove metadatas file")
				err := os.Remove(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to rename wallet with the given ID")
				err = parquetDriver.RenameWallet(
					[]byte("my new name"),
					[]byte("000"),
					[]byte("password"),
				)
				Expect(err).To(MatchError("wallet not found"))
			})

			It("fails if the directory for the wallet does not have a metadata.json file", func() {
				walletId := "fff"
				By("Creating a new directory within the wallets directory")
				err := os.Mkdir(walletDirName+"/"+walletId, 0700)
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to rename wallet with the given ID")
				err = parquetDriver.RenameWallet(
					[]byte("my new name"),
					[]byte(walletId),
					[]byte("password"),
				)
				Expect(err).To(MatchError("wallet not found"))
			})

			It("fails if the new name is too long", func() {
				By("Attempting to rename wallet with the given ID")
				err := parquetDriver.RenameWallet(
					[]byte("looooooooooooooooooooooooooooooooooooooooooooooooooooooooong new name"),
					[]byte("000"),
					[]byte("password"),
				)
				Expect(err.Error()).To(ContainSubstring("wallet name too long"))
			})

			It("fails if no ID is given", func() {
				By("Attempting to rename wallet with no ID")
				err := parquetDriver.RenameWallet(
					[]byte("looooooooooooooooooooooooooooooooooooooooooooooooooooooooong new name"),
					[]byte{},
					[]byte("password"),
				)
				Expect(err.Error()).To(ContainSubstring("no ID is given"))
			})
		})

		Describe("FetchWallet()", Ordered, func() {
			const walletDirName = ".test_pq_fetch_wallet"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails when there is no metadatas file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet
				By("Attempting to fetch wallet with an ID")
				_, err := parquetDriver.FetchWallet([]byte("000"))
				Expect(err).To(MatchError("wallet not found"))
			})

			It("returns wallet with the given ID", func() {
				By("Creating a wallet")
				walletId := "000"
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet")
				wallet, err := parquetDriver.FetchWallet([]byte("000"))
				Expect(err).ToNot(HaveOccurred())

				By("Checking fetched wallet's metadata")
				metadata, err := wallet.Metadata()
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.ID).To(Equal([]byte("000")))
			})

			It("fails when a wallet with the given ID does not exist", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created
				By("Attempting to fetch wallet with a different ID")
				_, err := parquetDriver.FetchWallet([]byte("111"))
				Expect(err).To(MatchError("wallet not found"))
			})

			It("fails if no ID is given", func() {
				By("Attempting to fetch wallet with no ID")
				_, err := parquetDriver.FetchWallet([]byte{})
				Expect(err.Error()).To(ContainSubstring("no ID is given"))
			})
		})
	})

	Describe("ParquetWallet", Ordered, func() {

		Describe("Metadata()", func() {
			const walletDirName = ".test_pq_wallet_metadata"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns the wallet's metadata if it is in the metadatas file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				walletId := "000"
				walletName := "Foo"
				err := parquetDriver.CreateWallet(
					[]byte(walletName),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet")
				wallet, err := parquetDriver.FetchWallet([]byte("000"))
				Expect(err).ToNot(HaveOccurred())

				By("Getting the wallet's metadata")
				metadata, err := wallet.Metadata()
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata.ID).To(Equal([]byte(walletId)))
				Expect(metadata.Name).To(Equal([]byte(walletName)))
				Expect(metadata.DriverName).To(Equal("parquet"))
			})

			It("fails if the wallet's metadata is not in the metadatas file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet")
				wallet, err := parquetDriver.FetchWallet([]byte("000"))
				Expect(err).ToNot(HaveOccurred())

				By("Remove metadatas file")
				err = os.Remove(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).ToNot(HaveOccurred())

				By("Getting the wallet's metadata")
				_, err = wallet.Metadata()
				Expect(err).To(MatchError("wallet not found"))
			})
		})

		Describe("CheckPassword()", func() {
			PIt("does not return error if the given wallet password is correct", func() {
				//
			})

			PIt("returns error if the given wallet password is incorrect", func() {
				//
			})
		})

		PDescribe("ExportMasterDerivationKey()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("ListKeys()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("ImportKey()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("ExportKey()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("GenerateKey()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("DeleteKey()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("ImportMultisigAddr()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("LookupMultisigPreimage()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("ListMultisigAddrs()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("DeleteMultisigAddr()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("SignTransaction()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("MultisigSignTransaction()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("SignProgram()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("MultisigSignProgram()", func() {
			It("", func() {
				//
			})
		})
	})
})

// setupParquetWalletDriver configures and initializes the given parquet driver
// for use in a test
func setupParquetWalletDriver(parquetDriver *driver.ParquetWalletDriver, walletDirName string) {
	logger := logging.New()
	logger.SetLevel(logging.InfoLevel)

	err := parquetDriver.InitWithConfig(config.KMDConfig{
		SessionLifetimeSecs: 3600,
		DriverConfig: config.DriverConfig{
			ParquetWalletDriverConfig: config.ParquetWalletDriverConfig{
				WalletsDir:   walletDirName,
				UnsafeScrypt: true, // For testing purposes only
				ScryptParams: config.ScryptParams{ScryptN: 2, ScryptR: 1, ScryptP: 1},
			},
			SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{UnsafeScrypt: true},
			LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
		},
	}, logger)

	Expect(err).NotTo(HaveOccurred())
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
