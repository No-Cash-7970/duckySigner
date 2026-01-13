"""Tests for Ducky Connect SDK."""

import json
import pathlib

import algosdk
import keyring
import pytest
from pytest_check import check
from requests.exceptions import HTTPError

from . import ducky_connect as dc

m = pytest.mark

class TestDCEstablishSession:
    """DappConnect.establish_session()"""

    test_dapp = dc.DappInfo(
        name='Test DApp 1',
        url=dc.DEFAULT_SERVER_BASE_URL,
        desc='This is not a real dApp. It is just a test.'
    )

    def test_create(self, requests_mock, mocker):
        """Creates a new session and returns the session data"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={
                'id': 'B8Q1qnt1nk5FVmojPRouexpudIQ7kIyljRE2j8Q4uwQ=',
                'code': '72725',
                'token': 'v4.local.3wnLGNMqWfIMjVbbehvD3jAkmTbL7gI6LgOQDrrFAvSfOqRTzNTPo8JDJcINKOYnXXXjs3AY7heqHi2aRYmyQldo5knnqF3WytQqhX-YZq79HE3QveDopvHu0ZzZ4CHFX9W0pHcif5ot-U0luIduj5zWCe0yxIkJbS3fCJTLkpswsGuG9CO-Ub1M9Qs57Ra73jhSV6ihbgXG4LT5NiE20h28Bt954za-eVo_6pXiF2Ep_BxSjQRAet_uF4NfNNl7_Y4pE9MbP2AVDC62HbtN6IEmWwRLzWaVxtzouD8-',  # noqa: E501
                'exp': 1767849999,
            },
            headers={ 'Content-Type': 'application/json' },
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={
                'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                'exp': 1768454311,
                'addrs': [],
            },
            headers={
                'Content-Type': 'application/json',
                'Server-Authorization': 'fake_header_that_should_be_ignored',
            },
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionCreate/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=test_keyring_prefix + test_connect_id_b64,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_create'

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        code_display_fn_spy = mocker.spy(connect_options, 'confirm_code_display_fn')
        duckconn = dc.DuckyConnect(connect_options)
        stored_session = duckconn.establish_session()

        check.equal(code_display_fn_spy.call_count, 1,
            'The confirm code was displayed once')
        check.equal(stored_session.session.id,
            'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
            'Established session has correct session ID')
        check.equal(stored_session.session.exp.timestamp(), 1768454311,
            'Established session has correct expiration data-time')
        check.equal(stored_session.session.addrs, [],
            'Established session has correct list of addresses')
        # TODO
        # check.is_not_none(duckconn.retrieve_session(),
        #     'Established session is stored')

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_init_fail(self, requests_mock):
        """Throws error when session initialization fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={
                'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                'exp': 1768454311,
                'addrs': [],
            },
            headers={
                'Content-Type': 'application/json',
                'Server-Authorization': 'fake_header_that_should_be_ignored',
            },
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionInitFail/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=test_keyring_prefix + test_connect_id_b64,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_init_fail'

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        with check.raises(HTTPError):
            duckconn.establish_session()

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_confirm_fail(self, requests_mock):
        """Throws error when session confirmation fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_INIT_ENDPOINT}',
            json={
                'id': 'B8Q1qnt1nk5FVmojPRouexpudIQ7kIyljRE2j8Q4uwQ=',
                'code': '72725',
                'token': 'v4.local.3wnLGNMqWfIMjVbbehvD3jAkmTbL7gI6LgOQDrrFAvSfOqRTzNTPo8JDJcINKOYnXXXjs3AY7heqHi2aRYmyQldo5knnqF3WytQqhX-YZq79HE3QveDopvHu0ZzZ4CHFX9W0pHcif5ot-U0luIduj5zWCe0yxIkJbS3fCJTLkpswsGuG9CO-Ub1M9Qs57Ra73jhSV6ihbgXG4LT5NiE20h28Bt954za-eVo_6pXiF2Ep_BxSjQRAet_uF4NfNNl7_Y4pE9MbP2AVDC62HbtN6IEmWwRLzWaVxtzouD8-',  # noqa: E501
                'exp': 1767849999,
            },
            headers={ 'Content-Type': 'application/json' },
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_CONFIRM_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionConfirmFail/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=test_keyring_prefix + test_connect_id_b64,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_confirm_fail'

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        with check.raises(HTTPError):
            duckconn.establish_session()

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

class TestDCLoadSession:
    """DappConnect.load_session()"""

    def test_data_exists(self):
        """Returns stored session data, if it exists"""
        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestGetSessionExist/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=test_keyring_prefix + test_connect_id_b64,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_get_exist'

        # Write session information to a file
        info_to_be_stored = {
            'connect_id': test_connect_id,
            'session': {
                'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                'exp': 1768454311,
                'addrs': [],
            },
            'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
            'server_url': dc.DEFAULT_SERVER_BASE_URL,
        }
        with open(test_session_file, 'w') as f:
            f.write(json.dumps(info_to_be_stored))

        connect_options = dc.ConnectOptions(
            dapp=dc.DappInfo(name='Test DApp'),
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        storedSession: dc.StoredSessionInfo = duckconn.load_session()

        check.is_not_none(storedSession, 'Stored session information was retrieved')
        check.equal(storedSession.connect_id, info_to_be_stored['connect_id'],
            'Stored session information has correct connect ID')
        check.equal(storedSession.session.id, info_to_be_stored['session']['id'],
            'Stored session information has correct session ID')
        check.equal(int(storedSession.session.exp.timestamp()),
            info_to_be_stored['session']['exp'],
            'Stored session information has correct session expiration date-time')
        check.equal(storedSession.session.addrs, info_to_be_stored['session']['addrs'],
            'Stored session information has correct list of addresses')
        check.equal(storedSession.dapp.name, info_to_be_stored['dapp']['name'],
            'Stored session information has correct dApp name')

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_data_not_exist(self):
        """Returns None if there is no stored session data"""
        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestGetSessionExist/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=test_keyring_prefix + test_connect_id_b64,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_not_exist'

        connect_options = dc.ConnectOptions(
            dapp=dc.DappInfo(name='Test DApp'),
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        storedSession: dc.StoredSessionInfo = duckconn.load_session()
        check.is_none(storedSession,
            'There was no stored session information to be retrieved')

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

class TestDCEndSession:
    """DappConnect.end_session()"""

    test_dapp = dc.DappInfo(
        name='Test DApp 2',
        url=dc.DEFAULT_SERVER_BASE_URL,
        desc='This is not a real dApp. It is just a test.'
    )

    def test_server_contact_success(self, requests_mock):
        """Removes stored session after successfully contacting server"""
        requests_mock.get(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_END_ENDPOINT}',
            text='OK'
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionEndContact/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_end_contact'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)
        duckconn.end_session()

        check.is_false(pathlib.Path(test_session_file).exists(),
            "Session file has been removed")
        check.is_not_none(keyring.get_password(keyring_service_name, test_connect_id),
            "Connect key pair is still stored")

        ### Clean up ###
        # Remove keyring entry
        keyring.delete_password(keyring_service_name, duckconn.connect_id)

    def test_server_contact_fail(self, requests_mock):
        """Still removes stored session after contacting server fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_END_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionEndContactFail/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_end_contact_fail'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)
        duckconn.end_session()

        check.is_false(pathlib.Path(test_session_file).exists(),
            "Session file has been removed")
        check.is_not_none(keyring.get_password(keyring_service_name, test_connect_id),
            "Connect key pair is still stored")

        ### Clean up ###
        # Remove keyring entry
        keyring.delete_password(keyring_service_name, duckconn.connect_id)

    def test_no_session(self, requests_mock):
        """Does not fail if there is no session"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SESSION_END_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionEndNone/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_end_none'

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)
        duckconn.load_session()
        duckconn.end_session()

        check.is_false(pathlib.Path(test_session_file).exists(),
            "There is still no session file")
        check.is_not_none(keyring.get_password(keyring_service_name, test_connect_id),
            "Connect key pair is still stored")

        ### Clean up ###
        # Remove keyring entry
        keyring.delete_password(keyring_service_name, duckconn.connect_id)

    def test_no_server_contact(self):
        """Does not attempt to contact the server if specified not to do so"""
        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionEndContactFail/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_end_contact_fail'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,

            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)
        duckconn.end_session(True)

        check.is_false(pathlib.Path(test_session_file).exists(),
            "Session file has been removed")
        check.is_not_none(keyring.get_password(keyring_service_name, test_connect_id),
            "Connect key pair is still stored")

        ### Clean up ###
        # Remove keyring entry
        keyring.delete_password(keyring_service_name, duckconn.connect_id)

class TestDCSignTransaction:
    """DappConnect.sign_transaction()"""

    test_dapp = dc.DappInfo(
        name='Test DApp 3',
        url=dc.DEFAULT_SERVER_BASE_URL,
        desc='This is not a real dApp. It is just a test.'
    )

    def test_no_signer(self, requests_mock):
        """Signs transaction WITHOUT signer address given"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
            headers={
                'Content-Type': 'application/json',
                'Server-Authorization': 'fake_header_that_should_be_ignored',
            },
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionSignTxn/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_sign_txn'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        test_txn = algosdk.transaction.PaymentTxn(
            sender='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
            receiver='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
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
        signed_txn = duckconn.sign_transaction(test_txn, '', lambda: None)

        check.equal(signed_txn.transaction.sender, test_txn.sender)
        check.equal(signed_txn.transaction.amt, 1_000_000)

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_with_signer(self, requests_mock):
        """Signs transaction WITH signer address given"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5lduckconn12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
        )
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
            headers={
                'Content-Type': 'application/json',
                'Server-Authorization': 'fake_header_that_should_be_ignored',
            },
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionSignTxn/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_sign_txn'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        test_txn = algosdk.transaction.PaymentTxn(
            sender='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
            receiver='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
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
        signed_txn = duckconn.sign_transaction(test_txn,
            'VCMJKWOY5P5P7SKMZFFOCEROPJCZOTIJMNIYNUCKH7LRO45JMJP6UYBIJA',
            lambda: None
        )

        check.equal(signed_txn.transaction.sender, test_txn.sender)
        check.equal(signed_txn.transaction.amt, 1_000_000)

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_signing_fail(self, requests_mock):
        """Throws error if transaction signing fails"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={'name': 'oops', 'message': 'Something went wrong!'},
            status_code=500,
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionSignTxn/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_sign_txn'

        # Write session information to a file
        with open(test_session_file, 'w') as f:
            f.write(json.dumps({
                'connect_id': test_connect_id,
                'session': {
                    'id': 'a7Qze6UXeRIz0sYWzn1aFEujqKj2RS+l+M5K3T71mBk=',
                    'exp': 1768454311,
                    'addrs': [],
                },
                'dapp': {'name': 'Test DApp', 'url': '', 'desc': '', 'icon': ''},
                'server_url': dc.DEFAULT_SERVER_BASE_URL,
            }))

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        test_txn = algosdk.transaction.PaymentTxn(
            sender='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
            receiver='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
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

        with check.raises(HTTPError):
            duckconn.sign_transaction(test_txn, '', lambda: None)

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )

    def test_no_session(self, requests_mock):
        """Throws error if no session has been established"""
        requests_mock.post(f'{dc.DEFAULT_SERVER_BASE_URL}{dc.SIGN_TXN_ENDPOINT}',
            json={
                'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5lduckconn12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',  # noqa: E501
            },
        )

        # Add test connect key pair to keyring (temporary)
        test_keyring_prefix = 'DuckySignerTestSessionSignTxn/'
        test_connect_id = 'L9PAqGp+UWyf5nRVyr/grVHagPjr+IdCgD0s2TNfJTs='
        test_connect_id_b64 = test_connect_id.replace('+', '-').replace('/', '_')
        keyring_service_name = test_keyring_prefix + test_connect_id_b64
        test_connect_key = 'EMgkOHnGurQ2f3FvteoGwcLvKcQIYvJQvaLk1n//ekA='
        keyring.set_password(
            service_name=keyring_service_name,
            username=test_connect_id,
            password=test_connect_key,
        )

        test_session_file = './.test_dc_session_sign_txn'

        connect_options = dc.ConnectOptions(
            dapp=self.test_dapp,
            session_file_path=test_session_file,
            keyring_prefix=test_keyring_prefix,
            connect_id=test_connect_id,
            allow_insecure_responses=True,
            confirm_code_display_fn=lambda code: None
        )
        duckconn = dc.DuckyConnect(connect_options)

        test_txn = algosdk.transaction.PaymentTxn(
            sender='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
            receiver='KU5F7QXHK2BE7EU3HPVZTMKS2V4HMIADHZVGRQC2PBL6IWYIE43HVWMFUA',
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

        with check.raises(RuntimeError):
            duckconn.sign_transaction(test_txn, '', lambda: None)

        ### Clean up ###
        # Remove session file
        pathlib.Path(test_session_file).unlink(missing_ok=True)
        # Remove keyring entry
        keyring.delete_password(
            test_keyring_prefix + test_connect_id_b64,
            duckconn.connect_id
        )
