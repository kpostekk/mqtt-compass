dist/mqui:
	go build -ldflags="-w -s" -o dist/mqui .
	upx -9 dist/mqui > /dev/null
	du -sh dist/mqui
	@echo "Dist build complete!"
