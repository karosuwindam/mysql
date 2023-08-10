
reset:
	sudo rm -rf mysql-docker/db_data

up:
	docker compose -f mysql-docker/compose.yml up -d

down: 
	docker compose -f mysql-docker/compose.yml down

test-run:
	go test -v -run $(target)

test:
	go test -v