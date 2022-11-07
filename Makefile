
DOCKER_COMPOSE := docker-compose

_check:
	@test "$(target)" != "" && echo "" || ($(MAKE) help && exit 1)

up:
	$(DOCKER_COMPOSE) up -d
down:
	$(DOCKER_COMPOSE) down
rm:
	sudo rm -rf ./services
log:
	$(DOCKER_COMPOSE) logs

test: 
	go test -v
test-run: _check 
	go test -v --run $(target)

help:
	@echo "make up|down|rm|test|test-run"
	@echo "make test-run target="