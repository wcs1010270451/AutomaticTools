import ctypes
import re
import sys
import threading
import time
import tkinter as tk
from ctypes import wintypes
from datetime import datetime
from pathlib import Path
from tkinter import messagebox, ttk

import qrcode
from PIL import Image, ImageTk

from auth_client import ApiClient, ApiError, NetworkError, SessionStore


user32 = ctypes.WinDLL("user32", use_last_error=True)


def apply_window_icon(window: tk.Misc) -> None:
    base_path = Path(getattr(sys, "_MEIPASS", Path(__file__).resolve().parent))
    icon_path = base_path / "assets" / "automatictools.png"
    try:
        icon_image = tk.PhotoImage(file=str(icon_path))
        window.iconphoto(True, icon_image)
        window._automatictools_icon = icon_image
    except (OSError, tk.TclError):
        pass

VK_F8 = 0x77
VK_F9 = 0x78
VK_F6 = 0x75
VK_F7 = 0x76
VK_CONTROL = 0x11
VK_1 = 0x31
VK_2 = 0x32
VK_3 = 0x33
VK_4 = 0x34
MOUSEEVENTF_LEFTDOWN = 0x0002
MOUSEEVENTF_LEFTUP = 0x0004
WM_LBUTTONDOWN = 0x0201
WM_LBUTTONUP = 0x0202
MK_LBUTTON = 0x0001


class POINT(ctypes.Structure):
    _fields_ = [("x", wintypes.LONG), ("y", wintypes.LONG)]


user32.WindowFromPoint.argtypes = [POINT]
user32.WindowFromPoint.restype = wintypes.HWND
user32.ScreenToClient.argtypes = [wintypes.HWND, ctypes.POINTER(POINT)]
user32.ScreenToClient.restype = wintypes.BOOL
user32.PostMessageW.argtypes = [
    wintypes.HWND,
    wintypes.UINT,
    wintypes.WPARAM,
    wintypes.LPARAM,
]
user32.PostMessageW.restype = wintypes.BOOL


def key_pressed(vk_code: int) -> bool:
    return bool(user32.GetAsyncKeyState(vk_code) & 0x0001)


def key_down(vk_code: int) -> bool:
    return bool(user32.GetAsyncKeyState(vk_code) & 0x8000)


def get_cursor_position() -> tuple[int, int]:
    point = POINT()
    if not user32.GetCursorPos(ctypes.byref(point)):
        raise ctypes.WinError(ctypes.get_last_error())
    return point.x, point.y


def click_at(x: int, y: int) -> None:
    if not user32.SetCursorPos(x, y):
        raise ctypes.WinError(ctypes.get_last_error())
    user32.mouse_event(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0)
    time.sleep(0.03)
    user32.mouse_event(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0)


def window_click_at(hwnd: int, client_x: int, client_y: int) -> None:
    lparam = (client_y << 16) | (client_x & 0xFFFF)
    if not user32.PostMessageW(hwnd, WM_LBUTTONDOWN, MK_LBUTTON, lparam):
        raise ctypes.WinError(ctypes.get_last_error())
    time.sleep(0.03)
    if not user32.PostMessageW(hwnd, WM_LBUTTONUP, 0, lparam):
        raise ctypes.WinError(ctypes.get_last_error())


def get_window_target(screen_x: int, screen_y: int) -> tuple[int, int, int]:
    point = POINT(screen_x, screen_y)
    hwnd = user32.WindowFromPoint(point)
    if not hwnd:
        raise ctypes.WinError(ctypes.get_last_error())
    client_point = POINT(screen_x, screen_y)
    if not user32.ScreenToClient(hwnd, ctypes.byref(client_point)):
        raise ctypes.WinError(ctypes.get_last_error())
    return hwnd, client_point.x, client_point.y


class AuthenticationWindow:
    BACKGROUND = "#f3f5f8"
    SURFACE = "#fbfcfe"
    TEXT = "#18212b"
    MUTED = "#667281"
    BORDER = "#d7dde7"
    ACCENT = "#2563eb"
    ACCENT_HOVER = "#1d4ed8"
    SUCCESS = "#0f766e"
    ERROR = "#b42318"

    def __init__(self, api: ApiClient, session_store: SessionStore) -> None:
        self.api = api
        self.session_store = session_store
        self.session: dict | None = None
        self.mode = "login"
        self.busy = False
        self.code_available_at = 0.0
        self.code_timer_id: str | None = None

        self.root = tk.Tk()
        apply_window_icon(self.root)
        self.root.title("AutomaticTools - 登录")
        self.root.configure(bg=self.BACKGROUND)
        self.root.resizable(False, False)
        self.root.protocol("WM_DELETE_WINDOW", self.close)
        self._configure_styles()

        self.account_var = tk.StringVar()
        self.login_password_var = tk.StringVar()
        self.email_var = tk.StringVar()
        self.email_code_var = tk.StringVar()
        self.register_password_var = tk.StringVar()
        self.confirm_password_var = tk.StringVar()
        self.remember_var = tk.BooleanVar(value=True)
        self.show_password_var = tk.BooleanVar(value=False)

        existing_session = self.session_store.load_session()
        if existing_session:
            self._show_loading()
            self.root.after(80, lambda: self._validate_session(existing_session))
        else:
            self._show_auth("login")

    def _configure_styles(self) -> None:
        style = ttk.Style(self.root)
        try:
            style.theme_use("clam")
        except tk.TclError:
            pass
        style.configure(
            "Auth.TEntry",
            fieldbackground=self.SURFACE,
            foreground=self.TEXT,
            bordercolor=self.BORDER,
            lightcolor=self.BORDER,
            darkcolor=self.BORDER,
            padding=(11, 9),
            font=("Microsoft YaHei UI", 10),
        )
        style.map(
            "Auth.TEntry",
            bordercolor=[("focus", self.ACCENT)],
            lightcolor=[("focus", self.ACCENT)],
            darkcolor=[("focus", self.ACCENT)],
        )
        style.configure(
            "Primary.TButton",
            background=self.ACCENT,
            foreground="#f8fbff",
            borderwidth=0,
            padding=(14, 11),
            font=("Microsoft YaHei UI", 10, "bold"),
        )
        style.map(
            "Primary.TButton",
            background=[
                ("active", self.ACCENT_HOVER),
                ("pressed", self.ACCENT_HOVER),
                ("disabled", "#9db5e8"),
            ],
            foreground=[("disabled", "#edf2fc")],
        )
        style.configure(
            "Secondary.TButton",
            background=self.SURFACE,
            foreground=self.ACCENT,
            bordercolor=self.BORDER,
            lightcolor=self.BORDER,
            darkcolor=self.BORDER,
            padding=(10, 8),
            font=("Microsoft YaHei UI", 9),
        )
        style.map(
            "Secondary.TButton",
            background=[("active", "#eef3fb"), ("disabled", "#f0f2f5")],
            foreground=[("disabled", "#9099a6")],
        )

    def _set_geometry(self, width: int, height: int) -> None:
        self.root.update_idletasks()
        screen_width = self.root.winfo_screenwidth()
        screen_height = self.root.winfo_screenheight()
        x = max(0, (screen_width - width) // 2)
        y = max(0, (screen_height - height) // 2)
        self.root.geometry(f"{width}x{height}+{x}+{y}")

    def _clear_root(self) -> None:
        if self.code_timer_id:
            self.root.after_cancel(self.code_timer_id)
            self.code_timer_id = None
        self.root.unbind("<Return>")
        for child in self.root.winfo_children():
            child.destroy()

    def _show_loading(self) -> None:
        self._clear_root()
        self._set_geometry(420, 250)
        container = tk.Frame(self.root, bg=self.BACKGROUND, padx=36, pady=34)
        container.pack(fill=tk.BOTH, expand=True)
        tk.Label(
            container,
            text="AT",
            width=3,
            height=1,
            bg=self.ACCENT,
            fg="#f8fbff",
            font=("Segoe UI", 15, "bold"),
        ).pack(anchor="w")
        tk.Label(
            container,
            text="AutomaticTools",
            bg=self.BACKGROUND,
            fg=self.TEXT,
            font=("Segoe UI", 18, "bold"),
        ).pack(anchor="w", pady=(18, 4))
        tk.Label(
            container,
            text="正在检查登录状态...",
            bg=self.BACKGROUND,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 10),
        ).pack(anchor="w")

    def _show_auth(self, mode: str, message: str = "", success: bool = False) -> None:
        self.mode = mode
        self.busy = False
        self._clear_root()
        self.root.title("AutomaticTools - 登录" if mode == "login" else "AutomaticTools - 注册")
        self._set_geometry(460, 570 if mode == "login" else 650)

        page = tk.Frame(self.root, bg=self.BACKGROUND, padx=38, pady=28)
        page.pack(fill=tk.BOTH, expand=True)

        brand_row = tk.Frame(page, bg=self.BACKGROUND)
        brand_row.pack(fill=tk.X, pady=(0, 24))
        tk.Label(
            brand_row,
            text="AT",
            width=3,
            height=1,
            bg=self.ACCENT,
            fg="#f8fbff",
            font=("Segoe UI", 14, "bold"),
        ).pack(side=tk.LEFT)
        brand_text = tk.Frame(brand_row, bg=self.BACKGROUND)
        brand_text.pack(side=tk.LEFT, padx=(12, 0))
        tk.Label(
            brand_text,
            text="AutomaticTools",
            bg=self.BACKGROUND,
            fg=self.TEXT,
            font=("Segoe UI", 15, "bold"),
        ).pack(anchor="w")
        tk.Label(
            brand_text,
            text="实用工具平台",
            bg=self.BACKGROUND,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
        ).pack(anchor="w")

        tabs = tk.Frame(page, bg="#e8ecf2", padx=3, pady=3)
        tabs.pack(fill=tk.X, pady=(0, 24))
        self._tab_button(tabs, "登录", "login").pack(
            side=tk.LEFT, fill=tk.X, expand=True
        )
        self._tab_button(tabs, "注册", "register").pack(
            side=tk.LEFT, fill=tk.X, expand=True
        )

        heading = "登录账户" if mode == "login" else "创建账户"
        subheading = (
            "使用邮箱、用户名或手机号登录"
            if mode == "login"
            else "邮箱验证通过后即可使用工具"
        )
        tk.Label(
            page,
            text=heading,
            bg=self.BACKGROUND,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 17, "bold"),
        ).pack(anchor="w")
        self.default_subheading = subheading
        self.message_label = tk.Label(
            page,
            text=message or subheading,
            bg=self.BACKGROUND,
            fg=(self.SUCCESS if success else self.ERROR) if message else self.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
            justify=tk.LEFT,
            wraplength=350,
        )
        self.message_label.pack(fill=tk.X, pady=(5, 18))

        self.form_container = tk.Frame(page, bg=self.BACKGROUND)
        self.form_container.pack(fill=tk.X)
        self.password_entries: list[ttk.Entry] = []
        if mode == "login":
            self._build_login_form(self.form_container)
        else:
            self._build_register_form(self.form_container)

        tk.Checkbutton(
            page,
            text="记住登录状态",
            variable=self.remember_var,
            bg=self.BACKGROUND,
            activebackground=self.BACKGROUND,
            fg=self.TEXT,
            selectcolor=self.SURFACE,
            font=("Microsoft YaHei UI", 9),
            highlightthickness=0,
        ).pack(anchor="w", pady=(2, 14))

        self.submit_button = ttk.Button(
            page,
            text="登录" if mode == "login" else "注册并登录",
            style="Primary.TButton",
            command=self._submit,
        )
        self.submit_button.pack(fill=tk.X)

        self.root.bind("<Return>", lambda _event: self._submit())

    def _tab_button(self, parent: tk.Widget, text: str, mode: str) -> tk.Button:
        selected = self.mode == mode
        return tk.Button(
            parent,
            text=text,
            command=lambda: self._show_auth(mode),
            bg=self.SURFACE if selected else "#e8ecf2",
            activebackground=self.SURFACE if selected else "#dde3eb",
            fg=self.TEXT if selected else self.MUTED,
            activeforeground=self.TEXT,
            relief=tk.FLAT,
            borderwidth=0,
            pady=8,
            cursor="arrow" if selected else "hand2",
            font=("Microsoft YaHei UI", 9, "bold" if selected else "normal"),
        )

    def _field_label(self, parent: tk.Widget, text: str) -> None:
        tk.Label(
            parent,
            text=text,
            bg=self.BACKGROUND,
            fg=self.TEXT,
            anchor="w",
            font=("Microsoft YaHei UI", 9),
        ).pack(fill=tk.X, pady=(0, 6))

    def _entry(
        self,
        parent: tk.Widget,
        variable: tk.StringVar,
        password: bool = False,
    ) -> ttk.Entry:
        entry = ttk.Entry(
            parent,
            textvariable=variable,
            show="" if self.show_password_var.get() or not password else "•",
            style="Auth.TEntry",
        )
        entry.pack(fill=tk.X, pady=(0, 14))
        if password:
            self.password_entries.append(entry)
        return entry

    def _build_login_form(self, page: tk.Widget) -> None:
        self._field_label(page, "账号")
        account_entry = self._entry(page, self.account_var)
        self._field_label(page, "密码")
        self._entry(page, self.login_password_var, password=True)
        self._password_visibility_toggle(page)
        account_entry.focus_set()

    def _build_register_form(self, page: tk.Widget) -> None:
        self._field_label(page, "邮箱")
        email_entry = self._entry(page, self.email_var)
        self._field_label(page, "邮箱验证码")
        code_row = tk.Frame(page, bg=self.BACKGROUND)
        code_row.pack(fill=tk.X, pady=(0, 14))
        code_entry = ttk.Entry(
            code_row,
            textvariable=self.email_code_var,
            style="Auth.TEntry",
            width=18,
        )
        code_entry.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(0, 8))
        self.code_button = ttk.Button(
            code_row,
            text="发送验证码",
            style="Secondary.TButton",
            command=self._send_code,
        )
        self.code_button.pack(side=tk.RIGHT)

        password_row = tk.Frame(page, bg=self.BACKGROUND)
        password_row.pack(fill=tk.X)
        password_column = tk.Frame(password_row, bg=self.BACKGROUND)
        password_column.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(0, 6))
        confirm_column = tk.Frame(password_row, bg=self.BACKGROUND)
        confirm_column.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(6, 0))
        self._field_label(password_column, "密码")
        self._entry(password_column, self.register_password_var, password=True)
        self._field_label(confirm_column, "确认密码")
        self._entry(confirm_column, self.confirm_password_var, password=True)
        self._password_visibility_toggle(page)
        email_entry.focus_set()
        self._update_code_button()

    def _password_visibility_toggle(self, parent: tk.Widget) -> None:
        tk.Checkbutton(
            parent,
            text="显示密码",
            variable=self.show_password_var,
            command=self._toggle_password_visibility,
            bg=self.BACKGROUND,
            activebackground=self.BACKGROUND,
            fg=self.MUTED,
            selectcolor=self.SURFACE,
            font=("Microsoft YaHei UI", 9),
            highlightthickness=0,
        ).pack(anchor="w", pady=(0, 8))

    def _toggle_password_visibility(self) -> None:
        show = "" if self.show_password_var.get() else "•"
        for entry in self.password_entries:
            entry.configure(show=show)

    def _show_message(self, message: str, success: bool = False) -> None:
        self.message_label.configure(
            text=message,
            bg=self.BACKGROUND,
            fg=self.SUCCESS if success else self.ERROR,
        )

    def _send_code(self) -> None:
        if self.busy:
            return
        email = self.email_var.get().strip().lower()
        if not self._valid_email(email):
            self._show_message("请输入正确的邮箱地址。")
            return

        self.code_button.configure(state=tk.DISABLED, text="正在发送...")

        def success(response: dict) -> None:
            resend_after = response.get("resendAfter", 60)
            try:
                seconds = max(1, int(resend_after))
            except (TypeError, ValueError):
                seconds = 60
            self.code_available_at = time.monotonic() + seconds
            self._show_message("验证码已发送，请检查邮箱。", success=True)
            self._update_code_button()

        def failed(exc: Exception) -> None:
            self.code_button.configure(state=tk.NORMAL, text="发送验证码")
            self._show_message(self._error_message(exc))

        self._run_async(lambda: self.api.send_registration_code(email), success, failed)

    def _update_code_button(self) -> None:
        if self.mode != "register" or not hasattr(self, "code_button"):
            return
        remaining = int(self.code_available_at - time.monotonic() + 0.999)
        if remaining > 0:
            self.code_button.configure(state=tk.DISABLED, text=f"{remaining} 秒后重发")
            self.code_timer_id = self.root.after(250, self._update_code_button)
        else:
            self.code_button.configure(state=tk.NORMAL, text="发送验证码")
            self.code_timer_id = None

    def _submit(self) -> None:
        if self.busy:
            return
        if self.mode == "login":
            self._submit_login()
        else:
            self._submit_register()

    def _submit_login(self) -> None:
        account = self.account_var.get().strip()
        password = self.login_password_var.get()
        if not account:
            self._show_message("请输入邮箱、用户名或手机号。")
            return
        if not password:
            self._show_message("请输入密码。")
            return

        self._set_busy(True, "正在登录...")
        device_id = self.session_store.get_device_id()
        self._run_async(
            lambda: self.api.login(account, password, device_id),
            self._complete_authentication,
            self._authentication_failed,
        )

    def _submit_register(self) -> None:
        email = self.email_var.get().strip().lower()
        code = self.email_code_var.get().strip()
        password = self.register_password_var.get()
        confirm_password = self.confirm_password_var.get()

        if not self._valid_email(email):
            self._show_message("请输入正确的邮箱地址。")
            return
        if not re.fullmatch(r"[0-9]{6}", code):
            self._show_message("邮箱验证码必须是 6 位数字。")
            return
        if len(password) < 6:
            self._show_message("密码至少需要 6 个字符。")
            return
        if len(password.encode("utf-8")) > 72:
            self._show_message("密码不能超过 72 个字节。")
            return
        if password != confirm_password:
            self._show_message("两次输入的密码不一致。")
            return

        self._set_busy(True, "正在注册...")
        device_id = self.session_store.get_device_id()
        self._run_async(
            lambda: self.api.register(email, code, password, "", device_id),
            self._complete_authentication,
            self._authentication_failed,
        )

    def _set_busy(self, busy: bool, text: str = "") -> None:
        self.busy = busy
        self.submit_button.configure(
            state=tk.DISABLED if busy else tk.NORMAL,
            text=text or ("登录" if self.mode == "login" else "注册并登录"),
        )

    def _authentication_failed(self, exc: Exception) -> None:
        self._set_busy(False)
        self._show_message(self._error_message(exc))

    def _complete_authentication(self, response: dict) -> None:
        token = response.get("token")
        user = response.get("user")
        if not isinstance(token, str) or not token or not isinstance(user, dict):
            self._authentication_failed(ApiError("服务器返回的登录信息不完整。"))
            return
        if self.remember_var.get():
            try:
                self.session_store.save_session(token, user)
            except OSError:
                self._authentication_failed(ApiError("无法安全保存登录状态，请重试。"))
                return
        else:
            try:
                self.session_store.clear_session()
            except OSError:
                self._authentication_failed(ApiError("无法清除旧的登录状态，请重试。"))
                return
        self.session = {"token": token, "user": user}
        self.root.destroy()

    def _validate_session(self, session: dict) -> None:
        token = session["token"]

        def success(response: dict) -> None:
            user = response.get("user")
            if isinstance(user, dict):
                session["user"] = user
                try:
                    self.session_store.save_session(token, user)
                except OSError:
                    pass
                self.session = session
                self.root.destroy()
                return
            self.session_store.clear_session()
            self._show_auth("login", "登录状态无效，请重新登录。")

        def failed(exc: Exception) -> None:
            if isinstance(exc, ApiError) and exc.status == 401:
                try:
                    self.session_store.clear_session()
                except OSError:
                    pass
                self._show_auth("login", "登录状态已过期，请重新登录。")
            else:
                self._show_auth("login", self._error_message(exc))

        self._run_async(lambda: self.api.current_user(token), success, failed)

    def _run_async(
        self,
        task,
        on_success,
        on_error,
    ) -> None:
        def worker() -> None:
            try:
                result = task()
            except Exception as exc:
                self.root.after(0, lambda error=exc: on_error(error))
                return
            self.root.after(0, lambda: on_success(result))

        threading.Thread(target=worker, daemon=True).start()

    @staticmethod
    def _valid_email(email: str) -> bool:
        return bool(re.fullmatch(r"[^\s@]+@[^\s@]+\.[^\s@]+", email)) and len(email) <= 254

    @staticmethod
    def _error_message(exc: Exception) -> str:
        if isinstance(exc, (ApiError, NetworkError)):
            message = str(exc)
            if isinstance(exc, ApiError) and exc.request_id:
                return f"{message}\n请求编号：{exc.request_id}"
            return message
        return "操作失败，请稍后重试。"

    def close(self) -> None:
        self.session = None
        self.root.destroy()

    def run(self) -> dict | None:
        self.root.mainloop()
        return self.session


class LicenseCodeDialog:
    def __init__(self, app, tool_name: str = "") -> None:
        self.app = app
        self.api = app.api
        self.token = app.token
        self.submitting = False
        self.closed = False

        self.window = tk.Toplevel(app.root)
        self.window.title("兑换授权码")
        self.window.geometry("460x350")
        self.window.resizable(False, False)
        self.window.configure(bg=app.BACKGROUND)
        self.window.transient(app.root)
        self.window.grab_set()
        self.window.protocol("WM_DELETE_WINDOW", self.close)
        self._center()

        shell = tk.Frame(self.window, bg=app.BACKGROUND, padx=30, pady=26)
        shell.pack(fill=tk.BOTH, expand=True)
        tk.Label(
            shell,
            text="兑换授权码",
            bg=app.BACKGROUND,
            fg=app.TEXT,
            font=("Microsoft YaHei UI", 16, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        subtitle = f"输入购买的授权码，兑换 {tool_name}" if tool_name else "输入购买的授权码，系统会自动识别对应工具"
        tk.Label(
            shell,
            text=subtitle,
            bg=app.BACKGROUND,
            fg=app.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
        ).pack(fill=tk.X, pady=(7, 22))

        tk.Label(
            shell,
            text="授权码",
            bg=app.BACKGROUND,
            fg=app.TEXT,
            font=("Microsoft YaHei UI", 9, "bold"),
            anchor="w",
        ).pack(fill=tk.X, pady=(0, 7))
        self.code_var = tk.StringVar()
        self.code_entry = ttk.Entry(
            shell,
            textvariable=self.code_var,
            style="App.TEntry",
            font=("Consolas", 11),
        )
        self.code_entry.pack(fill=tk.X)
        self.code_entry.bind("<Return>", lambda _event: self.submit())
        self.code_entry.focus_set()

        self.status_var = tk.StringVar(value="示例：AT-XXXX-XXXX-XXXX-XXXX")
        self.status_label = tk.Label(
            shell,
            textvariable=self.status_var,
            bg=app.BACKGROUND,
            fg=app.MUTED,
            font=("Microsoft YaHei UI", 8),
            anchor="w",
            justify=tk.LEFT,
            wraplength=390,
        )
        self.status_label.pack(fill=tk.X, pady=(8, 18))

        actions = tk.Frame(shell, bg=app.BACKGROUND)
        actions.pack(fill=tk.X, side=tk.BOTTOM)
        ttk.Button(
            actions,
            text="取消",
            style="AppSecondary.TButton",
            command=self.close,
        ).pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(0, 6))
        self.submit_button = ttk.Button(
            actions,
            text="确认兑换",
            style="AppPrimary.TButton",
            command=self.submit,
        )
        self.submit_button.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(6, 0))

    def _center(self) -> None:
        self.window.update_idletasks()
        width = 460
        height = 350
        x = self.app.root.winfo_rootx() + max(0, (self.app.root.winfo_width() - width) // 2)
        y = self.app.root.winfo_rooty() + max(0, (self.app.root.winfo_height() - height) // 2)
        self.window.geometry(f"{width}x{height}+{x}+{y}")

    def submit(self) -> None:
        if self.submitting or self.closed:
            return
        code = self.code_var.get().strip().upper()
        if not code:
            self.status_var.set("请输入授权码。")
            self.status_label.configure(fg=self.app.DANGER)
            self.code_entry.focus_set()
            return

        self.submitting = True
        self.code_entry.configure(state=tk.DISABLED)
        self.submit_button.configure(state=tk.DISABLED, text="兑换中...")
        self.status_var.set("正在验证授权码...")
        self.status_label.configure(fg=self.app.MUTED)

        def success(response: dict) -> None:
            if self.closed:
                return
            tool = response.get("tool")
            purchase = response.get("purchase")
            if not isinstance(tool, dict) or not isinstance(purchase, dict):
                failed(ApiError("服务器返回的兑换结果不完整，请刷新后确认。"))
                return
            name = str(tool.get("name") or tool.get("code") or "工具")
            self.status_var.set(f"兑换成功：{name} 已加入工具中心。")
            self.status_label.configure(fg=self.app.SUCCESS)
            self.submit_button.configure(text="兑换成功", state=tk.DISABLED)
            self.app.license_code_redeemed(purchase)
            self.window.after(1100, self.close)

        def failed(exc: Exception) -> None:
            if self.closed:
                return
            self.submitting = False
            self.code_entry.configure(state=tk.NORMAL)
            self.submit_button.configure(state=tk.NORMAL, text="确认兑换")
            self.status_var.set(self.app._error_message(exc))
            self.status_label.configure(fg=self.app.DANGER)
            self.code_entry.focus_set()

        self.app._run_async(
            lambda: self.api.redeem_license_code(self.token, code),
            success,
            failed,
        )

    def close(self) -> None:
        if self.closed:
            return
        self.closed = True
        try:
            self.window.grab_release()
        except tk.TclError:
            pass
        self.window.destroy()


class AlipayPaymentDialog:
    def __init__(
        self,
        app,
        tool: dict,
        payment: dict,
        on_paid,
    ) -> None:
        self.app = app
        self.api = app.api
        self.token = app.token
        self.on_paid = on_paid
        self.order = dict(payment.get("order") or {})
        self.order_no = str(self.order.get("orderNo") or "")
        self.expires_at = int(payment.get("expiresAt") or 0)
        self.polling = False
        self.closed = False

        self.window = tk.Toplevel(app.root)
        self.window.title("支付宝扫码支付")
        self.window.geometry("430x620")
        self.window.resizable(False, False)
        self.window.configure(bg=app.BACKGROUND)
        self.window.transient(app.root)
        self.window.grab_set()
        self.window.protocol("WM_DELETE_WINDOW", self.close)
        self._center()

        shell = tk.Frame(self.window, bg=app.BACKGROUND, padx=28, pady=24)
        shell.pack(fill=tk.BOTH, expand=True)
        tk.Label(
            shell,
            text="支付宝扫码支付",
            bg=app.BACKGROUND,
            fg=app.TEXT,
            font=("Microsoft YaHei UI", 16, "bold"),
        ).pack()
        tk.Label(
            shell,
            text=f"购买 {tool.get('name') or self.order.get('toolCode') or '工具'}",
            bg=app.BACKGROUND,
            fg=app.MUTED,
            font=("Microsoft YaHei UI", 9),
        ).pack(pady=(6, 16))

        qr_card = tk.Frame(
            shell,
            bg="#ffffff",
            highlightthickness=1,
            highlightbackground=app.BORDER,
            padx=14,
            pady=14,
        )
        qr_card.pack()
        self.qr_photo = self._make_qr(str(payment.get("qrCode") or ""))
        tk.Label(qr_card, image=self.qr_photo, bg="#ffffff").pack()

        amount_text = app._format_price(tool).replace(" / 永久", "")
        tk.Label(
            shell,
            text=amount_text,
            bg=app.BACKGROUND,
            fg=app.TEXT,
            font=("Microsoft YaHei UI", 18, "bold"),
        ).pack(pady=(16, 4))
        tk.Label(
            shell,
            text=f"订单号：{self.order_no}",
            bg=app.BACKGROUND,
            fg=app.MUTED,
            font=("Segoe UI", 8),
        ).pack()

        self.status_var = tk.StringVar(value="请使用支付宝扫描二维码")
        self.status_label = tk.Label(
            shell,
            textvariable=self.status_var,
            bg=app.WARNING_BG,
            fg=app.WARNING,
            font=("Microsoft YaHei UI", 9, "bold"),
            padx=12,
            pady=9,
        )
        self.status_label.pack(fill=tk.X, pady=(16, 12))
        ttk.Button(
            shell,
            text="暂不支付",
            style="AppSecondary.TButton",
            command=self.close,
        ).pack(fill=tk.X)
        self.window.after(300, self._poll)

    def _center(self) -> None:
        self.window.update_idletasks()
        width = 430
        height = 620
        x = self.app.root.winfo_rootx() + max(0, (self.app.root.winfo_width() - width) // 2)
        y = self.app.root.winfo_rooty() + max(0, (self.app.root.winfo_height() - height) // 2)
        self.window.geometry(f"{width}x{height}+{x}+{y}")

    @staticmethod
    def _make_qr(content: str):
        qr = qrcode.QRCode(version=None, box_size=8, border=2)
        qr.add_data(content)
        qr.make(fit=True)
        image = qr.make_image(fill_color="#111827", back_color="#ffffff")
        image = image.get_image().convert("RGB")
        image = image.resize((220, 220), Image.Resampling.NEAREST)
        return ImageTk.PhotoImage(image)

    def _poll(self) -> None:
        if self.closed or self.polling:
            return
        if self.expires_at and int(time.time()) >= self.expires_at:
            self.status_var.set("二维码已过期，请关闭后重新购买")
            self.status_label.configure(bg=self.app.DANGER_BG, fg=self.app.DANGER)
            return
        self.polling = True

        def worker() -> None:
            try:
                response = self.api.payment_order_status(self.token, self.order_no)
            except Exception as exc:
                self._schedule(lambda error=exc: self._poll_failed(error))
                return
            self._schedule(lambda: self._status_loaded(response))

        threading.Thread(target=worker, daemon=True).start()

    def _schedule(self, callback) -> None:
        if self.closed:
            return
        try:
            self.window.after(0, callback)
        except (RuntimeError, tk.TclError):
            pass

    def _status_loaded(self, response: dict) -> None:
        self.polling = False
        if self.closed:
            return
        order = response.get("order")
        status = str(order.get("status") if isinstance(order, dict) else "pending")
        if status == "paid":
            self.status_var.set("支付成功，工具已加入已购列表")
            self.status_label.configure(bg=self.app.SUCCESS_BG, fg=self.app.SUCCESS)
            self.on_paid()
            self.window.after(1200, self.close)
            return
        if status in {"cancelled", "failed", "refunded"}:
            text = {
                "cancelled": "订单已取消，请重新购买",
                "failed": "订单创建失败，请重新购买",
                "refunded": "订单已退款",
            }[status]
            self.status_var.set(text)
            self.status_label.configure(bg=self.app.DANGER_BG, fg=self.app.DANGER)
            return
        remaining = max(0, self.expires_at - int(time.time())) if self.expires_at else 0
        self.status_var.set(f"等待支付 · 二维码约 {max(1, remaining // 60)} 分钟后过期")
        self.window.after(2000, self._poll)

    def _poll_failed(self, exc: Exception) -> None:
        self.polling = False
        if self.closed:
            return
        self.status_var.set("正在等待支付结果，网络恢复后会自动重试")
        self.status_label.configure(bg=self.app.WARNING_BG, fg=self.app.WARNING)
        self.window.after(3000, self._poll)

    def close(self) -> None:
        if self.closed:
            return
        self.closed = True
        try:
            self.window.grab_release()
        except tk.TclError:
            pass
        self.window.destroy()


class AutomaticToolsApp:
    BACKGROUND = "#f3f5f7"
    SURFACE = "#fbfcfd"
    SURFACE_ALT = "#eef2f5"
    SIDEBAR = "#20272d"
    SIDEBAR_HOVER = "#2b343c"
    TEXT = "#17212b"
    MUTED = "#65717d"
    BORDER = "#d8dfe5"
    ACCENT = "#2864dc"
    ACCENT_HOVER = "#1f52b7"
    SUCCESS = "#14816f"
    SUCCESS_BG = "#e8f5f1"
    WARNING = "#9a6700"
    WARNING_BG = "#fff4d6"
    DANGER = "#b42318"
    DANGER_BG = "#fcebea"

    PAGE_META = {
        "home": ("首页", "查看账户和工具使用状态"),
        "tools": ("工具中心", "浏览并使用授权码开通 AutomaticTools 工具"),
        "clicker": ("自动点击", "锁定坐标并执行稳定的重复点击"),
        "account": ("账户信息", "查看个人资料和工具开通方式"),
    }

    def __init__(
        self,
        session_store: SessionStore,
        session: dict,
        api: ApiClient,
    ) -> None:
        self.session_store = session_store
        self.session = session
        self.api = api
        self.token = str(session.get("token", ""))
        self.user = dict(session.get("user", {}))
        self.tools: list[dict] = []
        self.purchases: list[dict] = []
        self.orders: list[dict] = []
        self.purchased_codes: set[str] = set()
        self.payment_submitting_codes: set[str] = set()
        self.remote_loading = False
        self.logout_requested = False
        self.active_view = "home"

        self.locked_position: tuple[int, int] | None = None
        self.locked_window_target: tuple[int, int, int] | None = None
        self.click_count = 0
        self.running = False
        self.always_on_top = False
        self.worker_thread: threading.Thread | None = None
        self.stop_event = threading.Event()

        self.root = tk.Tk()
        apply_window_icon(self.root)
        self.root.title("AutomaticTools")
        self.root.geometry("1040x680")
        self.root.minsize(920, 620)
        self.root.configure(bg=self.BACKGROUND)
        self.root.attributes("-topmost", False)

        self.current_position_var = tk.StringVar(value="当前鼠标：-，-")
        self.locked_position_var = tk.StringVar(value="锁定位置：未锁定")
        self.click_count_var = tk.StringVar(value="已点击 0 次")
        self.status_var = tk.StringVar(value="已停止")
        self.topmost_var = tk.StringVar(value="窗口未置顶")
        self.interval_var = tk.StringVar(value="1000")
        self.no_cursor_move_var = tk.BooleanVar(value=False)
        self.page_title_var = tk.StringVar()
        self.page_subtitle_var = tk.StringVar()
        self.sync_status_var = tk.StringVar(value="正在同步账户数据...")
        self.sidebar_name_var = tk.StringVar()
        self.sidebar_email_var = tk.StringVar()
        self.home_welcome_var = tk.StringVar()
        self.home_account_var = tk.StringVar(value="正常")
        self.home_purchase_var = tk.StringVar(value="0 项")
        self.home_order_var = tk.StringVar(value="授权码")
        self.home_tool_status_var = tk.StringVar(value="正在检查开通状态")

        self._configure_styles()
        self._build_ui()
        self._update_user_views()
        self.switch_view("home")
        self.refresh_mouse_position()
        self.poll_hotkeys()
        self.refresh_remote_data()
        self.root.protocol("WM_DELETE_WINDOW", self.close)

    def _configure_styles(self) -> None:
        style = ttk.Style(self.root)
        try:
            style.theme_use("clam")
        except tk.TclError:
            pass
        style.configure(
            "App.TEntry",
            fieldbackground=self.SURFACE,
            foreground=self.TEXT,
            bordercolor=self.BORDER,
            lightcolor=self.BORDER,
            darkcolor=self.BORDER,
            padding=(10, 8),
            font=("Microsoft YaHei UI", 10),
        )
        style.map(
            "App.TEntry",
            bordercolor=[("focus", self.ACCENT)],
            lightcolor=[("focus", self.ACCENT)],
            darkcolor=[("focus", self.ACCENT)],
        )
        style.configure(
            "AppPrimary.TButton",
            background=self.ACCENT,
            foreground="#f8fbff",
            borderwidth=0,
            padding=(14, 10),
            font=("Microsoft YaHei UI", 9, "bold"),
        )
        style.map(
            "AppPrimary.TButton",
            background=[
                ("active", self.ACCENT_HOVER),
                ("pressed", self.ACCENT_HOVER),
                ("disabled", "#a7b9d8"),
            ],
        )
        style.configure(
            "AppSecondary.TButton",
            background=self.SURFACE,
            foreground=self.TEXT,
            bordercolor=self.BORDER,
            lightcolor=self.BORDER,
            darkcolor=self.BORDER,
            padding=(12, 9),
            font=("Microsoft YaHei UI", 9),
        )
        style.map(
            "AppSecondary.TButton",
            background=[("active", self.SURFACE_ALT), ("disabled", "#eceff2")],
            foreground=[("disabled", "#9aa3ac")],
        )

    def _build_ui(self) -> None:
        shell = tk.Frame(self.root, bg=self.BACKGROUND)
        shell.pack(fill=tk.BOTH, expand=True)

        self._build_sidebar(shell)
        main = tk.Frame(shell, bg=self.BACKGROUND)
        main.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        self._build_header(main)

        self.content_host = tk.Frame(main, bg=self.BACKGROUND, padx=28, pady=22)
        self.content_host.pack(fill=tk.BOTH, expand=True)
        self.views: dict[str, tk.Frame] = {}
        self._build_home_view()
        self._build_tools_view()
        self._build_clicker_view()
        self._build_account_view()

    def _build_sidebar(self, parent: tk.Widget) -> None:
        sidebar = tk.Frame(parent, bg=self.SIDEBAR, width=220)
        sidebar.pack(side=tk.LEFT, fill=tk.Y)
        sidebar.pack_propagate(False)

        brand = tk.Frame(sidebar, bg=self.SIDEBAR, padx=20, pady=22)
        brand.pack(fill=tk.X)
        tk.Label(
            brand,
            text="AT",
            width=3,
            height=1,
            bg=self.ACCENT,
            fg="#f8fbff",
            font=("Segoe UI", 13, "bold"),
        ).pack(side=tk.LEFT)
        brand_text = tk.Frame(brand, bg=self.SIDEBAR)
        brand_text.pack(side=tk.LEFT, padx=(11, 0))
        tk.Label(
            brand_text,
            text="AutomaticTools",
            bg=self.SIDEBAR,
            fg="#f3f6f8",
            font=("Segoe UI", 12, "bold"),
        ).pack(anchor="w")
        tk.Label(
            brand_text,
            text="工具平台",
            bg=self.SIDEBAR,
            fg="#aeb8c1",
            font=("Microsoft YaHei UI", 8),
        ).pack(anchor="w")

        tk.Label(
            sidebar,
            text="导航",
            bg=self.SIDEBAR,
            fg="#87939e",
            font=("Microsoft YaHei UI", 8),
            anchor="w",
            padx=24,
        ).pack(fill=tk.X, pady=(12, 7))

        self.nav_buttons: dict[str, tk.Button] = {}
        for key, label in (("home", "首页"), ("tools", "工具中心")):
            button = tk.Button(
                sidebar,
                text=label,
                command=lambda page=key: self.switch_view(page),
                bg=self.SIDEBAR,
                activebackground=self.SIDEBAR_HOVER,
                fg="#cbd3da",
                activeforeground="#f7f9fa",
                relief=tk.FLAT,
                borderwidth=0,
                anchor="w",
                padx=24,
                pady=11,
                cursor="hand2",
                font=("Microsoft YaHei UI", 10),
            )
            button.pack(fill=tk.X, padx=10, pady=2)
            self.nav_buttons[key] = button

        self.purchased_nav_host = tk.Frame(sidebar, bg=self.SIDEBAR)
        self.purchased_nav_host.pack(fill=tk.X)

        account_button = tk.Button(
            sidebar,
            text="账户信息",
            command=lambda: self.switch_view("account"),
            bg=self.SIDEBAR,
            activebackground=self.SIDEBAR_HOVER,
            fg="#cbd3da",
            activeforeground="#f7f9fa",
            relief=tk.FLAT,
            borderwidth=0,
            anchor="w",
            padx=24,
            pady=11,
            cursor="hand2",
            font=("Microsoft YaHei UI", 10),
        )
        account_button.pack(fill=tk.X, padx=10, pady=2)
        self.nav_buttons["account"] = account_button

        account_box = tk.Frame(sidebar, bg="#181e23", padx=18, pady=16)
        account_box.pack(side=tk.BOTTOM, fill=tk.X)
        user_row = tk.Frame(account_box, bg="#181e23")
        user_row.pack(fill=tk.X)
        self.sidebar_avatar = tk.Label(
            user_row,
            text="U",
            width=3,
            height=1,
            bg=self.SUCCESS,
            fg="#f5fffc",
            font=("Segoe UI", 11, "bold"),
        )
        self.sidebar_avatar.pack(side=tk.LEFT)
        identity = tk.Frame(user_row, bg="#181e23")
        identity.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(10, 0))
        tk.Label(
            identity,
            textvariable=self.sidebar_name_var,
            bg="#181e23",
            fg="#f0f4f6",
            anchor="w",
            font=("Microsoft YaHei UI", 9, "bold"),
        ).pack(fill=tk.X)
        tk.Label(
            identity,
            textvariable=self.sidebar_email_var,
            bg="#181e23",
            fg="#94a1ac",
            anchor="w",
            font=("Microsoft YaHei UI", 8),
        ).pack(fill=tk.X)
        tk.Button(
            account_box,
            text="退出登录",
            command=self.logout,
            bg="#181e23",
            activebackground=self.SIDEBAR_HOVER,
            fg="#b9c3cb",
            activeforeground="#f7f9fa",
            relief=tk.FLAT,
            borderwidth=0,
            anchor="w",
            pady=8,
            cursor="hand2",
            font=("Microsoft YaHei UI", 9),
        ).pack(fill=tk.X, pady=(10, 0))

    def _render_purchased_navigation(self) -> None:
        if not hasattr(self, "purchased_nav_host"):
            return
        for child in self.purchased_nav_host.winfo_children():
            child.destroy()
        for key in list(self.nav_buttons):
            if key not in {"home", "tools", "account"}:
                self.nav_buttons.pop(key, None)
        if not self.purchased_codes:
            return

        tk.Label(
            self.purchased_nav_host,
            text="已开通工具",
            bg=self.SIDEBAR,
            fg="#7f8b96",
            anchor="w",
            padx=35,
            font=("Microsoft YaHei UI", 8),
        ).pack(fill=tk.X, pady=(6, 3))
        tools_by_code = {
            str(tool.get("code") or ""): tool
            for tool in self.tools
            if tool.get("code")
        }
        for code in sorted(self.purchased_codes):
            tool = tools_by_code.get(code, {})
            page_key = "clicker" if code == "auto_click" else f"tool:{code}"
            button = tk.Button(
                self.purchased_nav_host,
                text=f"•  {tool.get('name') or code}",
                command=lambda value=code: self._open_tool(value),
                bg=self.SIDEBAR,
                activebackground=self.SIDEBAR_HOVER,
                fg="#b9c4ce",
                activeforeground="#f7f9fa",
                relief=tk.FLAT,
                borderwidth=0,
                anchor="w",
                padx=35,
                pady=8,
                cursor="hand2",
                font=("Microsoft YaHei UI", 9),
            )
            button.pack(fill=tk.X, padx=10, pady=1)
            self.nav_buttons[page_key] = button

    def _build_header(self, parent: tk.Widget) -> None:
        header = tk.Frame(parent, bg=self.SURFACE, height=84, padx=28, pady=16)
        header.pack(fill=tk.X)
        header.pack_propagate(False)
        titles = tk.Frame(header, bg=self.SURFACE)
        titles.pack(side=tk.LEFT, fill=tk.X, expand=True)
        tk.Label(
            titles,
            textvariable=self.page_title_var,
            bg=self.SURFACE,
            fg=self.TEXT,
            anchor="w",
            font=("Microsoft YaHei UI", 15, "bold"),
        ).pack(fill=tk.X)
        tk.Label(
            titles,
            textvariable=self.page_subtitle_var,
            bg=self.SURFACE,
            fg=self.MUTED,
            anchor="w",
            font=("Microsoft YaHei UI", 9),
        ).pack(fill=tk.X, pady=(4, 0))

        actions = tk.Frame(header, bg=self.SURFACE)
        actions.pack(side=tk.RIGHT)
        self.sync_label = tk.Label(
            actions,
            textvariable=self.sync_status_var,
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
        )
        self.sync_label.pack(side=tk.LEFT, padx=(0, 12))
        self.refresh_button = ttk.Button(
            actions,
            text="刷新数据",
            style="AppSecondary.TButton",
            command=self.refresh_remote_data,
        )
        self.refresh_button.pack(side=tk.LEFT)

    def _new_view(self, key: str) -> tk.Frame:
        view = tk.Frame(self.content_host, bg=self.BACKGROUND)
        self.views[key] = view
        return view

    def _panel(self, parent: tk.Widget, **options) -> tk.Frame:
        return tk.Frame(
            parent,
            bg=self.SURFACE,
            highlightthickness=1,
            highlightbackground=self.BORDER,
            **options,
        )

    def _build_home_view(self) -> None:
        view = self._new_view("home")
        welcome = self._panel(view, padx=24, pady=20)
        welcome.pack(fill=tk.X)
        tk.Label(
            welcome,
            textvariable=self.home_welcome_var,
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 16, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            welcome,
            text="这里集中展示你的账户和已开通工具。",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
        ).pack(fill=tk.X, pady=(6, 0))

        metrics = tk.Frame(view, bg=self.BACKGROUND)
        metrics.pack(fill=tk.X, pady=16)
        self._metric(metrics, "账户状态", self.home_account_var, self.SUCCESS, 0)
        self._metric(metrics, "已开通工具", self.home_purchase_var, self.ACCENT, 1)
        self._metric(metrics, "开通方式", self.home_order_var, self.WARNING, 2)

        tk.Label(
            view,
            text="快速开始",
            bg=self.BACKGROUND,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 11, "bold"),
            anchor="w",
        ).pack(fill=tk.X, pady=(4, 10))

        quick = self._panel(view, padx=20, pady=17)
        quick.pack(fill=tk.X)
        mark = tk.Label(
            quick,
            text="点",
            width=3,
            height=1,
            bg="#e7eefc",
            fg=self.ACCENT,
            font=("Microsoft YaHei UI", 12, "bold"),
        )
        mark.pack(side=tk.LEFT)
        tool_text = tk.Frame(quick, bg=self.SURFACE)
        tool_text.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(14, 0))
        tk.Label(
            tool_text,
            text="自动点击",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 11, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            tool_text,
            text="锁定屏幕或窗口坐标，按设定间隔持续点击。",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
        ).pack(fill=tk.X, pady=(4, 0))
        action = tk.Frame(quick, bg=self.SURFACE)
        action.pack(side=tk.RIGHT, padx=(18, 0))
        self.home_tool_status_label = tk.Label(
            action,
            textvariable=self.home_tool_status_var,
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
        )
        self.home_tool_status_label.pack(anchor="e", pady=(0, 6))
        self.home_clicker_button = ttk.Button(
            action,
            text="打开功能",
            style="AppPrimary.TButton",
            command=lambda: self.switch_view("clicker"),
        )
        self.home_clicker_button.pack(anchor="e")

    def _metric(
        self,
        parent: tk.Widget,
        title: str,
        variable: tk.StringVar,
        color: str,
        column: int,
    ) -> None:
        parent.columnconfigure(column, weight=1, uniform="metrics")
        box = self._panel(parent, padx=18, pady=15)
        box.grid(
            row=0,
            column=column,
            sticky="nsew",
            padx=(0 if column == 0 else 6, 0 if column == 2 else 6),
        )
        tk.Label(
            box,
            text=title,
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            box,
            textvariable=variable,
            bg=self.SURFACE,
            fg=color,
            font=("Microsoft YaHei UI", 13, "bold"),
            anchor="w",
        ).pack(fill=tk.X, pady=(6, 0))

    def _build_tools_view(self) -> None:
        view = self._new_view("tools")
        intro = self._panel(view, padx=20, pady=16)
        intro.pack(fill=tk.X, pady=(0, 14))
        ttk.Button(
            intro,
            text="兑换授权码",
            style="AppPrimary.TButton",
            command=self.start_license_redeem,
        ).pack(side=tk.RIGHT, padx=(18, 0))
        tk.Label(
            intro,
            text="全部工具",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 11, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            intro,
            text="已开通工具可直接使用；输入购买的授权码即可开通新工具。",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
        ).pack(fill=tk.X, pady=(5, 0))
        self.tools_container = tk.Frame(view, bg=self.BACKGROUND)
        self.tools_container.pack(fill=tk.BOTH, expand=True)
        self._render_tools()

    def _build_clicker_view(self) -> None:
        view = self._new_view("clicker")
        self.clicker_access_notice = tk.Label(
            view,
            text="正在检查开通状态...",
            bg=self.WARNING_BG,
            fg=self.WARNING,
            padx=14,
            pady=9,
            anchor="w",
            font=("Microsoft YaHei UI", 9),
        )
        self.clicker_access_notice.pack(fill=tk.X, pady=(0, 12))

        coordinates = tk.Frame(view, bg=self.BACKGROUND)
        coordinates.pack(fill=tk.X)
        current_box = self._panel(coordinates, padx=18, pady=15)
        current_box.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(0, 7))
        locked_box = self._panel(coordinates, padx=18, pady=15)
        locked_box.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(7, 0))
        for parent, title, variable in (
            (current_box, "鼠标位置", self.current_position_var),
            (locked_box, "点击目标", self.locked_position_var),
        ):
            tk.Label(
                parent,
                text=title,
                bg=self.SURFACE,
                fg=self.MUTED,
                font=("Microsoft YaHei UI", 8),
                anchor="w",
            ).pack(fill=tk.X)
            tk.Label(
                parent,
                textvariable=variable,
                bg=self.SURFACE,
                fg=self.TEXT,
                font=("Microsoft YaHei UI", 10, "bold"),
                anchor="w",
                wraplength=310,
            ).pack(fill=tk.X, pady=(6, 0))

        settings = self._panel(view, padx=20, pady=16)
        settings.pack(fill=tk.X, pady=14)
        settings_top = tk.Frame(settings, bg=self.SURFACE)
        settings_top.pack(fill=tk.X)
        tk.Label(
            settings_top,
            text="点击设置",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 10, "bold"),
        ).pack(side=tk.LEFT)
        tk.Label(
            settings_top,
            text="F9 锁定  F8 开始/停止  F7 置顶  F6 后台模式",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
        ).pack(side=tk.RIGHT)

        option_row = tk.Frame(settings, bg=self.SURFACE)
        option_row.pack(fill=tk.X, pady=(14, 12))
        self.no_cursor_check = tk.Checkbutton(
            option_row,
            text="后台点击，不移动光标（F6）",
            variable=self.no_cursor_move_var,
            bg=self.SURFACE,
            activebackground=self.SURFACE,
            fg=self.TEXT,
            selectcolor=self.SURFACE,
            highlightthickness=0,
            font=("Microsoft YaHei UI", 9),
        )
        self.no_cursor_check.pack(side=tk.LEFT)
        self.topmost_button = ttk.Button(
            option_row,
            text="窗口置顶（F7）",
            style="AppSecondary.TButton",
            command=self.toggle_topmost,
        )
        self.topmost_button.pack(side=tk.RIGHT)

        interval_row = tk.Frame(settings, bg=self.SURFACE)
        interval_row.pack(fill=tk.X)
        tk.Label(
            interval_row,
            text="点击间隔",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 9),
        ).pack(side=tk.LEFT)
        interval_entry = ttk.Entry(
            interval_row,
            textvariable=self.interval_var,
            width=9,
            style="App.TEntry",
        )
        interval_entry.pack(side=tk.LEFT, padx=(10, 5))
        tk.Label(
            interval_row,
            text="毫秒",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
        ).pack(side=tk.LEFT)

        quick_row = tk.Frame(settings, bg=self.SURFACE)
        quick_row.pack(fill=tk.X, pady=(12, 0))
        self.quick_buttons: dict[int, tk.Button] = {}
        for index, (milliseconds, text) in enumerate(
            ((100, "100 ms\nCtrl+1"), (500, "500 ms\nCtrl+2"), (1000, "1 秒\nCtrl+3"), (2000, "2 秒\nCtrl+4"))
        ):
            button = tk.Button(
                quick_row,
                text=text,
                command=lambda value=milliseconds: self.set_interval(value),
                bg=self.SURFACE_ALT,
                activebackground="#e2e8ed",
                fg=self.TEXT,
                activeforeground=self.TEXT,
                relief=tk.FLAT,
                borderwidth=0,
                pady=8,
                cursor="hand2",
                font=("Microsoft YaHei UI", 9),
            )
            button.pack(
                side=tk.LEFT,
                fill=tk.X,
                expand=True,
                padx=(0 if index == 0 else 5, 0 if index == 3 else 5),
            )
            self.quick_buttons[milliseconds] = button
        self._update_interval_buttons()

        action_row = tk.Frame(view, bg=self.BACKGROUND)
        action_row.pack(fill=tk.X)
        self.lock_button = tk.Button(
            action_row,
            text="锁定当前坐标  F9",
            command=self.lock_position,
            bg=self.SURFACE,
            activebackground=self.SURFACE_ALT,
            fg=self.TEXT,
            activeforeground=self.TEXT,
            disabledforeground="#9ba5ae",
            relief=tk.FLAT,
            highlightthickness=1,
            highlightbackground=self.BORDER,
            borderwidth=0,
            pady=14,
            cursor="hand2",
            font=("Microsoft YaHei UI", 10, "bold"),
            state=tk.DISABLED,
        )
        self.lock_button.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(0, 7))
        self.start_button = tk.Button(
            action_row,
            text="开始点击  F8",
            command=self.toggle_clicking,
            bg=self.ACCENT,
            activebackground=self.ACCENT_HOVER,
            fg="#f8fbff",
            activeforeground="#f8fbff",
            disabledforeground="#e5ebf4",
            relief=tk.FLAT,
            borderwidth=0,
            pady=15,
            cursor="hand2",
            font=("Microsoft YaHei UI", 10, "bold"),
            state=tk.DISABLED,
        )
        self.start_button.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(7, 0))

        status_row = self._panel(view, padx=18, pady=12)
        status_row.pack(fill=tk.X, pady=(14, 0))
        status_dot = tk.Label(
            status_row,
            text="●",
            bg=self.SURFACE,
            fg=self.SUCCESS,
            font=("Segoe UI", 9),
        )
        status_dot.pack(side=tk.LEFT)
        tk.Label(
            status_row,
            textvariable=self.status_var,
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 9),
        ).pack(side=tk.LEFT, padx=(7, 0))
        tk.Label(
            status_row,
            textvariable=self.click_count_var,
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
        ).pack(side=tk.RIGHT, padx=(12, 0))
        ttk.Button(
            status_row,
            text="清零",
            style="AppSecondary.TButton",
            command=self.reset_click_count,
        ).pack(side=tk.RIGHT)

    def _build_account_view(self) -> None:
        view = self._new_view("account")
        profile = self._panel(view, padx=22, pady=18)
        profile.pack(fill=tk.X)
        self.account_avatar = tk.Label(
            profile,
            text="U",
            width=4,
            height=2,
            bg=self.SUCCESS,
            fg="#f5fffc",
            font=("Segoe UI", 13, "bold"),
        )
        self.account_avatar.pack(side=tk.LEFT)
        profile_text = tk.Frame(profile, bg=self.SURFACE)
        profile_text.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(14, 0))
        self.account_name_var = tk.StringVar()
        self.account_email_var = tk.StringVar()
        tk.Label(
            profile_text,
            textvariable=self.account_name_var,
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 12, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            profile_text,
            textvariable=self.account_email_var,
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 9),
            anchor="w",
        ).pack(fill=tk.X, pady=(4, 0))
        self.account_status_label = tk.Label(
            profile,
            text="正常",
            bg=self.SUCCESS_BG,
            fg=self.SUCCESS,
            padx=10,
            pady=5,
            font=("Microsoft YaHei UI", 8, "bold"),
        )
        self.account_status_label.pack(side=tk.RIGHT)

        details = self._panel(view, padx=20, pady=16)
        details.pack(fill=tk.X, pady=14)
        tk.Label(
            details,
            text="基本资料",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 10, "bold"),
            anchor="w",
        ).grid(row=0, column=0, columnspan=4, sticky="ew", pady=(0, 13))
        self.account_detail_vars = {
            "id": tk.StringVar(),
            "username": tk.StringVar(),
            "created": tk.StringVar(),
            "last_login": tk.StringVar(),
        }
        detail_items = (
            ("用户 ID", "id", 1, 0),
            ("用户名", "username", 1, 2),
            ("注册时间", "created", 2, 0),
            ("最近登录", "last_login", 2, 2),
        )
        for label, key, row, column in detail_items:
            details.columnconfigure(column + 1, weight=1)
            tk.Label(
                details,
                text=label,
                bg=self.SURFACE,
                fg=self.MUTED,
                font=("Microsoft YaHei UI", 8),
                anchor="w",
            ).grid(row=row, column=column, sticky="w", padx=(0, 10), pady=7)
            tk.Label(
                details,
                textvariable=self.account_detail_vars[key],
                bg=self.SURFACE,
                fg=self.TEXT,
                font=("Microsoft YaHei UI", 9),
                anchor="w",
            ).grid(row=row, column=column + 1, sticky="ew", padx=(0, 24), pady=7)

        orders_header = tk.Frame(view, bg=self.BACKGROUND)
        orders_header.pack(fill=tk.X, pady=(2, 9))
        tk.Label(
            orders_header,
            text="授权码兑换",
            bg=self.BACKGROUND,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 10, "bold"),
        ).pack(side=tk.LEFT)
        tk.Label(
            orders_header,
            text="登录令牌使用 Windows DPAPI 加密保存在本机",
            bg=self.BACKGROUND,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
        ).pack(side=tk.RIGHT)
        self.orders_container = tk.Frame(view, bg=self.BACKGROUND)
        self.orders_container.pack(fill=tk.BOTH, expand=True)
        self._render_orders()

    def switch_view(self, key: str) -> None:
        if key not in self.views:
            return
        if key == "clicker" and not self.has_tool_access("auto_click"):
            key = "tools"
        for view in self.views.values():
            view.pack_forget()
        self.views[key].pack(fill=tk.BOTH, expand=True)
        self.active_view = key
        title, subtitle = self.PAGE_META[key]
        self.page_title_var.set(title)
        self.page_subtitle_var.set(subtitle)
        for page, button in self.nav_buttons.items():
            active = page == key
            parent_active = page == "tools" and key not in {"home", "tools", "account"}
            button.configure(
                bg=self.ACCENT if active else self.SIDEBAR_HOVER if parent_active else self.SIDEBAR,
                activebackground=self.ACCENT_HOVER if active else self.SIDEBAR_HOVER,
                fg="#f8fbff" if active else "#e0e6eb" if parent_active else "#cbd3da",
                font=("Microsoft YaHei UI", 10, "bold" if active else "normal"),
            )

    def refresh_remote_data(self) -> None:
        if self.remote_loading:
            return
        self.remote_loading = True
        self.refresh_button.configure(state=tk.DISABLED, text="刷新中...")
        self.sync_status_var.set("正在同步账户数据...")
        self.sync_label.configure(fg=self.MUTED)

        def task() -> dict:
            return {
                "user": self.api.current_user(self.token).get("user", {}),
                "tools": self.api.list_tools().get("tools", []),
                "purchases": self.api.my_purchases(self.token).get("purchases", []),
                "orders": [],
            }

        self._run_async(task, self._remote_data_loaded, self._remote_data_failed)

    def _remote_data_loaded(self, payload: dict) -> None:
        self.remote_loading = False
        self.refresh_button.configure(state=tk.NORMAL, text="刷新数据")
        user = payload.get("user")
        if isinstance(user, dict) and user:
            self.user = user
            self.session["user"] = user
            try:
                self.session_store.save_session(self.token, user)
            except OSError:
                pass
        self.tools = self._dict_list(payload.get("tools"))
        self.purchases = self._dict_list(payload.get("purchases"))
        self.orders = self._dict_list(payload.get("orders"))
        self.purchased_codes = {
            str(item.get("toolCode", ""))
            for item in self.purchases
            if item.get("toolCode")
        }
        self.sync_status_var.set("数据已更新")
        self.sync_label.configure(fg=self.SUCCESS)
        self._update_user_views()
        self._render_tools()
        self._render_orders()
        self._render_purchased_navigation()
        self._update_clicker_access_ui()

    def _remote_data_failed(self, exc: Exception) -> None:
        self.remote_loading = False
        self.refresh_button.configure(state=tk.NORMAL, text="重新加载")
        if isinstance(exc, ApiError) and exc.status == 401:
            messagebox.showwarning("登录已过期", "登录状态已过期，请重新登录。")
            self.logout()
            return
        self.sync_status_var.set(self._error_message(exc))
        self.sync_label.configure(fg=self.DANGER)

    @staticmethod
    def _dict_list(value) -> list[dict]:
        if not isinstance(value, list):
            return []
        return [item for item in value if isinstance(item, dict)]

    def _update_user_views(self) -> None:
        email = str(self.user.get("email") or "未填写邮箱")
        username = str(self.user.get("username") or "用户")
        display_name = email.split("@", 1)[0] if "@" in email else username
        initial = (display_name[:1] or "U").upper()
        short_email = email if len(email) <= 24 else email[:21] + "..."
        self.sidebar_name_var.set(display_name)
        self.sidebar_email_var.set(short_email)
        self.sidebar_avatar.configure(text=initial)
        self.account_avatar.configure(text=initial)
        self.account_name_var.set(display_name)
        self.account_email_var.set(email)
        self.home_welcome_var.set(f"你好，{display_name}")

        status = str(self.user.get("status") or "active")
        status_text = "正常" if status == "active" else "已停用"
        self.home_account_var.set(status_text)
        self.account_status_label.configure(
            text=status_text,
            bg=self.SUCCESS_BG if status == "active" else self.DANGER_BG,
            fg=self.SUCCESS if status == "active" else self.DANGER,
        )
        self.account_detail_vars["id"].set(str(self.user.get("id") or "-"))
        self.account_detail_vars["username"].set(username)
        self.account_detail_vars["created"].set(
            self._format_time(self.user.get("createdAt"))
        )
        self.account_detail_vars["last_login"].set(
            self._format_time(self.user.get("lastLoginAt"))
        )

        self.home_purchase_var.set(f"{len(self.purchased_codes)} 项")
        self.home_order_var.set("授权码")
        if self.has_tool_access("auto_click"):
            self.home_tool_status_var.set("已开通")
            self.home_tool_status_label.configure(fg=self.SUCCESS)
        else:
            self.home_tool_status_var.set("未开通")
            self.home_tool_status_label.configure(fg=self.MUTED)
        self.home_clicker_button.configure(
            text="打开工具" if self.has_tool_access("auto_click") else "输入授权码",
            command=lambda: self.switch_view(
                "clicker" if self.has_tool_access("auto_click") else "tools"
            ),
        )

    def _render_tools(self) -> None:
        if not hasattr(self, "tools_container"):
            return
        for child in self.tools_container.winfo_children():
            child.destroy()
        if not self.tools:
            empty = self._panel(self.tools_container, padx=20, pady=24)
            empty.pack(fill=tk.X)
            tk.Label(
                empty,
                text="正在加载工具列表..." if self.remote_loading else "暂无可用工具",
                bg=self.SURFACE,
                fg=self.MUTED,
                font=("Microsoft YaHei UI", 9),
            ).pack(anchor="w")
            return

        for index, tool in enumerate(self.tools):
            code = str(tool.get("code") or "")
            row = self._panel(self.tools_container, padx=20, pady=17)
            row.pack(fill=tk.X, pady=(0 if index == 0 else 6, 6))
            mark = tk.Label(
                row,
                text="点" if code == "auto_click" else "工",
                width=3,
                height=1,
                bg="#e7eefc",
                fg=self.ACCENT,
                font=("Microsoft YaHei UI", 12, "bold"),
            )
            mark.pack(side=tk.LEFT)
            text = tk.Frame(row, bg=self.SURFACE)
            text.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(14, 0))
            tk.Label(
                text,
                text=str(tool.get("name") or code),
                bg=self.SURFACE,
                fg=self.TEXT,
                font=("Microsoft YaHei UI", 11, "bold"),
                anchor="w",
            ).pack(fill=tk.X)
            tk.Label(
                text,
                text=str(tool.get("description") or ""),
                bg=self.SURFACE,
                fg=self.MUTED,
                font=("Microsoft YaHei UI", 9),
                anchor="w",
            ).pack(fill=tk.X, pady=(4, 0))

            action = tk.Frame(row, bg=self.SURFACE)
            action.pack(side=tk.RIGHT, padx=(18, 0))
            if self.has_tool_access(code):
                tk.Label(
                    action,
                    text="已开通",
                    bg=self.SUCCESS_BG,
                    fg=self.SUCCESS,
                    padx=9,
                    pady=4,
                    font=("Microsoft YaHei UI", 8, "bold"),
                ).pack(anchor="e", pady=(0, 6))
                ttk.Button(
                    action,
                    text="打开工具",
                    style="AppPrimary.TButton",
                    command=lambda value=code: self._open_tool(value),
                ).pack(anchor="e")
            else:
                tk.Label(
                    action,
                    text="授权码开通",
                    bg=self.SURFACE,
                    fg=self.MUTED,
                    font=("Microsoft YaHei UI", 8, "bold"),
                ).pack(anchor="e", pady=(0, 6))
                ttk.Button(
                    action,
                    text="输入授权码",
                    style="AppPrimary.TButton",
                    command=lambda value=code: self.start_license_redeem(value),
                ).pack(anchor="e")

    def _render_orders(self) -> None:
        if not hasattr(self, "orders_container"):
            return
        for child in self.orders_container.winfo_children():
            child.destroy()
        panel = self._panel(self.orders_container, padx=18, pady=18)
        panel.pack(fill=tk.X)
        tk.Label(
            panel,
            text="购买授权码后，在工具中心点击“兑换授权码”完成开通。",
            bg=self.SURFACE,
            fg=self.TEXT,
            font=("Microsoft YaHei UI", 9, "bold"),
            anchor="w",
        ).pack(fill=tk.X)
        tk.Label(
            panel,
            text="每个授权码只能由一个账户兑换一次；兑换后工具会永久绑定到当前账户。",
            bg=self.SURFACE,
            fg=self.MUTED,
            font=("Microsoft YaHei UI", 8),
            anchor="w",
        ).pack(fill=tk.X, pady=(6, 0))

    def start_license_redeem(self, tool_code: str = "") -> None:
        tool_name = ""
        if tool_code:
            tool = next(
                (item for item in self.tools if str(item.get("code") or "") == tool_code),
                None,
            )
            if tool is not None:
                tool_name = str(tool.get("name") or tool_code)
        LicenseCodeDialog(self, tool_name)

    def license_code_redeemed(self, purchase: dict) -> None:
        tool_code = str(purchase.get("toolCode") or "")
        if tool_code:
            self.purchases = [
                item for item in self.purchases if item.get("toolCode") != tool_code
            ]
            self.purchases.append(purchase)
            self.purchased_codes.add(tool_code)
        self.sync_status_var.set("授权码兑换成功，工具已开通")
        self.sync_label.configure(fg=self.SUCCESS)
        self._update_user_views()
        self._render_tools()
        self._render_purchased_navigation()
        self._update_clicker_access_ui()
        self.refresh_remote_data()

    def start_tool_purchase(self, tool_code: str) -> None:
        if (
            self.remote_loading
            or tool_code in self.payment_submitting_codes
            or self.has_tool_access(tool_code)
        ):
            return
        tool = next(
            (item for item in self.tools if str(item.get("code") or "") == tool_code),
            None,
        )
        if tool is None:
            messagebox.showerror("无法购买", "工具信息不存在，请刷新后重试。")
            return
        self.payment_submitting_codes.add(tool_code)
        self.sync_status_var.set("正在创建支付宝订单...")
        self.sync_label.configure(fg=self.MUTED)
        self._render_tools()

        def success(response: dict) -> None:
            self.payment_submitting_codes.discard(tool_code)
            order = response.get("order")
            if not isinstance(order, dict) or not response.get("qrCode"):
                failed(ApiError("支付服务返回的数据不完整，请稍后重试。"))
                return
            self.orders = [
                existing
                for existing in self.orders
                if existing.get("orderNo") != order.get("orderNo")
            ]
            self.orders.insert(0, order)
            self.sync_status_var.set("支付宝订单已创建")
            self.sync_label.configure(fg=self.ACCENT)
            self._update_user_views()
            self._render_tools()
            self._render_orders()
            AlipayPaymentDialog(self, tool, response, self._payment_completed)

        def failed(exc: Exception) -> None:
            self.payment_submitting_codes.discard(tool_code)
            self.sync_status_var.set(self._error_message(exc))
            self.sync_label.configure(fg=self.DANGER)
            self._render_tools()

        self._run_async(
            lambda: self.api.create_alipay_payment(self.token, tool_code), success, failed
        )

    def _payment_completed(self) -> None:
        self.sync_status_var.set("支付成功，正在更新已购工具...")
        self.sync_label.configure(fg=self.SUCCESS)
        self.refresh_remote_data()

    def _open_tool(self, tool_code: str) -> None:
        if tool_code == "auto_click":
            self.switch_view("clicker")
            return
        messagebox.showinfo("暂不支持", "该工具的 Windows 功能页尚未接入。")

    def has_tool_access(self, tool_code: str) -> bool:
        return tool_code in self.purchased_codes

    def _has_pending_order(self, tool_code: str) -> bool:
        return any(
            order.get("toolCode") == tool_code and order.get("status") == "pending"
            for order in self.orders
        )

    @staticmethod
    def _format_price(tool: dict) -> str:
        cents = tool.get("priceCents")
        try:
            amount = int(cents) / 100
        except (TypeError, ValueError):
            return "价格待定"
        currency = str(tool.get("currency") or "CNY")
        prefix = "¥" if currency == "CNY" else currency + " "
        suffix = " / 永久" if tool.get("lifetime") else ""
        return f"{prefix}{amount:.2f}{suffix}"

    @staticmethod
    def _format_time(value) -> str:
        try:
            timestamp = int(value)
        except (TypeError, ValueError):
            return "-"
        return datetime.fromtimestamp(timestamp).strftime("%Y-%m-%d %H:%M")

    def _update_clicker_access_ui(self) -> None:
        if not hasattr(self, "clicker_access_notice"):
            return
        allowed = self.has_tool_access("auto_click")
        if allowed:
            self.clicker_access_notice.configure(
                text="自动点击已开通，可以使用全部功能。",
                bg=self.SUCCESS_BG,
                fg=self.SUCCESS,
            )
            if not self.running:
                self.lock_button.configure(state=tk.NORMAL)
                self.start_button.configure(
                    state=tk.NORMAL if self.locked_position else tk.DISABLED,
                    text="开始点击  F8",
                )
        else:
            self.clicker_access_notice.configure(
                text="当前账户尚未开通自动点击，请先在工具中心兑换授权码。",
                bg=self.WARNING_BG,
                fg=self.WARNING,
            )
            if not self.running:
                self.lock_button.configure(state=tk.DISABLED)
                self.start_button.configure(state=tk.DISABLED, text="尚未开通")

    def set_interval(self, milliseconds: int) -> None:
        self.interval_var.set(str(milliseconds))
        self._update_interval_buttons()

    def _update_interval_buttons(self) -> None:
        if not hasattr(self, "quick_buttons"):
            return
        try:
            selected = int(self.interval_var.get())
        except ValueError:
            selected = -1
        for value, button in self.quick_buttons.items():
            active = value == selected
            button.configure(
                bg="#dce8ff" if active else self.SURFACE_ALT,
                fg=self.ACCENT if active else self.TEXT,
                font=("Microsoft YaHei UI", 9, "bold" if active else "normal"),
            )

    def reset_click_count(self) -> None:
        self.click_count = 0
        self.click_count_var.set("已点击 0 次")

    def record_click(self) -> None:
        self.click_count += 1
        self.click_count_var.set(f"已点击 {self.click_count} 次")

    def toggle_topmost(self) -> None:
        self.always_on_top = not self.always_on_top
        self.root.attributes("-topmost", self.always_on_top)
        if self.always_on_top:
            self.topmost_var.set("窗口已置顶")
            self.topmost_button.configure(text="取消置顶（F7）")
        else:
            self.topmost_var.set("窗口未置顶")
            self.topmost_button.configure(text="窗口置顶（F7）")

    def refresh_mouse_position(self) -> None:
        try:
            x, y = get_cursor_position()
            self.current_position_var.set(f"当前鼠标：{x}，{y}")
        except OSError as exc:
            self.current_position_var.set(f"获取鼠标位置失败：{exc}")
        self.root.after(80, self.refresh_mouse_position)

    def poll_hotkeys(self) -> None:
        clicker_active = self.active_view == "clicker"
        if clicker_active and key_pressed(VK_F6):
            self.no_cursor_move_var.set(not self.no_cursor_move_var.get())
        if clicker_active and key_pressed(VK_F7):
            self.toggle_topmost()
        if clicker_active and key_pressed(VK_F9):
            self.lock_position()
        if (clicker_active or self.running) and key_pressed(VK_F8):
            self.toggle_clicking()
        if clicker_active and key_down(VK_CONTROL):
            if key_pressed(VK_1):
                self.set_interval(100)
            if key_pressed(VK_2):
                self.set_interval(500)
            if key_pressed(VK_3):
                self.set_interval(1000)
            if key_pressed(VK_4):
                self.set_interval(2000)
        self.root.after(50, self.poll_hotkeys)

    def lock_position(self) -> None:
        if self.running or not self.has_tool_access("auto_click"):
            return
        try:
            self.locked_position = get_cursor_position()
            x, y = self.locked_position
            self.locked_window_target = get_window_target(x, y)
        except OSError as exc:
            messagebox.showerror("锁定失败", f"无法获取目标位置：{exc}")
            return
        if self.no_cursor_move_var.get():
            hwnd, client_x, client_y = self.locked_window_target
            self.locked_position_var.set(
                f"屏幕 {x}，{y}  ·  窗口 {hwnd}  ·  客户区 {client_x}，{client_y}"
            )
        else:
            self.locked_position_var.set(f"屏幕坐标：{x}，{y}")
        self.start_button.configure(state=tk.NORMAL, text="开始点击  F8")
        self.status_var.set("坐标已锁定")

    def get_interval_seconds(self) -> float | None:
        try:
            milliseconds = int(self.interval_var.get())
        except ValueError:
            messagebox.showerror("间隔无效", "点击间隔必须是整数，例如 100 或 1000。")
            return None
        if milliseconds < 100:
            messagebox.showerror("间隔无效", "点击间隔不能小于 100 毫秒。")
            return None
        return milliseconds / 1000

    def toggle_clicking(self) -> None:
        if self.running:
            self.stop_clicking()
            return
        if not self.has_tool_access("auto_click"):
            messagebox.showwarning("尚未开通", "请先在工具中心兑换自动点击授权码。")
            return
        if self.locked_position is None:
            messagebox.showwarning("尚未锁定", "将鼠标移动到目标位置后按 F9 锁定坐标。")
            return
        interval_seconds = self.get_interval_seconds()
        if interval_seconds is None:
            return

        self.running = True
        self.stop_event.clear()
        self.start_button.configure(text="停止点击  F8", bg=self.DANGER)
        self.lock_button.configure(state=tk.DISABLED)
        self.status_var.set("正在自动点击")
        self.worker_thread = threading.Thread(
            target=self.click_loop,
            args=(
                self.locked_position,
                self.locked_window_target,
                self.no_cursor_move_var.get(),
                interval_seconds,
            ),
            daemon=True,
        )
        self.worker_thread.start()

    def click_loop(
        self,
        position: tuple[int, int],
        window_target: tuple[int, int, int] | None,
        no_cursor_move: bool,
        interval_seconds: float,
    ) -> None:
        while not self.stop_event.is_set():
            try:
                if no_cursor_move and window_target:
                    window_click_at(*window_target)
                else:
                    click_at(*position)
                self.root.after(0, self.record_click)
            except OSError as exc:
                self.stop_event.set()
                self.root.after(0, lambda error=str(exc): self._click_failed(error))
                break
            self.stop_event.wait(interval_seconds)

    def _click_failed(self, error: str) -> None:
        self.running = False
        self.start_button.configure(text="开始点击  F8", bg=self.ACCENT)
        self.lock_button.configure(state=tk.NORMAL)
        self.status_var.set(f"点击失败：{error}")

    def stop_clicking(self) -> None:
        self.running = False
        self.stop_event.set()
        if hasattr(self, "start_button"):
            self.start_button.configure(text="开始点击  F8", bg=self.ACCENT)
            self.lock_button.configure(state=tk.NORMAL)
            self._update_clicker_access_ui()
        self.status_var.set("坐标已锁定" if self.locked_position else "已停止")

    def _run_async(self, task, on_success, on_error) -> None:
        def worker() -> None:
            try:
                result = task()
            except Exception as exc:
                try:
                    self.root.after(0, lambda error=exc: on_error(error))
                except (RuntimeError, tk.TclError):
                    pass
                return
            try:
                self.root.after(0, lambda: on_success(result))
            except (RuntimeError, tk.TclError):
                pass

        threading.Thread(target=worker, daemon=True).start()

    @staticmethod
    def _error_message(exc: Exception) -> str:
        if isinstance(exc, ApiError):
            suffix = f"（请求编号：{exc.request_id}）" if exc.request_id else ""
            return f"{exc}{suffix}"
        if isinstance(exc, NetworkError):
            return str(exc)
        return "数据加载失败，请稍后重试。"

    def close(self) -> None:
        self.stop_clicking()
        self.root.destroy()

    def logout(self) -> None:
        self.stop_clicking()
        try:
            self.session_store.clear_session()
        except OSError as exc:
            messagebox.showerror("退出失败", f"无法清除登录状态：{exc}")
            return
        self.logout_requested = True
        self.root.destroy()

    def run(self) -> bool:
        self.root.mainloop()
        return self.logout_requested


def main() -> None:
    store = SessionStore()
    api_client = ApiClient()
    while True:
        login_session = AuthenticationWindow(api_client, store).run()
        if login_session is None:
            break
        if not AutomaticToolsApp(store, login_session, api_client).run():
            break


if __name__ == "__main__":
    main()
