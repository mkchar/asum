echo:
	@echo $(NAME)

docs:
	@swag init -g cmd/main.go