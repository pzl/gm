TARGET=gm
SRCS=$(shell find . -type f -name *.go)


ALL: bin/$(TARGET)

bin/:
	mkdir -p $@

bin/$(TARGET): $(SRCS) bin/ cmd/$(TARGET)/assets.go
	CGO_ENABLED=0 go build -o bin/ ./cmd/$(TARGET)


cmd/$(TARGET)/assets.go: cmd/$(TARGET)/assets_gen.go frontend/dist/index.html
	go generate ./cmd/$(TARGET)


frontend/dist/index.html: frontend/node_modules frontend/static/favicon.ico $(shell find -type f -name '*.vue') $(shell find frontend -type f -name '*.js')
	cd frontend && npm run build && npm run generate

frontend/node_modules:
	cd frontend && npm install

frontend/static/favicon.ico: frontend/static/icon.png
	convert -background transparent $< -define icon:auto-resize=16,32,48,64,256 $@

run:
	cd frontend && npm run dev

clean:
	$(RM) -rf bin/
	$(RM) -rf frontend/dist

.PHONY: clean run

