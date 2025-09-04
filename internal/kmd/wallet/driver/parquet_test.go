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
				logger := logging.New()
				logger.SetLevel(logging.InfoLevel)

				parquetDriver.InitWithConfig(config.KMDConfig{
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

				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("returns no wallet metadatas and does not generate a `metadatas.parquet` file if there are no wallets", func() {
				Expect(parquetDriver.ListWalletMetadatas()).To(BeEmpty(), "Returns no wallets")

				_, err := os.Stat(walletDirName + "/" + driver.ParquetMetadatasFile)
				Expect(err).To(MatchError(os.ErrNotExist), "No `metadatas.parquet` file was generated")
			})

			PIt("returns the wallet metadatas of the wallets in the wallet directory if `metadatas.parquet` file exists", func() {
				By("Creating 2 wallets")
				// TODO

				By("Listing wallets")
				// TODO
			})

			PIt("generates a `metadatas.parquet` file containing the metadata of each wallet if the file does not exist", func() {
				By("Removing the generated `metadatas.parquet` file")
				// TODO

				By("Listing wallets")
				// TODO

				By("Checking if a new `metadatas.parquet` file was generated")
				//
			})
		})

		Describe("CreateWallet()", Ordered, func() {
			const walletDirName = ".test_pq_create_wallet"
			var parquetDriver driver.ParquetWalletDriver

			BeforeAll(func() {
				logger := logging.New()
				logger.SetLevel(logging.InfoLevel)

				parquetDriver.InitWithConfig(config.KMDConfig{
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

				DeferCleanup(func() {
					createKmdServiceCleanup(walletDirName)
				})
			})

			It("fails if wallet name is too long", func() {
				err := parquetDriver.CreateWallet(
					[]byte("Foooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo"),
					[]byte("abc123"),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err.Error()).To(ContainSubstring("wallet name too long"))
			})

			It("fails if wallet ID is too long", func() {
				err := parquetDriver.CreateWallet(
					[]byte("Foo"),
					[]byte("000000000000000000000000000000000000000000000000000000000000000000000"),
					[]byte("password"),
					algoTypes.MasterDerivationKey{},
				)
				Expect(err.Error()).To(ContainSubstring("wallet id too long"))
			})

			It("creates a new wallet with the given name and ID if there are NO existing wallets", func() {
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

		PDescribe("RenameWallet()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("FetchWallet()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("Metadata()", func() {
			It("", func() {
				//
			})
		})

		PDescribe("()", func() {
			It("", func() {
				//
			})
		})
	})

	PDescribe("ParquetWallet", func() {
		Describe("Init()", func() {
			It("", func() {
				//
			})
		})

		Describe("CheckPassword()", func() {
			It("", func() {
				//
			})
		})

		Describe("ExportMasterDerivationKey()", func() {
			It("", func() {
				//
			})
		})

		Describe("Metadata()", func() {
			It("", func() {
				//
			})
		})

		Describe("ListKeys()", func() {
			It("", func() {
				//
			})
		})

		Describe("ImportKey()", func() {
			It("", func() {
				//
			})
		})

		Describe("ExportKey()", func() {
			It("", func() {
				//
			})
		})

		Describe("GenerateKey()", func() {
			It("", func() {
				//
			})
		})

		Describe("DeleteKey()", func() {
			It("", func() {
				//
			})
		})

		Describe("ImportMultisigAddr()", func() {
			It("", func() {
				//
			})
		})

		Describe("LookupMultisigPreimage()", func() {
			It("", func() {
				//
			})
		})

		Describe("ListMultisigAddrs()", func() {
			It("", func() {
				//
			})
		})

		Describe("DeleteMultisigAddr()", func() {
			It("", func() {
				//
			})
		})

		Describe("SignTransaction()", func() {
			It("", func() {
				//
			})
		})

		Describe("MultisigSignTransaction()", func() {
			It("", func() {
				//
			})
		})

		Describe("SignProgram()", func() {
			It("", func() {
				//
			})
		})

		Describe("MultisigSignProgram()", func() {
			It("", func() {
				//
			})
		})
	})
})

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
