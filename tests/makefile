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
	@go test -v ./...

building_test:
	@$(MAKE) dc_up_test
	@$(MAKE) check_container_test
	@$(MAKE) check_table_employee_test
	@$(MAKE) go_test
	@$(MAKE) dc_down_test