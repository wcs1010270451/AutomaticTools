@echo off
cd /d "%~dp0"
set "AUTOMATIC_TOOLS_API_BASE_URL=http://127.0.0.1:8088"
python automatic_tools_gui.py
