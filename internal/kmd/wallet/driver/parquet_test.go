package driver_test

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	algoTypes "github.com/algorand/go-algorand-sdk/v2/types"
	logging "github.com/sirupsen/logrus"

	"duckysigner/internal/kmd/config"
	"duckysigner/internal/kmd/wallet"
	"duckysigner/internal/kmd/wallet/driver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parquet Wallet Driver", func() {

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
				const walletId1 = "000"
				const walletId2 = "001"
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
				const walletId = "000"
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Renaming the wallet")
				const newWalletName = "my new name"
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
				const walletId = "fff"
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
				const walletId = "000"
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

	Describe("ParquetWallet", func() {

		Describe("Metadata()", Ordered, func() {
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
				const walletId = "000"
				const walletName = "Foo"
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

		Describe("CheckPassword()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_ck_pw"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("does not return error if the given wallet password is correct", func() {
				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Checking the given password")
				err = wallet.CheckPassword([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred(), "The password is correct")
			})

			It("returns error if the given wallet password is incorrect", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Checking the given password")
				err = wallet.CheckPassword([]byte("not the password"))
				Expect(err).To(HaveOccurred(), "The password is incorrect")
			})
		})

		Describe("ExportMasterDerivationKey()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_export_mdk"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns master derivation key when given the correct password", func() {
				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Exporting the MDK with the correct password")
				mdk, err := wallet.ExportMasterDerivationKey([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())
				Expect(mdk).ToNot(BeEmpty(), "Returns MDK")
			})

			It("fails if given password is incorrect", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export MDK with an incorrect password")
				_, err = wallet.ExportMasterDerivationKey([]byte("wrong password"))
				Expect(err).To(HaveOccurred(), "Exporting MDK failed because password is incorrect")
			})
		})

		Describe("CheckAddrInWallet()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_check_addr"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			const acctAddr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
			const acctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

			var wallet wallet.Wallet

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err = parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Generating a key")
				_, err = wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false if address is NOT in wallet", func() {
				By("Checking if a certain address is stored in wallet")
				check, err := wallet.CheckAddrInWallet(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				Expect(check).To(BeFalse())
			})

			It("returns true if address is in wallet", func() {
				By("Importing a key")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())

				By("Checking if a certain address is stored in wallet")
				check, err := wallet.CheckAddrInWallet(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				Expect(check).To(BeTrue())
			})
		})

		Describe("ListKeys()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_list_keys"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns no addresses if there are no keys stored", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Listing all keys")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(0), "No addresses were returned")
			})

			It("returns all keys stored within the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Generating 2 keys")
				_, err = wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())

				By("Listing all keys")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(2), "All keys were returned")
			})
		})

		Describe("ImportKey()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_import_key"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			const acctAddr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
			const acctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("imports the key when there are no keys in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				addr, err := wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())
				Expect(algoTypes.Address(addr).String()).To(Equal(acctAddr),
					"Imported key has correct address")

				By("Checking if new key was added to the file")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(1), "New key was added to the keys file")
			})

			It("imports the key when there is at least one key in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with one key within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				const testAcctAddr = "3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY"
				const testAcctMnemonic = "sugar bronze century excuse animal jacket what rail biology symbol want craft annual soul increase question army win execute slim girl chief exhaust abstract wink"
				sk, err := mnemonic.ToPrivateKey(testAcctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				addr, err := wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())
				Expect(algoTypes.Address(addr).String()).To(Equal(testAcctAddr),
					"Imported key has correct address")

				By("Checking if new key was added to the file")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(2), "New key was added to the keys file")
			})

			It("imports the key when there is at least one generated (indexed) key in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with 2 keys within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Generating a key")
				_, err = wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				const testAcctAddr = "DEEZK32T3M7W5HAG5LNPOQK2E7LNOBLEBKI42L5HYVD4Z3JRLIZSLJ2OEU"
				const testAcctMnemonic = "rapid wire salon common praise rifle sunset save hurdle dawn mail average process icon just tooth fiction home kiwi tuna example stage reflect absent typical"
				sk, err := mnemonic.ToPrivateKey(testAcctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				addr, err := wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())
				Expect(algoTypes.Address(addr).String()).To(Equal(testAcctAddr),
					"Imported key has correct address")

				By("Checking if new key was added to the file")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(4), "New key was added to the keys file")
			})

			It("fails to import key if it already exists in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to import a key that is already in wallet")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportKey(sk)
				Expect(err).To(MatchError("key already exists in wallet"), "Duplicate key is not accepted")
			})
		})

		Describe("ExportKey()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_export_key"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			const acctAddr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
			const acctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if there are no keys stored in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a key")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ExportKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).To(MatchError("key does not exist in this wallet"), "Key export failed")
			})

			It("returns the key for the given address if it is stored in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with no keys within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())

				By("Exporting a key")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				exportedKey, err := wallet.ExportKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())
				Expect(mnemonic.FromPrivateKey(exportedKey)).To(Equal(acctMnemonic),
					"The correct key was exported")
			})

			It("fails if given the wrong password", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a key with the wrong password")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ExportKey(algoTypes.Digest(decodedAcctAddr), []byte("not the password"))
				Expect(err).To(HaveOccurred(), "Key export failed")
			})

			It("fails if the key for the given address is not stored in the non-empty wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a key that is not in the wallet")
				const testAcctAddr = "DEEZK32T3M7W5HAG5LNPOQK2E7LNOBLEBKI42L5HYVD4Z3JRLIZSLJ2OEU"
				decodedAcctAddr, err := algoTypes.DecodeAddress(testAcctAddr)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ExportKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).To(MatchError("key does not exist in this wallet"), "Key export failed")
			})
		})

		Describe("GenerateKey()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_gen_key"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("generates a new key when there are no keys in the wallet", func() {
				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Generating a key")
				newAddr, err := wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(newAddr).To(HaveLen(32))

				By("Checking if new key was added to the file")
				_, err = os.Stat(walletDirName + "/" + walletId + "/" + driver.ParquetWalletKeysFile)
				Expect(err).ToNot(HaveOccurred(), "The keys file was created")
			})

			It("generates a new key when there is at least one key in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Generating another key")
				newAddr, err := wallet.GenerateKey(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(newAddr).To(HaveLen(32))

				By("Checking if new key was added to the file")
				addrs, err := wallet.ListKeys()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(2), "New key was added to the keys file")
			})
		})

		Describe("DeleteKey()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_del_key"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			const acctAddr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
			const acctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if there is no keys file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to delete a key")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				err = wallet.DeleteKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).To(HaveOccurred())
			})

			It("removes the key for the given address", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with no keys within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())

				By("Deleting a key")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				err = wallet.DeleteKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Checking if key has been removed")
				_, err = wallet.ExportKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).To(HaveOccurred(), "Key was removed")
			})

			It("fails if given the wrong password", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a key with the wrong password")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				err = wallet.DeleteKey(algoTypes.Digest(decodedAcctAddr), []byte("not the password"))
				Expect(err).To(HaveOccurred(), "Key deletion failed")
			})

			It("does not fail if key to be removed is not in wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to delete a key that is not in the wallet")
				decodedAcctAddr, err := algoTypes.DecodeAddress(acctAddr)
				Expect(err).ToNot(HaveOccurred())
				err = wallet.DeleteKey(algoTypes.Digest(decodedAcctAddr), []byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("ImportMultisigAddr()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_import_msig"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("imports the multisig address when there are no multisig addresses in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a multisignature address")
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())
				msigAcct, err := crypto.MultisigAccountWithParams(1, 1, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				msigAddr, err := msigAcct.Address()
				Expect(err).ToNot(HaveOccurred())

				addr, err := wallet.ImportMultisigAddr(msigAcct.Version, msigAcct.Threshold, msigAcct.Pks)
				Expect(err).ToNot(HaveOccurred())
				Expect(algoTypes.Address(addr).String()).To(Equal(msigAddr.String()),
					"Imported multisig address")

				By("Checking if new multisig address was added to the file")
				addrs, err := wallet.ListMultisigAddrs()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(1), "New multisig address was added to the file")
			})

			It("imports the multisig address when there is at least one multisig address in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with one multisig
				// address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a multisignature address")
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())
				msigAcct, err := crypto.MultisigAccountWithParams(1, 2, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				msigAddr, err := msigAcct.Address()
				Expect(err).ToNot(HaveOccurred())

				addr, err := wallet.ImportMultisigAddr(msigAcct.Version, msigAcct.Threshold, msigAcct.Pks)
				Expect(err).ToNot(HaveOccurred())
				Expect(algoTypes.Address(addr).String()).To(Equal(msigAddr.String()),
					"Imported multisig address")

				By("Checking if new key was added to the file")
				addrs, err := wallet.ListMultisigAddrs()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(2), "New multisig address was added to the file")
			})

			It("fails to import multisig address if it already exists in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one
				// multisig address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a multisignature address")
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())
				msigAcct, err := crypto.MultisigAccountWithParams(1, 1, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())

				_, err = wallet.ImportMultisigAddr(msigAcct.Version, msigAcct.Threshold, msigAcct.Pks)
				Expect(err).To(MatchError("multisignature address already exists in wallet"),
					"Duplicate multisig address is not accepted")
			})
		})

		Describe("LookupMultisigPreimage()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_multisig_lookup"
			var parquetDriver driver.ParquetWalletDriver
			var msigAcct crypto.MultisigAccount
			var msigAddr algoTypes.Address

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				// Create multisig account
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())
				msigAcct, err = crypto.MultisigAccountWithParams(1, 2, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				msigAddr, err = msigAcct.Address()
				Expect(err).ToNot(HaveOccurred())

				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if there are no multisig addresses stored in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to look up a multisignature preimage (version, threshold, public keys)")
				decodedAcctAddr, err := algoTypes.DecodeAddress(msigAddr.String())
				Expect(err).ToNot(HaveOccurred())
				_, _, _, err = wallet.LookupMultisigPreimage(algoTypes.Digest(decodedAcctAddr))
				Expect(err.Error()).To(ContainSubstring("does not exist in this wallet"), "Lookup failed")
			})

			It("returns the multisig preimage for the given address if it is stored in the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with no keys within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a multisig address")
				_, err = wallet.ImportMultisigAddr(msigAcct.Version, msigAcct.Threshold, msigAcct.Pks)
				Expect(err).ToNot(HaveOccurred())

				By("Looking up a multisig preimage")
				version, threshold, pks, err := wallet.LookupMultisigPreimage(algoTypes.Digest(msigAddr))
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(Equal(msigAcct.Version), "Returned the correct version")
				Expect(threshold).To(Equal(msigAcct.Threshold), "Returned the correct threshold")
				Expect(pks).To(HaveLen(2), "Returned the correct number of public keys")
			})

			It("fails if the given address is not a multisig address", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with one multisig
				// address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to look up with an non-multisig address")
				decodedAcctAddr, err := algoTypes.DecodeAddress("DEEZK32T3M7W5HAG5LNPOQK2E7LNOBLEBKI42L5HYVD4Z3JRLIZSLJ2OEU")
				Expect(err).ToNot(HaveOccurred())
				_, _, _, err = wallet.LookupMultisigPreimage(algoTypes.Digest(decodedAcctAddr))
				Expect(err).To(HaveOccurred(), "Lookup failed")
			})

			It("fails if the key for the given address is not stored in the non-empty wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with one multisig
				// address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a key that is not in the wallet")
				const testAcctAddr = "GQ3QPLJL4VKVGQCHPXT5UZTNZIJAGVJPXUHCJLRWQMFRVL4REVW7LJ3FGY"
				decodedAcctAddr, err := algoTypes.DecodeAddress(testAcctAddr)
				Expect(err).ToNot(HaveOccurred())
				_, _, _, err = wallet.LookupMultisigPreimage(algoTypes.Digest(decodedAcctAddr))
				Expect(err.Error()).To(ContainSubstring("does not exist in this wallet"), "Lookup failed")
			})
		})

		Describe("ListMultisigAddrs()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_list_msigs"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns no addresses if there are no multisig addresses stored", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Listing all multisig addresses")
				addrs, err := wallet.ListMultisigAddrs()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(0), "No addresses were returned")
			})

			It("returns all keys stored within the wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one
				// multisig address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing 2 multisig addresses")
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())

				msigAcct1, err := crypto.MultisigAccountWithParams(1, 1, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportMultisigAddr(msigAcct1.Version, msigAcct1.Threshold, msigAcct1.Pks)
				Expect(err).ToNot(HaveOccurred())

				msigAcct2, err := crypto.MultisigAccountWithParams(1, 2, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportMultisigAddr(msigAcct2.Version, msigAcct2.Threshold, msigAcct2.Pks)
				Expect(err).ToNot(HaveOccurred())

				By("Listing all multisig addresses")
				addrs, err := wallet.ListMultisigAddrs()
				Expect(err).ToNot(HaveOccurred())
				Expect(addrs).To(HaveLen(2), "All multisig addresses were returned")
			})
		})

		Describe("DeleteMultisigAddr()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_del_msig"
			var parquetDriver driver.ParquetWalletDriver
			var msigAcct crypto.MultisigAccount
			var msigAddr algoTypes.Address

			const walletId = "000"
			const walletPassword = "password"

			BeforeAll(func() {
				// Create multisig account
				addr1, err := algoTypes.DecodeAddress("RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A")
				Expect(err).ToNot(HaveOccurred())
				addr2, err := algoTypes.DecodeAddress("3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY")
				Expect(err).ToNot(HaveOccurred())
				msigAcct, err = crypto.MultisigAccountWithParams(1, 2, []algoTypes.Address{addr1, addr2})
				Expect(err).ToNot(HaveOccurred())
				msigAddr, err = msigAcct.Address()
				Expect(err).ToNot(HaveOccurred())

				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if there is no keys file", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that no wallets have been created yet

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to delete a multisig address")
				err = wallet.DeleteMultisigAddr(algoTypes.Digest(msigAddr), []byte(walletPassword))
				Expect(err).To(HaveOccurred())
			})

			It("removes the given multisig address", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with no keys within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a multisignature address")
				addr, err := wallet.ImportMultisigAddr(msigAcct.Version, msigAcct.Threshold, msigAcct.Pks)
				Expect(err).ToNot(HaveOccurred())

				By("Deleting a multisig address")
				err = wallet.DeleteMultisigAddr(algoTypes.Digest(addr), []byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Checking if multisig address has been removed")
				_, _, _, err = wallet.LookupMultisigPreimage(algoTypes.Digest(addr))
				Expect(err).To(HaveOccurred(), "Key was removed")
			})

			It("fails if given the wrong password", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one
				// multisig address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to export a multisig address with the wrong password")
				err = wallet.DeleteMultisigAddr(algoTypes.Digest(msigAddr), []byte("not the password"))
				Expect(err).To(HaveOccurred(), "Key deletion failed")
			})

			It("does not fail if multisig address to be removed is not in wallet", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one
				// multisig address within it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to delete a multisig address that is not in the wallet")
				err = wallet.DeleteMultisigAddr(algoTypes.Digest(msigAddr), []byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("SignTransaction()", Ordered, func() {
			const walletDirName = ".test_pq_wallet_sign_txn"
			var parquetDriver driver.ParquetWalletDriver

			const walletId = "000"
			const walletPassword = "password"

			const knownSignedTxnB64 = "gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ=="
			var knownSignedTxn algoTypes.SignedTxn

			var acctAddr algoTypes.Address
			const acctAddrStr = "RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A"
			const acctMnemonic = "minor print what witness play daughter matter light sign tip blossom anger artwork profit cart garment buzz resemble warm hole speed super bamboo abandon bonus"

			BeforeAll(func() {
				var err error

				// Decode string address to an Address
				acctAddr, err = algoTypes.DecodeAddress(acctAddrStr)
				Expect(err).ToNot(HaveOccurred())

				// Convert Base64 encoded signed transaction into a SignedTxn struct
				err = knownSignedTxn.FromBase64String(knownSignedTxnB64)
				Expect(err).NotTo(HaveOccurred())

				setupParquetWalletDriver(&parquetDriver, walletDirName)
				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("signs the given transaction", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with no keys within it

				By("Creating a wallet")
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte(walletId),
					[]byte(walletPassword),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err).ToNot(HaveOccurred())

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Importing a key")
				sk, err := mnemonic.ToPrivateKey(acctMnemonic)
				Expect(err).ToNot(HaveOccurred())
				_, err = wallet.ImportKey(sk)
				Expect(err).ToNot(HaveOccurred())

				By("Signing a transaction")
				outputSignedTxn, err := wallet.SignTransaction(
					knownSignedTxn.Txn,
					acctAddr[:],
					[]byte(walletPassword),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(base64.StdEncoding.EncodeToString(outputSignedTxn)).To(Equal(knownSignedTxnB64),
					"Transaction was signed correctly")
			})

			It("fails if given the wrong password", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to sign a transaction with the wrong password")
				_, err = wallet.SignTransaction(
					knownSignedTxn.Txn,
					acctAddr[:],
					[]byte("not the password"),
				)
				Expect(err).To(HaveOccurred(), "Signing transaction failed")
			})

			It("fails if given the public key is not for an account the wallet has the key for", func() {
				// NOTE: Because this `Describe` container is "Ordered", it is
				// assumed that a wallet has been created with at least one key
				// in it

				By("Fetching the wallet and initializing it")
				wallet, err := parquetDriver.FetchWallet([]byte(walletId))
				Expect(err).ToNot(HaveOccurred())
				err = wallet.Init([]byte(walletPassword))
				Expect(err).ToNot(HaveOccurred())

				By("Attempting to sign a transaction with an account not in wallet")
				const testAcctAddrStr = "3F3FPW6ZQQYD6JDC7FKKQHNGVVUIBIZOUI5WPSJEHBRABZDRN6LOTBMFEY"
				testAcctAddr, err := algoTypes.DecodeAddress(testAcctAddrStr)
				Expect(err).ToNot(HaveOccurred())

				_, err = wallet.SignTransaction(
					knownSignedTxn.Txn,
					testAcctAddr[:],
					[]byte(walletPassword),
				)
				Expect(err).To(HaveOccurred(), "Signing transaction failed")
			})
		})

		// PDescribe("MultisigSignTransaction()", Ordered, func() {
		// 	It("signs the given transaction", func() {
		// 		// TODO
		// 	})

		// 	It("fails if given the wrong password", func() {
		// 		// TODO
		// 	})
		// })

		// PDescribe("SignProgram()", Ordered, func() {
		// 	It("signs the given program", func() {
		// 		// TODO
		// 	})

		// 	It("fails if given the wrong password", func() {
		// 		// TODO
		// 	})
		// })

		// PDescribe("MultisigSignProgram()", Ordered, func() {
		// 	It("signs the given program", func() {
		// 		// TODO
		// 	})

		// 	It("fails if given the wrong password", func() {
		// 		// TODO
		// 	})
		// })
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
			SQLiteWalletDriverConfig: config.SQLiteWalletDriverConfig{
				UnsafeScrypt: true,
				Disable:      true,
				WalletsDir:   walletDirName,
			},
			LedgerWalletDriverConfig: config.LedgerWalletDriverConfig{Disable: true},
		},
	}, logger)

	Expect(err).NotTo(HaveOccurred())
}
