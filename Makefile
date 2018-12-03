GO_SRCS=manager.go system.go
SERVER=manager

FRONTEND=frontend/dist


ALL: $(SERVER) $(FRONTEND)


install-deps:
	go mod vendor
	cd frontend && npm install

dev:
	cd frontend && npm run dev
	go run manage.go

$(SERVER): $(GO_SRCS)
	go build -mod=vendor

$(FRONTEND): frontend/
	cd frontend && npm run generate


clean:
	$(RM) manager
	$(RM) -rf $(FRONTEND)