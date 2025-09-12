ENV_FILE=.env

ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
endif

dc_up_test:
	@echo "Starting Docker containers..."
	docker compose -f $(COMPOSE_FILE) up -d

dc_down_test:
	@echo "Stopping Docker containers..."
	docker compose -f $(COMPOSE_FILE) down

check_container_test:
	@echo "Checking Postgres container ..."
	@docker exec postgres-server-test pg_isready -U postgres -d postgres > NUL 2>&1 && \
		(echo Postgres is available) || \
		(echo Postgres is not available!)

check_goose_version_test:
	@echo "Checking Goose version in container '$(DB_CONTAINER)'..."
	@docker exec $(DB_CONTAINER) goose --version || echo "Goose is not installed!"

goose_up_test:
	@echo "Running Goose migrations in container '$(DB_CONTAINER)'..."
	@docker exec $(DB_CONTAINER) sh -c '\
		GOOSE_DRIVER=$(DB_DRIVER_NAME) \
		GOOSE_DBSTRING="$(GOOSE_DBSTRING)" \
		goose -dir . up'

goose_down_test:
	@echo "Running Goose rollback in container '$(DB_CONTAINER)'..."
	@docker exec $(DB_CONTAINER) sh -c '\
    		GOOSE_DRIVER=$(DB_DRIVER_NAME) \
    		GOOSE_DBSTRING="$(GOOSE_DBSTRING)" \
    		goose -dir . down'

check_table_employee_test:
	@echo "Checking if '$(DB_TABLE_NAME_EM)' table exists in the database..."
	@docker exec $(DB_CONTAINER) psql -U postgres -d $(DB_NAME) -tAc "\
		SELECT CASE WHEN EXISTS ( \
			SELECT 1 FROM information_schema.tables \
			WHERE table_name='$(DB_TABLE_NAME_EM)' \
		) THEN 'Table exists' ELSE 'Table does NOT exist' END;"
	@echo "Columns in table '$(DB_TABLE_NAME_EM)':"
	@docker exec $(DB_CONTAINER) psql -U postgres -d $(DB_NAME) -tAc "\
		SELECT column_name \
		FROM information_schema.columns \
		WHERE table_name='$(DB_TABLE_NAME_EM)' \
		ORDER BY ordinal_position;"

go_test:
	@echo "Running Go tests..."
	@go test -v -timeout 30s ./...

container_run_test:
	@$(MAKE) dc_up_test
	@$(MAKE) check_container_test
	@$(MAKE) goose_up_test
	@$(MAKE) check_table_employee_test

deleting_container_test:
	@$(MAKE) goose_down_test
	@$(MAKE) dc_down_test