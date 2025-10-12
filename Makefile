# ╔════════════════════════════════════════════════════════════════════════╗
#                          ⭐ Start App ⭐
# ╚════════════════════════════════════════════════════════════════════════╝
.PHONY: run
run:
	go mod download && go run cmd/main.go --config=/app/config/application.yaml

.PHONY: docker-up
docker-up:
	docker-compose -f docker-compose.yml up --build

up:
	docker compose up -d

down:
	docker compose down

clean:
	docker compose down -v

logs:
	docker compose logs -f --tail=200


include migrations.mk

# ╔════════════════════════════════════════════════════════════════════════╗
#                 📜 Swagger Documentation Generation
# ╚════════════════════════════════════════════════════════════════════════╝

install-swagger2openapi:
	@echo "Installing swagger2openapi globally..."
	@npm install -g swagger2openapi
	@echo "swagger2openapi installed successfully."

GEN_DIR := "./docs"
.PHONY: swag
swag: swagger-gen
	@echo "Converting client Swagger to OpenAPI..."
	@swagger2openapi --outfile $(GEN_DIR)/user_openapi.yaml $(GEN_DIR)/user_swagger.json
	@echo "Ensuring 'servers' section exists in user OpenAPI..."
	@if ! grep -q "^servers:" $(GEN_DIR)/user_openapi.yaml; then \
	  echo "servers:" | cat - $(GEN_DIR)/user_openapi.yaml > $(GEN_DIR)/user_openapi.tmp && mv $(GEN_DIR)/user_openapi.tmp $(GEN_DIR)/user_openapi.yaml; \
	fi
	@echo "Adding additional servers to client OpenAPI..."
	@sed -i '/^servers:/a \
  - url: http://127.0.0.1:8080\n\
    description: Local access\n\
  - url: https://user-app.brands.dev.itemcloud.ru\n\
    description: Dev Swagger UI' $(GEN_DIR)/user_openapi.yaml


.PHONY: swagger-gen
swagger-gen:
	@swag init -d ./internal/api -g api.go --instanceName user --parseDependency --parseInternal --outputTypes json
	@echo "Swagger documentation generated."
	@git add ./docs/*
