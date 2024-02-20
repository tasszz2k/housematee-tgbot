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
