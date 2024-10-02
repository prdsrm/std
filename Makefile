migrate.create:
	migrate create -ext sql -dir $(folder) -seq $(name)

migrate.up:
	migrate -path $(folder) -database "$(DATABASE_URL)" up

migrate.down:
	migrate -path $(folder) -database "$(DATABASE_URL)" down

migrate.force:
	migrate -path $(folder) -database "$(DATABASE_URL)" force $(version)
