import ctypes
import re
import threading
import time
import tkinter as tk
from ctypes import wintypes
from tkinter import messagebox, ttk

from auth_client import ApiClient, ApiError, NetworkError, SessionStore


user32 = ctypes.WinDLL("user32", use_last_error=True)

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


class AutomaticToolsApp:
    def __init__(self, session_store: SessionStore, session: dict) -> None:
        self.session_store = session_store
        self.session = session
        self.logout_requested = False
        self.root = tk.Tk()
        self.root.title("AutomaticTools")
        self.root.geometry("400x490")
        self.root.resizable(False, False)
        self.root.attributes("-topmost", True)

        self.locked_position: tuple[int, int] | None = None
        self.locked_window_target: tuple[int, int, int] | None = None
        self.click_count = 0
        self.running = False
        self.always_on_top = True
        self.worker_thread: threading.Thread | None = None
        self.stop_event = threading.Event()

        self.current_position_var = tk.StringVar(value="Current: -, -")
        self.locked_position_var = tk.StringVar(value="Locked: none")
        self.click_count_var = tk.StringVar(value="Clicks sent: 0")
        self.status_var = tk.StringVar(value="Status: stopped")
        self.topmost_var = tk.StringVar(value="Topmost: on")
        self.interval_var = tk.StringVar(value="1000")
        self.no_cursor_move_var = tk.BooleanVar(value=False)

        self.build_ui()
        self.refresh_mouse_position()
        self.poll_hotkeys()
        self.root.protocol("WM_DELETE_WINDOW", self.close)

    def build_ui(self) -> None:
        frame = tk.Frame(self.root, padx=16, pady=12)
        frame.pack(fill=tk.BOTH, expand=True)

        user = self.session.get("user", {})
        account = user.get("email") or user.get("username") or "已登录"
        session_row = tk.Frame(frame)
        session_row.pack(fill=tk.X, pady=(0, 10))
        tk.Label(
            session_row,
            text=f"账户：{account}",
            anchor="w",
            fg="#334155",
        ).pack(side=tk.LEFT, expand=True, fill=tk.X)
        tk.Button(session_row, text="退出登录", command=self.logout).pack(side=tk.RIGHT)

        tk.Label(frame, textvariable=self.current_position_var, anchor="w").pack(
            fill=tk.X, pady=(0, 8)
        )
        tk.Label(frame, textvariable=self.locked_position_var, anchor="w").pack(
            fill=tk.X, pady=(0, 8)
        )
        tk.Label(
            frame,
            text="Hotkeys: F9 lock, F8 start/stop, F7 pin, F6 no-cursor mode",
            anchor="w",
            wraplength=340,
        ).pack(fill=tk.X, pady=(0, 10))

        topmost_row = tk.Frame(frame)
        topmost_row.pack(fill=tk.X, pady=(0, 10))
        tk.Label(topmost_row, textvariable=self.topmost_var, anchor="w").pack(
            side=tk.LEFT, expand=True, fill=tk.X
        )
        self.topmost_button = tk.Button(
            topmost_row, text="Pin off (F7)", width=12, command=self.toggle_topmost
        )
        self.topmost_button.pack(side=tk.RIGHT)

        self.no_cursor_check = tk.Checkbutton(
            frame,
            text="No cursor move (F6)",
            variable=self.no_cursor_move_var,
            anchor="w",
        )
        self.no_cursor_check.pack(fill=tk.X, pady=(0, 12))

        interval_row = tk.Frame(frame)
        interval_row.pack(fill=tk.X, pady=(0, 10))
        tk.Label(interval_row, text="Interval (ms):").pack(side=tk.LEFT)
        interval_entry = tk.Entry(interval_row, textvariable=self.interval_var, width=8)
        interval_entry.pack(side=tk.LEFT, padx=(8, 0))

        quick_row = tk.Frame(frame)
        quick_row.pack(fill=tk.X, pady=(0, 10))
        tk.Button(
            quick_row,
            text="100ms\nCtrl+1",
            height=2,
            command=lambda: self.set_interval(100),
        ).pack(
            side=tk.LEFT, expand=True, fill=tk.X, padx=(0, 6)
        )
        tk.Button(
            quick_row,
            text="500ms\nCtrl+2",
            height=2,
            command=lambda: self.set_interval(500),
        ).pack(side=tk.LEFT, expand=True, fill=tk.X, padx=6)
        tk.Button(
            quick_row,
            text="1000ms\nCtrl+3",
            height=2,
            command=lambda: self.set_interval(1000),
        ).pack(side=tk.LEFT, expand=True, fill=tk.X, padx=6)
        tk.Button(
            quick_row,
            text="2s\nCtrl+4",
            height=2,
            command=lambda: self.set_interval(2000),
        ).pack(side=tk.LEFT, expand=True, fill=tk.X, padx=(6, 0))

        button_row = tk.Frame(frame)
        button_row.pack(fill=tk.X, pady=(0, 12))
        button_row.columnconfigure(0, weight=1, uniform="main_actions")
        button_row.columnconfigure(1, weight=1, uniform="main_actions")
        self.lock_button = tk.Button(
            button_row, text="Lock\nF9", height=2, command=self.lock_position
        )
        self.lock_button.grid(row=0, column=0, sticky="ew", padx=(0, 6))

        self.start_button = tk.Button(
            button_row,
            text="Start\nF8",
            height=2,
            command=self.toggle_clicking,
            state=tk.DISABLED,
        )
        self.start_button.grid(row=0, column=1, sticky="ew", padx=(6, 0))

        count_row = tk.Frame(frame)
        count_row.pack(fill=tk.X, pady=(0, 10))
        tk.Label(count_row, textvariable=self.click_count_var, anchor="w").pack(
            side=tk.LEFT, expand=True, fill=tk.X
        )
        tk.Button(count_row, text="Reset count", command=self.reset_click_count).pack(
            side=tk.RIGHT
        )

        tk.Label(frame, textvariable=self.status_var, anchor="w").pack(fill=tk.X)

    def set_interval(self, milliseconds: int) -> None:
        self.interval_var.set(str(milliseconds))

    def reset_click_count(self) -> None:
        self.click_count = 0
        self.click_count_var.set("Clicks sent: 0")

    def record_click(self) -> None:
        self.click_count += 1
        self.click_count_var.set(f"Clicks sent: {self.click_count}")

    def toggle_topmost(self) -> None:
        self.always_on_top = not self.always_on_top
        self.root.attributes("-topmost", self.always_on_top)
        if self.always_on_top:
            self.topmost_var.set("Topmost: on")
            self.topmost_button.config(text="Pin off (F7)")
        else:
            self.topmost_var.set("Topmost: off")
            self.topmost_button.config(text="Pin on (F7)")

    def refresh_mouse_position(self) -> None:
        try:
            x, y = get_cursor_position()
            self.current_position_var.set(f"Current: {x}, {y}")
        except OSError as exc:
            self.current_position_var.set(f"Current: failed ({exc})")
        self.root.after(80, self.refresh_mouse_position)

    def poll_hotkeys(self) -> None:
        if key_pressed(VK_F6):
            self.no_cursor_move_var.set(not self.no_cursor_move_var.get())
        if key_pressed(VK_F7):
            self.toggle_topmost()
        if key_pressed(VK_F9):
            self.lock_position()
        if key_pressed(VK_F8):
            self.toggle_clicking()
        if key_down(VK_CONTROL):
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
        if self.running:
            return
        self.locked_position = get_cursor_position()
        x, y = self.locked_position
        self.locked_window_target = get_window_target(x, y)
        if self.no_cursor_move_var.get():
            hwnd, client_x, client_y = self.locked_window_target
            self.locked_position_var.set(
                f"Locked: {x}, {y} / hwnd {hwnd} client {client_x}, {client_y}"
            )
        else:
            self.locked_position_var.set(f"Locked: {x}, {y}")
        self.start_button.config(state=tk.NORMAL)
        self.status_var.set("Status: locked")

    def get_interval_seconds(self) -> float | None:
        try:
            milliseconds = int(self.interval_var.get())
        except ValueError:
            messagebox.showerror("Invalid interval", "Interval must be an integer, like 100 or 1000.")
            return None

        if milliseconds < 100:
            messagebox.showerror("Invalid interval", "Interval cannot be less than 100 ms.")
            return None
        return milliseconds / 1000

    def toggle_clicking(self) -> None:
        if self.running:
            self.stop_clicking()
            return

        if self.locked_position is None:
            messagebox.showwarning(
                "No locked position",
                "Move the mouse to the target and press F9, or click Lock current.",
            )
            return

        interval_seconds = self.get_interval_seconds()
        if interval_seconds is None:
            return

        self.running = True
        self.stop_event.clear()
        self.start_button.config(text="Stop\nF8")
        self.lock_button.config(state=tk.DISABLED)
        self.status_var.set("Status: clicking")
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
                self.root.after(0, lambda: self.status_var.set(f"Status: click failed ({exc})"))
                self.stop_event.set()
                break
            self.stop_event.wait(interval_seconds)

    def stop_clicking(self) -> None:
        self.running = False
        self.stop_event.set()
        self.start_button.config(text="Start\nF8")
        self.lock_button.config(state=tk.NORMAL)
        if self.locked_position:
            self.status_var.set("Status: locked")
        else:
            self.status_var.set("Status: stopped")

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
        if not AutomaticToolsApp(store, login_session).run():
            break


if __name__ == "__main__":
    main()
