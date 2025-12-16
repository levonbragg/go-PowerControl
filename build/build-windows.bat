@echo off
echo Building Go PowerControl for Windows...
cd ..
wails build -platform windows/amd64
echo Build complete! Output is in build\bin\
pause
