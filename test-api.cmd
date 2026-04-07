rem  start server before
rem go run cmd/server/main.go --config ./config_examples/postgres_storage.json
rem go run cmd/server/main.go --config ./config_examples/memory_storage.json


@echo off
set BASE_URL=http://localhost:8080

echo === Testing Anti-BruteForce API ===

echo 1. Check auth (should be true):
curl -s -X POST %BASE_URL%/check -H "Content-Type: application/json" -d "{\"login\":\"alice\",\"password\":\"12345\",\"ip\":\"192.168.1.100\"}"
echo.

echo 2. Add to blacklist:
curl -s -X POST %BASE_URL%/blacklist/add -H "Content-Type: application/json" -d "{\"subnet\":\"192.168.1.0/24\"}"
echo.

echo 3. Check auth after blacklist (should be false):
curl -s -X POST %BASE_URL%/check -H "Content-Type: application/json" -d "{\"login\":\"alice\",\"password\":\"12345\",\"ip\":\"192.168.1.100\"}"
echo.

echo 4. Get blacklist:
curl -s %BASE_URL%/blacklist
echo.

echo 5. Add to whitelist:
curl -s -X POST %BASE_URL%/whitelist/add -H "Content-Type: application/json" -d "{\"subnet\":\"10.0.0.0/8\"}"
echo.

echo 6. Check IP from whitelist (should be true):
curl -s -X POST %BASE_URL%/check -H "Content-Type: application/json" -d "{\"login\":\"bob\",\"password\":\"67890\",\"ip\":\"10.0.0.50\"}"
echo.

echo 7. Get whitelist:
curl -s %BASE_URL%/whitelist
echo.

echo 8. Remove from blacklist:
curl -s -X POST %BASE_URL%/blacklist/remove -H "Content-Type: application/json" -d "{\"subnet\":\"192.168.1.0/24\"}"
echo.

echo 9. Remove from whitelist:
curl -s -X POST %BASE_URL%/whitelist/remove -H "Content-Type: application/json" -d "{\"subnet\":\"10.0.0.0/8\"}"
echo.

echo 10. Reset buckets:
curl -s -X POST %BASE_URL%/reset -H "Content-Type: application/json" -d "{\"login\":\"alice\",\"ip\":\"192.168.1.100\"}"
echo.

echo 11. Get stats:
curl -s %BASE_URL%/stats
echo.

echo === Done ===