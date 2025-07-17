workspace {

    model {
        user = person "User" "A user of the Ducky Signer desktop wallet"
        wallet = softwareSystem "Ducky Signer" "Algorand wallet software for desktop computers" {
            walletUI = container "User Interface"
            connectServer = container "DApp Connect Server" "Allows dApp to connect to the wallet"
            kmd = container "Key Management Daemon (KMD)" "Manages the private keys for the user's Algorand accounts"
            acctKeyStore = container "Account Key Store" "Encrypted storage of the private keys for the user's Algorand accounts" {
                tags Database
            }
            sessionManager = container "Session Manager" "Manages the dApp connect sessions & the session store"
            sessionStore = container "DApp Connect Session Store" "Storage for connect session key pairs and data" {
                tags Database
            }
            settingsManager = container "Settings Manager" "Manages the user's wallet settings & the settings store"
            settingsStore = container "Settings Store" "Storage for the user's wallet settings" {
                tags Database
            }
        }
        dapp1 = softwareSystem "DApp #1" "Typical web app using Algorand" {
            tags DApp External
        }
        dapp2 = softwareSystem "DApp #2" "Another app using Algorand, but is not in the browser" {
            tags NonBrowserDApp External
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

        ### System landscape relationships ###

        user -> wallet "Uses"

        user -> dapp1 "Interacts with"
        dapp1 -> user "Responds to"
        dapp1 -> wallet "Communicates with" "HTTP/2"
        dapp1 -> algoNode "Interacts with"

        user -> dapp2 "Also interacts with"
        dapp2 -> user "Also responds to"
        dapp2 -> wallet "Also communicates with" "HTTP/2"
        dapp2 -> algoNode "Also interacts with"

        algoNode -> algorand "Connected to"
        wallet -> algoNode "Pull blockchain data using"

        user -> ledgerDevice "Signs using"
        ledgerDevice -> user "Requests approval to sign from"

        ### Wallet relationships ###

        user -> walletUI "Interfaces with"
        walletUI -> user "Requests input from"

        dapp1 -> connectServer "Makes requests to" "HTTP/2"
        dapp2 -> connectServer "Also makes requests to" "HTTP/2"
        walletUI -> connectServer "Send user response to" "Event"
        connectServer -> walletUI "Request action from user through" "Event"

        walletUI -> kmd "Manages user's Algorand accounts using"
        kmd -> connectServer "Provides data about user's Algorand accounts to"
        kmd -> walletUI "Provides data about user's Algorand accounts to"
        kmd -> acctKeyStore "Stores private keys into & retrieves the from" "With password from user"
        kmd -> ledgerDevice "Sends signing request to"
        ledgerDevice -> kmd "Sends signed data to"

        walletUI -> settingsManager "Manages user preferences using"
        settingsManager -> walletUI "Provides data about user's preferences to"
        connectServer -> settingsManager "Accesses user preferences through"
        settingsManager -> settingsStore "Stores & retrieves user preferences from" "With password from user"

        walletUI -> sessionManager "Manages dApp connect sessions using"
        sessionManager -> walletUI "Provides data about dApp connect sessions to"
        connectServer -> sessionManager "Manages dApp connect sessions through"
        sessionManager -> connectServer "Provides data about dApp connect sessions to"
        sessionManager -> sessionStore "Stores & retrieves session keys and data from" "With password from user"
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

        dynamic wallet DAppConnectSessionEstablishment "DApp establishes connect session with wallet" {
            autoLayout
            # Initialization
            user -> dapp1 "Needs to connect to wallet within"
            dapp1 -> connectServer "Sends request to initialize session to"
            connectServer -> sessionManager "Generates confirmation key pair using"
            sessionManager -> sessionStore "Stores generated key pair into"
            connectServer -> dapp1 "Responds with confirmation token, code and data to"
            # Confirmation
            dapp1 -> user "Displays confirmation code to"
            dapp1 -> connectServer "Sends request to confirm session to"
            connectServer -> sessionManager "Attempts to get confirmation key using"
            sessionManager -> sessionStore "Retrieves confirmation key from"
            sessionManager -> connectServer "Returns with confirmation key to"
            connectServer -> walletUI "Request for user approval through"
            walletUI -> user "Asks for approval from"
            user -> walletUI "Approves connection by entering wallet password & confirmation code into"
            walletUI -> connectServer "Forwards user approval to"
            connectServer -> sessionManager "Extracts session key pair from confirmation token using"
            sessionManager -> sessionStore "Stores generated key pair into"
            connectServer -> dapp1 "Responds with session ID & data to"
            dapp1 -> user "Shows it is connected to wallet to"
        }

        dynamic wallet DappConnectDisconnectThroughDApp "User terminates connect session though dApp (preferred method)" {
            autoLayout
            user -> dapp1 "Initiates session termination within"
            dapp1 -> connectServer "Sends request to terminate session to"
            connectServer -> sessionManager "Removes session data for dApp using"
            sessionManager -> sessionStore "Delete session data from"
            connectServer -> dapp1 "Responds with success message to"
            dapp1 -> user "Deletes its session data & shows it is disconnected from wallet to"
        }

        dynamic wallet DappConnectDisconnectThroughDAppAlt "User terminates connect session though dApp (alternate method)" {
            autoLayout
            user -> dapp1 "Initiates session termination within"
            dapp1 -> user "Deletes its session data & shows it is disconnected from wallet to"
        }

        dynamic wallet DappConnectDisconnectThroughWallet "User terminates connect session though wallet" {
            autoLayout
            user -> walletUI "Initiates termination of session with dApp within"
            walletUI -> connectServer "Forwards request to terminate dApp's session to"
            connectServer -> sessionManager "Removes session data for dApp using"
            sessionManager -> sessionStore "Delete session data from"
            connectServer -> walletUI "Responds with success message to"
            walletUI -> user "Shows it is disconnected from dApp to"
        }

        dynamic wallet SignTransaction "DApp requests user to sign a transaction" {
            autoLayout
            user -> dapp1 "Needs to sign a transaction within"
            dapp1 -> connectServer "Sends request to sign transaction to"
            connectServer -> walletUI "Forwards request to sign transaction to"
            walletUI -> user "Asks for approval to sign transaction from"
            user -> walletUI "Approves signing transaction using"
            walletUI -> kmd "Attempt to sign transaction using"
            kmd -> acctKeyStore "Retrieves key for signing from"
            kmd -> walletUI "Signs transaction & returns signed transaction data to"
            walletUI -> connectServer "Forwards signed transaction data to"
            connectServer -> dapp1 "Responds with signed transaction data to"
            dapp1 -> algoNode "Sends signed transaction using"
        }

        dynamic wallet SignTransactionWithLedger "DApp requests user (with Ledger device) to sign a transaction" {
            autoLayout
            user -> dapp1 "Needs to sign a transaction within"
            dapp1 -> connectServer "Sends request to sign transaction to"
            connectServer -> walletUI "Forwards request to sign transaction to"
            walletUI -> kmd "Sends unsigned transaction data to"
            kmd -> ledgerDevice "Forward unsigned transaction data to"
            walletUI -> user "Asks for approval to sign transaction from"
            user -> ledgerDevice "Signs transaction using"
            ledgerDevice -> kmd "Sends signed transaction data to"
            kmd -> walletUI "Forwards signed transaction data to"
            walletUI -> connectServer "Forwards signed transaction data to"
            connectServer -> dapp1 "Responds with signed transaction data to"
            dapp1 -> algoNode "Sends signed transaction using"
        }

        theme default

        styles {
            element External {
                background #e0d5e5
                color #222222
            }
            element DApp {
                shape WebBrowser
            }
            element NonBrowserDApp {
                shape Robot
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
