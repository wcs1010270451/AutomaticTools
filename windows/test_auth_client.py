import json
import os
import tempfile
import unittest
from pathlib import Path

from auth_client import ApiClient, SessionStore, _parse_error_response


class ApiClientTests(unittest.TestCase):
    def test_base_url_has_no_trailing_slash(self) -> None:
        client = ApiClient("https://example.com/")
        self.assertEqual(client.base_url, "https://example.com")

    def test_error_response_uses_server_message(self) -> None:
        body = json.dumps(
            {"error": "验证码错误。", "requestId": "request-1"},
            ensure_ascii=False,
        ).encode("utf-8")
        self.assertEqual(
            _parse_error_response(body), ("验证码错误。", "request-1")
        )


@unittest.skipUnless(os.name == "nt", "Windows DPAPI is required")
class SessionStoreTests(unittest.TestCase):
    def test_session_round_trip_is_encrypted(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            store = SessionStore(Path(directory))
            store.save_session(
                "secret-token",
                {"id": 12, "username": "test", "email": "test@example.com"},
            )

            stored_text = store.session_path.read_text(encoding="ascii")
            self.assertNotIn("secret-token", stored_text)
            self.assertEqual(store.load_session()["token"], "secret-token")

            store.clear_session()
            self.assertIsNone(store.load_session())

    def test_device_id_is_stable(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            store = SessionStore(Path(directory))
            first = store.get_device_id()
            second = store.get_device_id()
            self.assertEqual(first, second)


if __name__ == "__main__":
    unittest.main()
