@echo off
echo Generating Swagger documentation...
cd /d "%~dp0src"
swag fmt
swag init --output ./swagger
echo Swagger documentation generated successfully!
echo Files created/updated:
echo - swagger/docs.go
echo - swagger/swagger.json
echo - swagger/swagger.yaml
pause