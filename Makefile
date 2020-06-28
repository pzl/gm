TARGET=manager
SRCS=$(shell find . -type f -name '*.go')

ALL: $(TARGET)

$(TARGET): $(SRCS) assets.go
	go build -o $@

assets.go: assets_gen.go frontend/dist/index.html
	go generate

frontend/dist/index.html: frontend/node_modules $(shell find frontend -type f -name '*.vue') $(shell find frontend -type f -name '*.js')
	cd frontend && npm run build

frontend/node_modules: frontend/package.json frontend/package-lock.json
	cd frontend && npm install

run:
	cd frontend && npm run dev


clean:
	$(RM) $(TARGET)
	$(RM) -rf frontend/dist


.PHONY: clean run