workspace {

    model {
        user = person "Algorand User" "A user of my software system."
        wallet = softwareSystem "duckySigner" "Algorand wallet app" {
            walletUI = container "User Interface" ""
            walletServer = container "Wallet Connection Server" ""
            keyStore = container "Key Store" "Encrypted storage of the private keys for the user's accounts" {
                tags Database
            }
            grantStore = container "DApp Grants Store" "Storage for credentials and permissions granted to DApps" {
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
        ledgerDevice = softwareSystem "Ledger Device" "Ledger hardware wallet" {
            tags Database External
        }

        # System landscape relationships
        user -> wallet "Uses"
        user -> dapp "Interacts with"
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
        walletServer -> walletUI "Notifies of DApp request" "Event"
        walletUI -> keyStore "Temporarily retrieves key from" "With password from user"
        walletUI -> settingsStore "Retrieves user preferences from" ""
        walletUI -> grantStore "Saves credentials & perms given to DApps" ""
        walletUI -> ledgerDevice "Sends signing request to"
        # walletServer -> grantStore "" ""

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
            dapp -> walletServer "Send request for authentication credentials from"
            walletServer -> walletUI "Forwards request to approve DApp connection to"
            walletUI -> user "Requests approval for DApp connection from"
            user -> walletUI "Approves DApp connection using"
            walletUI -> grantStore "Saves DApp connection approval data into"
            walletUI -> walletServer "Forwards DApp connection approval data to"
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