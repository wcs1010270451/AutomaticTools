import ctypes
import threading
import time
import tkinter as tk
from ctypes import wintypes
from tkinter import messagebox


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


class AutomaticToolsApp:
    def __init__(self) -> None:
        self.root = tk.Tk()
        self.root.title("AutomaticTools")
        self.root.geometry("380x445")
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

    def run(self) -> None:
        self.root.mainloop()


if __name__ == "__main__":
    AutomaticToolsApp().run()
