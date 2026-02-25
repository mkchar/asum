NAME := asum
MAIN_FILE := cmd/main.go
OUTPUT := $(NAME)

.PHONY: all echo docs run build clean help

all: help

docs:
	@echo "生成 Swagger docs..."
	swag init -g $(MAIN_FILE)

run: 
	@echo "运行程序 $(NAME)..."
	go mod tidy
	go run $(MAIN_FILE)

build: 
	@echo "编译程序 $(NAME)..."
	go mod tidy
	go build -o $(OUTPUT) $(MAIN_FILE)
	@echo "编译成功: $(OUTPUT)"

clean:
	@echo "清除文件..."
	rm -rf $(OUTPUT)
	@echo "Done."

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  docs   生成 Swagger docs..."
	@echo "  run    运行程序"
	@echo "  build  编译程序"
	@echo "  clean  清除文件"
