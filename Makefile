NAME := asum
MAIN_FILE := cmd/main.go
OUTPUT := $(NAME)

PID_FILE := .run/$(NAME).pid
LOG_FILE := ../app.log

PORT := 63117
HEALTH_URL := http://127.0.0.1:$(PORT)/healthz
START_TIMEOUT := 50

.PHONY: all docs run build start stop status clean help

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
	go build -ldflags="-s -w" -o $(OUTPUT) $(MAIN_FILE)
	@echo "编译成功: $(OUTPUT)"
	go clean -modcache

start: build
	@echo "启动程序 $(NAME)..."
	@mkdir -p .run
	@$(MAKE) stop >/dev/null 2>&1 || true
	@nohup ./$(OUTPUT) >> $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)
	@echo "PID: $$(cat $(PID_FILE)), 日志: $(LOG_FILE)"
	@echo "等待健康检查通过（超时 $(START_TIMEOUT)s）：$(HEALTH_URL)"
	@i=0; \
	while [ $$i -lt $(START_TIMEOUT) ]; do \
		code="$$(curl -sS --max-time 1 -o /dev/null -w '%{http_code}' '$(HEALTH_URL)' || true)"; \
		if [ "$$code" = "200" ]; then \
			echo "启动成功：healthz OK (200)"; exit 0; \
		fi; \
		echo "等待中：health http_code=$$code"; \
		sleep 1; i=$$((i+1)); \
	done; \
	echo "启动失败/超时：$(START_TIMEOUT)s 内 healthz 未通过。请查看日志：$(LOG_FILE)"; \
	exit 1

stop:
	@if [ -f "$(PID_FILE)" ]; then \
		pid=$$(cat $(PID_FILE)); \
		if kill -0 $$pid >/dev/null 2>&1; then \
			kill $$pid; \
			echo "已停止：PID $$pid"; \
		else \
			echo "PID 文件存在但进程不存在：$$pid"; \
		fi; \
		rm -f $(PID_FILE); \
	else \
		echo "未运行（无 PID 文件）"; \
	fi

status:
	@if [ -f "$(PID_FILE)" ] && kill -0 $$(cat $(PID_FILE)) >/dev/null 2>&1; then \
		echo "运行中：PID $$(cat $(PID_FILE))"; \
	else \
		echo "未运行"; \
	fi

clean:
	@echo "清除文件..."
	@rm -rf $(OUTPUT) .run
	@echo "Done."

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  docs    生成 Swagger docs..."
	@echo "  run     运行程序"
	@echo "  build   编译程序"
	@echo "  start   编译并启动（带启动成功检测）"
	@echo "  stop    停止（基于 PID 文件）"
	@echo "  status  查看状态"
	@echo "  clean   清除文件"
