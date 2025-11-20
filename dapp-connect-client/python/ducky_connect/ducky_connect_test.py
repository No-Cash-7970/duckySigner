"""Tests for Ducky Connect SDK."""

from datetime import datetime

import algosdk
import pytest
from pytest_check import check

import ducky_connect as dc

m = pytest.mark

class TestDCEstablishSession:
    """DappConnect.establish_session()"""

    @m.skip
    def test_create(self, requests_mock):
        """Creates a new session and returns the session data"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={
                'id': 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
                'code': '88048',
                'token': 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',  # noqa: E501
                'exp': 1760003247,
            },
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={
                'id': 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
                'exp': 1760591204,
                'addrs': ('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'),
            },
        )

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))
        session = duckconn.establish_session()

        check.equal(session.id, 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
            'Established session has correct session ID')
        check.equal(session.exp.timestamp(), 1760591204,
            'Established session has correct expiration data-time')
        check.equal(session.addrs,
            ('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'),
            'Established session has correct list of addresses')
        check.is_not_none(duckconn.retrieve_session(), 'Established session is stored')

    @m.skip
    def test_init_fail(self, requests_mock):
        """Throws error when session initialization fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={
                'id': 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
                'exp': 1760591204,
                'addrs': ('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'),
            },
        )

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))

        with check.raises(Exception):
            duckconn.establish_session()

    @m.skip
    def test_confirm_fail(self, requests_mock):
        """Throws error when session confirmation fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={
                'id': 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
                'code': '88048',
                'token': 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',  # noqa: E501
                'exp': 1760003247,
            },
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))

        with check.raises(Exception):
            duckconn.establish_session()

class TestDCRetrieveSession:
    """DappConnect.retrieve_session()"""

    @m.skip
    def test_data_exists(self):
        """Returns stored session data, if it exists"""
        info_to_be_stored = dc.StoredSessionInfo(
            connect_id='7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
            session=dc.SessionInfo(
                id='XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
                exp=datetime.fromtimestamp(1760591204),
                addrs=('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'),
            ),
            dapp=dc.DappInfo(name='Test DApp'),
        )

        # TODO: Set stored session info into a file

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo(name='')))
        storedSession: dc.StoredSessionInfo = duckconn.retrieve_session()

        check.is_not_none(storedSession, 'Stored session information was retrieved')
        check.equal(storedSession.connect_id, info_to_be_stored.connect_id,
            'Stored session information has correct connect ID')
        check.equal(storedSession.session.id, info_to_be_stored.session.id,
            'Stored session information has correct session ID')
        check.equal(storedSession.session.exp, info_to_be_stored.session.exp,
            'Stored session information has correct session expiration date-time')
        check.equal(storedSession.session.addrs, info_to_be_stored.session.addrs,
            'Stored session information has correct list of addresses')
        check.equal(storedSession.dapp.name, info_to_be_stored.dapp.name,
            'Stored session information has correct dApp name')
        # TODO: Remove session info file

    @m.skip
    def test_data_not_exist(self):
        """Returns null if there is no stored session data"""
        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo(name='')))
        storedSession: dc.StoredSessionInfo = duckconn.retrieve_session()
        check.is_not_none(storedSession,
            'There was no stored session information to be retrieved')

class TestDCEndSession:
    """DappConnect.end_session()"""

    @m.skip
    def test_server_contact_success(self, requests_mock):
        """Removes stored session after successfully contacting server"""
        requests_mock.get(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_END_ENDPOINT}',
            text='OK'
        )

        # TODO: Remove stored session info from storage

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))
        duckconn.end_session()

        check.is_none(duckconn.retrieve_session(),
            'Stored session information has been removed')

    @m.skip
    def test_server_contact_fail(self, requests_mock):
        """Still removes stored session after contacting server fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_END_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # TODO: Remove stored session info from storage

        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))
        duckconn.end_session()

        check.is_none(duckconn.retrieve_session(),
            'Stored session information has been removed')

class TestDCSignTransaction:
    """DappConnect.sign_transaction()"""

    @m.skip
    def test_no_signer(self, requests_mock):
        """Signs transaction WITHOUT signer address given"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5lduckconn12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
        )

        # TODO: Set stored session info into a file

        test_txn = algosdk.transaction.PaymentTxn(
            sender='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            receiver='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            amt=5_000_000,
            sp=algosdk.transaction.SuggestedParams(
                fee=1000, # 0.001 Algos
                first=6000000,
                last=6001000,
                gen='testnet-v1.0',
                gh='SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=',
                min_fee=1000
            ),
        )
        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))
        signed_txn = duckconn.sign_transaction(test_txn)

        check.equal(signed_txn.transaction.sender, test_txn.sender)
        check.equal(signed_txn.transaction.amt, 5_000_000)
        # TODO: Remove session info file

    @m.skip
    def test_with_signer(self, requests_mock):
        """Signs transaction WITH signer address given"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5lduckconn12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
        )

        # TODO: Set stored session info into a file

        test_txn = algosdk.transaction.PaymentTxn(
            sender='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            receiver='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            amt=5_000_000,
            sp=algosdk.transaction.SuggestedParams(
                fee=1000, # 0.001 Algos
                first=6000000,
                last=6001000,
                gen='testnet-v1.0',
                gh='SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=',
                min_fee=1000
            ),
        )
        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))
        signed_txn = duckconn.sign_transaction(
            test_txn,
            'VCMJKWOY5P5P7SKMZFFOCEROPJCZOTIJMNIYNUCKH7LRO45JMJP6UYBIJA'
        )

        check.equal(signed_txn.transaction.sender, test_txn.sender)
        check.equal(signed_txn.transaction.amt, 5_000_000)
        # TODO: Remove session info file

    @m.skip
    def test_signing_fail(self, requests_mock):
        """Throws error if transaction signing fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # TODO: Set stored session info into a file

        test_txn = algosdk.transaction.PaymentTxn(
            sender='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            receiver='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            amt=5_000_000,
            sp=algosdk.transaction.SuggestedParams(
                fee=1000, # 0.001 Algos
                first=6000000,
                last=6001000,
                gen='testnet-v1.0',
                gh='SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=',
                min_fee=1000
            ),
        )
        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))

        with check.raises(Exception):
            duckconn.sign_transaction(test_txn)

        # TODO: Remove session info file

    @m.skip
    def test_no_session(self, requests_mock):
        """Throws error if no session has been established"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5lduckconn12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
        )

        test_txn = algosdk.transaction.PaymentTxn(
            sender='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            receiver='RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
            amt=5_000_000,
            sp=algosdk.transaction.SuggestedParams(
                fee=1000, # 0.001 Algos
                first=6000000,
                last=6001000,
                gen='testnet-v1.0',
                gh='SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=',
                min_fee=1000
            ),
        )
        duckconn = dc.DuckyConnect(dc.ConnectOptions(dapp=dc.DappInfo('Test DApp')))

        with check.raises(Exception):
            duckconn.sign_transaction(test_txn)
