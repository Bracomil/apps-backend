migrate-up:
	migrate -path cmd/internal/db/migrations -database "postgres://bracomil-api:troque_isso_em_prod@localhost:5432/bracomil-api?sslmode=disable" up

migrate-down:
	migrate -path cmd/internal/db/migrations -database "postgres://bracomil-api:troque_isso_em_prod@localhost:5432/bracomil-api?sslmode=disable" down

migrate-version:
	migrate -path cmd/internal/db/migrations -database "postgres://bracomil-api:troque_isso_em_prod@localhost:5432/bracomil-api?sslmode=disable" version

db-cli

ifeq ($(firstword $(MAKECMDGOALS)),migrate-force)
  # Captura o argumento que vem depois de migrate-force
  VERSION := $(word 2,$(MAKECMDGOALS))
  # Cria um alvo fantasma para o número da versão não dar erro de "No rule to make target"
  $(eval $(VERSION):;@:)
endif

migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "Erro: informe a versão. Ex: make migrate-force 4"; exit 1; fi
	migrate -path cmd/internal/db/migrations -database "postgres://bracomil-api:troque_isso_em_prod@localhost:5432/bracomil-api?sslmode=disable" force $(VERSION)
