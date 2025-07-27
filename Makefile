lint:
	golangci-lint run --fast -v

lint-fix:
	golangci-lint run --fix -v

deploy:
	flyctl deploy --ha=false
	flyctl status

redeploy:
	flyctl start --ha=false
	flyctl status

turn_off:
	flyctl scale count 0 --yes
	flyctl status

scale_to_1_instance:
	flyctl scale count 1 --yes
	flyctl status

docker_build:
	docker build -t housematee-tgbot:latest .

docker_run_local_with_config_file:
	docker run -it --rm -p 8080:8080 housematee-tgbot:latest

docker_run:
	docker run -it --rm -p 8080:8080 housematee-tgbot:latest \
	-e CONFIG_READER_MODE=secret
	-e ENCODED_CONFIG=your_encoded_value
