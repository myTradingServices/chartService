start:
	go run main.go
mock-gen-repo: 
	mockery --dir ./internal/repository --all --output ./internal/service/mock --with-expecter
mock-gen-service:
	mockery --dir ./internal/service --all --output ./internal/rpc/mock --with-expecter