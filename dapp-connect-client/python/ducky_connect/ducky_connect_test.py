"""Tests for Ducky Connect SDK."""

import pytest
from pytest_check import check

m = pytest.mark

class TestDCEstablishSession:
    """DappConnect.establish_session()"""

    @m.skip
    def test_create(self):
        """Creates a new session and returns the session data"""
        pass

    @m.skip
    def test_init_fail(self):
        """Throws error when session initialization fails"""
        pass

    @m.skip
    def test_confirm_fail(self):
        """Throws error when session confirmation fails"""
        pass

class TestDCRetrieveSession:
    """DappConnect.retrieve_session()"""

    @m.skip
    def test_data_exists(self):
        """Returns stored session data, if it exists"""
        pass

    @m.skip
    def test_data_not_exist(self):
        """Returns null if there is no stored session data"""
        pass

class TestDCEndSession:
    """DappConnect.end_session()"""

    @m.skip
    def test_server_contact_success(self):
        """Removes stored session after successfully contacting server"""
        pass

    @m.skip
    def test_server_contact_fail(self):
        """Still removes stored session after contacting server fails"""
        pass

class TestDCSignTransaction:
    """DappConnect.sign_transaction()"""

    @m.skip
    def test_no_signer(self):
        """Signs transaction WITHOUT signer address given"""
        pass

    @m.skip
    def test_with_signer(self):
        """Signs transaction WITH signer address given"""
        pass

    @m.skip
    def test_signing_fail(self):
        """Throws error if transaction signing fails"""
        pass

    @m.skip
    def test_no_session(self):
        """Throws error if no session has been established"""
        pass
