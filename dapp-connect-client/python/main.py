"""The main file, which is used for the example script."""

import keyring
from algosdk import transaction

from ducky_connect import ducky_connect as dc

keyring_prefix='DuckySignerPythonExample/'

def main() -> None:  # noqa: PLR0911, PLR0912, PLR0915
    """Run the Ducky Connect example."""
    print('')
    print('>>>>>>>>>>>>>>>>>>>>>>>><<<<<<<<<<<<<<<<<<<<<<<<')
    print('>>> Ducky Signer DApp Connect Python Example <<<')
    print('>>>>>>>>>>>>>>>>>>>>>>>><<<<<<<<<<<<<<<<<<<<<<<<')
    print('')
    print('NOTICE: Log into a wallet and turn on the dApp connect server for the wallet')  # noqa: E501
    print('')
    print('→ Generating and storing dApp connect key pair...')

    # Creating a new DuckyConnect creates a new connect key pair if no dApp connect ID
    # is given
    duckconn = dc.DuckyConnect(dc.ConnectOptions(
        dapp=dc.DappInfo(
            name='Python Connect Client Example',
            desc='An example script demonstrating the use of the Python connect client'
        ),
        keyring_prefix=keyring_prefix,
        confirm_code_display_fn=show_confirm_code,
    ))

    print(f'→ Connect ID: {duckconn.connect_id}\n')

    while True:
        continue_script = input('Continue? (Yes/no)\n> ')
        continue_script = continue_script.lower()

        if continue_script in {'', 'yes'}:
            break
        elif continue_script == 'no':
            end_example(duckconn)
            return
        else:
            continue

    print('\n→ Contacting dApp connect server...')

    try:
        session = duckconn.establish_session()
    except Exception:
        print('→ Failed to create session. ⛔')
        end_example(duckconn)
        return

    if session is None:
        print('→ Failed to create session. ⛔')
        end_example(duckconn)
        return

    print('→ Session established! 🦆')

    if len(session.session.addrs) == 0:
        print('→ No address connected in session. ⛔')
        end_example(duckconn)
        return

    print(f'→ You are connected to: {session.session.addrs[0]}\n')

    while True:
        sign_txn_ask = input('Sign a transaction? (Yes/no)\n> ')
        sign_txn_ask = sign_txn_ask.lower()

        if sign_txn_ask in {'', 'yes'}:
            break
        elif sign_txn_ask == 'no':
            end_example(duckconn)
            return
        else:
            continue

    print('\n→ Creating a 0 Algo transaction to sign...')

    # Create 0 algo transaction
    unsigned_txn = transaction.PaymentTxn(
        sender=session.session.addrs[0],
        receiver=session.session.addrs[0],
        amt=0,
        # Retrieving real suggested parameters for an example is unnecessary
        sp=transaction.SuggestedParams(
            flat_fee=True,
            fee=1000, # 0.001 Algos
            first=6000000,
            last=6001000,
            gen='testnet-v1.0',
            gh='SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=',
            min_fee=1000
        ),
    )

    print('→ Contacting dApp connect server...')

    try:
        signed_txn = duckconn.sign_transaction(unsigned_txn, '', prompt_sign_txn)
    except Exception:
        print('→ Failed to sign transaction. ⛔')
        end_example(duckconn)
        return

    if signed_txn is None:
        print('→ Failed to sign transaction. ⛔')
        end_example(duckconn)
        return

    print('→ Transaction approved! ✅')
    print('→ This is just an example. The transaction will not be sent.')
    end_example(duckconn)
    return


def show_confirm_code(code):
    """
    Show the dApp connect session confirmation code.

    :param code: Confirmation code to show
    """
    print('\nℹ️️ Go to Ducky Signer, where you should be prompted for a confirmation code')  # noqa: E501
    print(f'ℹ️ Enter this confirmation code: {code}')
    print('')
    print('→ Waiting for confirmation from Ducky Signer...')


def prompt_sign_txn():
    """Direct the user to approve signing transaction."""
    print('\nℹ️ Go to Ducky Signer, where you should be prompted to approve the transaction')  # noqa: E501
    print('')
    print('→ Waiting for response from Ducky Signer...')

def end_example(duckconn: dc.DuckyConnect):
    """Clean up before ending the example script."""
    print('→ Ending dApp connect session...')
    duckconn.end_session()

    print('→ Removing dApp connect key pair...')
    keyring.delete_password(
        keyring_prefix + duckconn.connect_id_urlsafe,
        duckconn.connect_id
    )

    print('Bye! 👋\n')


if __name__ == "__main__":
    main()
