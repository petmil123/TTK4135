mkdir -p bin
go build -o main/elev_sys elevator-system/main.go
wget -O bin/hall_request_assigner https://github.com/TTK4145/Project-resources/releases/download/v1.1.3/hall_request_assigner
chmod +x bin/hall_request_assigner