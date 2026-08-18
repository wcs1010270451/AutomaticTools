import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

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

    def test_platform_data_endpoints(self) -> None:
        client = ApiClient("https://example.com")
        with patch.object(client, "_request", return_value={}) as api_request:
            client.list_tools()
            api_request.assert_called_once_with("GET", "/api/tools")

        with patch.object(client, "_request", return_value={}) as api_request:
            client.my_purchases("token")
            api_request.assert_called_once_with(
                "GET", "/api/me/purchases", token="token"
            )

        with patch.object(client, "_request", return_value={}) as api_request:
            client.my_orders("token")
            api_request.assert_called_once_with("GET", "/api/me/orders", token="token")

        with patch.object(client, "_request", return_value={}) as api_request:
            client.redeem_license_code("token", "AT-2345-6789-ABCD-EFGH")
            api_request.assert_called_once_with(
                "POST",
                "/api/license-codes/redeem",
                {"code": "AT-2345-6789-ABCD-EFGH"},
                token="token",
            )

        with patch.object(client, "_request", return_value={}) as api_request:
            client.create_alipay_payment("token", "auto_click")
            api_request.assert_called_once_with(
                "POST",
                "/api/payments/alipay/precreate",
                {"toolCode": "auto_click"},
                token="token",
            )

        with patch.object(client, "_request", return_value={}) as api_request:
            client.payment_order_status("token", "ord_1")
            api_request.assert_called_once_with(
                "GET", "/api/payments/orders/ord_1/status", token="token"
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
