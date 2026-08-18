import base64
import ctypes
import json
import os
import platform
import tempfile
import uuid
from ctypes import wintypes
from pathlib import Path
from typing import Any, Callable
from urllib import error, request


DEFAULT_API_BASE_URL = "https://autumnwind.top"
DEFAULT_TIMEOUT_SECONDS = 15


class ApiError(Exception):
    def __init__(self, message: str, status: int = 0, request_id: str = "") -> None:
        super().__init__(message)
        self.status = status
        self.request_id = request_id


class NetworkError(Exception):
    pass


class ApiClient:
    def __init__(
        self,
        base_url: str | None = None,
        timeout: int = DEFAULT_TIMEOUT_SECONDS,
        opener: Callable[..., Any] | None = None,
    ) -> None:
        configured_url = base_url or os.getenv(
            "AUTOMATIC_TOOLS_API_BASE_URL", DEFAULT_API_BASE_URL
        )
        self.base_url = configured_url.rstrip("/")
        self.timeout = timeout
        self._opener = opener or request.urlopen

    def send_registration_code(self, email: str) -> dict[str, Any]:
        return self._request("POST", "/api/auth/email-code", {"email": email})

    def register(
        self,
        email: str,
        email_code: str,
        password: str,
        username: str,
        device_id: str,
    ) -> dict[str, Any]:
        payload = {
            "email": email,
            "emailCode": email_code,
            "password": password,
            "deviceId": device_id,
            "deviceName": platform.node() or "Windows PC",
            "platform": "windows",
        }
        if username:
            payload["username"] = username
        return self._request("POST", "/api/auth/register", payload)

    def login(self, account: str, password: str, device_id: str) -> dict[str, Any]:
        return self._request(
            "POST",
            "/api/auth/login",
            {
                "account": account,
                "password": password,
                "deviceId": device_id,
                "deviceName": platform.node() or "Windows PC",
                "platform": "windows",
            },
        )

    def current_user(self, token: str) -> dict[str, Any]:
        return self._request("GET", "/api/me", token=token)

    def list_tools(self) -> dict[str, Any]:
        return self._request("GET", "/api/tools")

    def my_purchases(self, token: str) -> dict[str, Any]:
        return self._request("GET", "/api/me/purchases", token=token)

    def my_orders(self, token: str) -> dict[str, Any]:
        return self._request("GET", "/api/me/orders", token=token)

    def redeem_license_code(self, token: str, code: str) -> dict[str, Any]:
        return self._request(
            "POST",
            "/api/license-codes/redeem",
            {"code": code},
            token=token,
        )

    def create_alipay_payment(self, token: str, tool_code: str) -> dict[str, Any]:
        return self._request(
            "POST",
            "/api/payments/alipay/precreate",
            {"toolCode": tool_code},
            token=token,
        )

    def payment_order_status(self, token: str, order_no: str) -> dict[str, Any]:
        return self._request(
            "GET",
            f"/api/payments/orders/{order_no}/status",
            token=token,
        )

    def _request(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None = None,
        token: str = "",
    ) -> dict[str, Any]:
        body = None
        headers = {
            "Accept": "application/json",
            "User-Agent": "AutomaticTools-Windows/1.0",
        }
        if payload is not None:
            body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if token:
            headers["Authorization"] = f"Bearer {token}"

        api_request = request.Request(
            self.base_url + path,
            data=body,
            headers=headers,
            method=method,
        )
        try:
            with self._opener(api_request, timeout=self.timeout) as response:
                response_body = response.read()
        except error.HTTPError as exc:
            response_body = exc.read()
            message, request_id = _parse_error_response(response_body)
            raise ApiError(message, exc.code, request_id) from exc
        except (error.URLError, TimeoutError, OSError) as exc:
            raise NetworkError("无法连接服务器，请检查网络后重试。") from exc

        if not response_body:
            return {}
        try:
            result = json.loads(response_body.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError) as exc:
            raise ApiError("服务器返回了无法识别的数据。") from exc
        if not isinstance(result, dict):
            raise ApiError("服务器返回了无法识别的数据。")
        return result


def _parse_error_response(body: bytes) -> tuple[str, str]:
    try:
        payload = json.loads(body.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError):
        return "服务器请求失败，请稍后重试。", ""
    if not isinstance(payload, dict):
        return "服务器请求失败，请稍后重试。", ""
    message = payload.get("error")
    request_id = payload.get("requestId")
    return (
        message if isinstance(message, str) and message else "服务器请求失败，请稍后重试。",
        request_id if isinstance(request_id, str) else "",
    )


class _DataBlob(ctypes.Structure):
    _fields_ = [
        ("cbData", wintypes.DWORD),
        ("pbData", ctypes.POINTER(ctypes.c_ubyte)),
    ]


def _protect_data(data: bytes) -> bytes:
    if os.name != "nt":
        raise OSError("Windows DPAPI is unavailable on this system")
    source_buffer = ctypes.create_string_buffer(data)
    source = _DataBlob(
        len(data), ctypes.cast(source_buffer, ctypes.POINTER(ctypes.c_ubyte))
    )
    protected = _DataBlob()
    crypt32 = ctypes.windll.crypt32
    if not crypt32.CryptProtectData(
        ctypes.byref(source),
        "AutomaticTools login session",
        None,
        None,
        None,
        0,
        ctypes.byref(protected),
    ):
        raise ctypes.WinError(ctypes.get_last_error())
    try:
        return ctypes.string_at(protected.pbData, protected.cbData)
    finally:
        ctypes.windll.kernel32.LocalFree(protected.pbData)


def _unprotect_data(data: bytes) -> bytes:
    if os.name != "nt":
        raise OSError("Windows DPAPI is unavailable on this system")
    source_buffer = ctypes.create_string_buffer(data)
    source = _DataBlob(
        len(data), ctypes.cast(source_buffer, ctypes.POINTER(ctypes.c_ubyte))
    )
    plain = _DataBlob()
    crypt32 = ctypes.windll.crypt32
    if not crypt32.CryptUnprotectData(
        ctypes.byref(source),
        None,
        None,
        None,
        None,
        0,
        ctypes.byref(plain),
    ):
        raise ctypes.WinError(ctypes.get_last_error())
    try:
        return ctypes.string_at(plain.pbData, plain.cbData)
    finally:
        ctypes.windll.kernel32.LocalFree(plain.pbData)


class SessionStore:
    def __init__(self, data_dir: Path | None = None) -> None:
        if data_dir is None:
            app_data = os.getenv("APPDATA")
            data_dir = (
                Path(app_data) / "AutomaticTools"
                if app_data
                else Path.home() / ".automatic_tools"
            )
        self.data_dir = Path(data_dir)
        self.session_path = self.data_dir / "session.dat"
        self.client_path = self.data_dir / "client.json"

    def get_device_id(self) -> str:
        try:
            payload = json.loads(self.client_path.read_text(encoding="utf-8"))
            value = payload.get("deviceId")
            if isinstance(value, str) and value:
                return value
        except (OSError, json.JSONDecodeError, AttributeError):
            pass

        device_id = str(uuid.uuid4())
        self._write_text_atomic(
            self.client_path,
            json.dumps({"deviceId": device_id}, ensure_ascii=False),
        )
        return device_id

    def save_session(self, token: str, user: dict[str, Any]) -> None:
        payload = json.dumps(
            {"token": token, "user": user}, ensure_ascii=False
        ).encode("utf-8")
        protected = base64.b64encode(_protect_data(payload)).decode("ascii")
        self._write_text_atomic(self.session_path, protected)

    def load_session(self) -> dict[str, Any] | None:
        try:
            protected = base64.b64decode(
                self.session_path.read_text(encoding="ascii"), validate=True
            )
            payload = json.loads(_unprotect_data(protected).decode("utf-8"))
        except (OSError, ValueError, UnicodeDecodeError, json.JSONDecodeError):
            return None
        if not isinstance(payload, dict):
            return None
        token = payload.get("token")
        user = payload.get("user")
        if not isinstance(token, str) or not token or not isinstance(user, dict):
            return None
        return payload

    def clear_session(self) -> None:
        try:
            self.session_path.unlink()
        except FileNotFoundError:
            pass

    def _write_text_atomic(self, path: Path, content: str) -> None:
        self.data_dir.mkdir(parents=True, exist_ok=True)
        file_descriptor, temporary_name = tempfile.mkstemp(
            prefix=path.name + ".", dir=self.data_dir, text=True
        )
        try:
            with os.fdopen(file_descriptor, "w", encoding="utf-8") as temporary_file:
                temporary_file.write(content)
            os.replace(temporary_name, path)
        except Exception:
            try:
                os.unlink(temporary_name)
            except OSError:
                pass
            raise
