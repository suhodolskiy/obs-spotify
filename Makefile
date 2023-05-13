build_app:
	go build -ldflags="-s -w" -o ./dist/obs-spotify ./main.go