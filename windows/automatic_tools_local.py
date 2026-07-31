import os


os.environ.setdefault("AUTOMATIC_TOOLS_API_BASE_URL", "http://127.0.0.1:8088")

from automatic_tools_gui import main


if __name__ == "__main__":
    main()
