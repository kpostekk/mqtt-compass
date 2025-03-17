dist:
	go build -ldflags="-w -s" -o dist/mqui .
	upx dist/mqui
	@echo "Dist build complete!"

