set path="C:\Program Files\Far Manager";%path%

@echo off
setlocal enabledelayedexpansion

if "%1"=="" (
    echo Usage: %0 ^<file:line:column:^>
    echo Example: %0 internal\api\handlers.go:223:51:
    goto :eof
)

set INPUT=%1

:: Запускаем Python для парсинга строки
for /f "tokens=1,2,3 delims=:" %%a in ('python -c "import sys; parts=sys.argv[1].split(':'); print( parts[0] + ':' + parts[1] + ':' + parts[2] ) " %INPUT% 2^>nul') do (
    set FILE=%%a
    set LINE=%%b
    set COL=%%c
)

:: Проверяем, удалось ли распарсить
if "%FILE%"=="" (
    echo Error: Failed to parse input string
    goto :eof
)

:: Вызываем FAR с параметрами
:: /e - открыть файл для редактирования
:: LINE:COL - позиция курсора
far /e"%LINE%:%COL%" "%FILE%"

if errorlevel 1 (
    echo Error: Failed to open FAR or file not found
)

endlocal