workspace {

    model {
        user = person "User" "A user of the Ducky Signer desktop wallet"
        wallet = softwareSystem "Ducky Signer" "Algorand wallet software for desktop computers" {
            walletUI = container "User Interface" ""
            walletServer = container "Wallet Connection Server" ""
            keyStore = container "Key Store" "Encrypted storage of the private keys for the user's accounts" {
                tags Database
            }
            grantStore = container "DApp Grants Store" "Storage for credentials and permissions granted to dApps" {
                tags Database
            }
            settingsStore = container "Settings Store" "Storage for the user's wallet settings. Maybe a file or localStorage." {
                tags Database
            }
        }
        dapp = softwareSystem "DApp" "App using Algorand" {
            tags DApp External
        }
        algorand = softwareSystem "Algorand" "Blockchain network" {
            tags Algorand External
        }
        algoNode = softwareSystem "Algorand Node" {
            tags AlgoNode External
        }
        ledgerDevice = softwareSystem "Ledger Device" "Hardware wallet device" {
            tags Database External
        }

        # System landscape relationships
        user -> wallet "Uses"
        user -> dapp "Interacts with"
        dapp -> user "Responds to"
        dapp -> wallet "Communicates with" "HTTP/2"
        dapp -> algoNode "Interacts with"
        algoNode -> algorand "Connected to"
        user -> ledgerDevice "Signs using"
        ledgerDevice -> user "Requests approval to sign from"
        # wallet -> algorand "Interacts with"

        # Wallet relationships
        user -> walletUI "Interfaces with"
        walletUI -> user "Requests input from"
        dapp -> walletServer "Makes requests to" "HTTP/2"
        walletUI -> walletServer "Notifies of user response" "Event"
        walletServer -> walletUI "Notifies of dApp request" "Event"
        walletUI -> keyStore "Temporarily retrieves key from" "With password from user"
        walletUI -> settingsStore "Retrieves user preferences from" ""
        walletUI -> ledgerDevice "Sends signing request to"
        walletServer -> grantStore "Manages credentials & permissions given to dApps" ""

    }

    views {
        systemlandscape "SystemLandscape" "System-wide view" {
            include *
            autoLayout
        }

        systemContext wallet "WalletContext" {
            include *
            autoLayout
        }

        container wallet "Containers" {
            include *
            autoLayout
        }

        dynamic wallet DAppConnect "DApp establish connection with wallet" {
            autoLayout
            user -> dapp "Initiates action that requires connecting to wallet within"
            dapp -> walletServer "Requests authentication credentials from"
            walletServer -> walletUI "Forwards request to approve dApp connection to"
            walletUI -> user "Asks for approval of dApp connection from"
            user -> walletUI "Approves dApp connection using"
            walletUI -> walletServer "Forwards dApp connection approval data to"
            walletServer -> grantStore "Saves dApp connection approval data into"
            walletServer -> dapp "Responds with the set of authentication credentials created from approval data to"
        }

        dynamic wallet SignTransaction "DApp requests user to sign a transaction" {
            autoLayout
            user -> dapp "Initiates action that requires signing a transaction within"
            dapp -> walletServer "Sends request to sign transaction to"
            walletServer -> walletUI "Forwards request to sign transaction to"
            walletUI -> user "Asks for password to retrieve key to sign transaction from"
            user -> walletUI "Enters password into"
            walletUI -> keyStore "Retrieves key for signing from"
            walletUI -> walletServer "Creates signed transaction data & forwards it to"
            walletServer -> dapp "Responds with signed transaction data to"
            dapp -> algoNode "Sends signed transaction using"
        }

        dynamic wallet DisconnectThroughDApp "User disconnects wallet from dApp though the dApp (e.g. \"Disconnect\" button)" {
            autoLayout
            user -> dapp "Initiates disconnect by clicking \"Disconnect wallet\" button within"
            dapp -> walletServer "Sends request to remove dApp connection approval data to"
            walletServer -> grantStore "Removes dApp connection approval data from"
            walletServer -> dapp "Responds with success message to"
            dapp -> user "Shows it is disconnected from wallet after removing its now old and invalid authentication credentials to"
        }

        dynamic wallet DisconnectThroughWallet "User disconnects wallet from dApp though the wallet" {
            autoLayout
            user -> walletUI "Initiates disconnect by clicking \"Disconnect wallet\" button within"
            walletUI -> walletServer "Sends request to disconnect from dApp to"
            walletServer -> grantStore "Removes dApp connection approval data from"
            walletServer -> walletUI "Responds with success message to"
            walletUI -> user "Shows it is disconnected from dApp to"
        }

        dynamic wallet DAppRenewCredentials "DApp renews its set of authentication credentials created from dApp connection approval data" {
            autoLayout
            dapp -> walletServer "Send request to renew authentication credentials to"
            walletServer -> grantStore "Checks if dApp connection approval is valid by looking in"
            walletServer -> dapp "Responds with new set of authentication credentials created from approval data to"
        }

        theme default

        styles {
            element External {
                background #e0d5e5
                color #222222
            }
            element DApp {
                shape Hexagon
            }
            element AlgoNode {
                shape Hexagon
            }
            element Algorand {
                shape Ellipse
            }
            element Database {
                shape Cylinder
            }
        }
    }

}
