deploy:
	flyctl deploy
	flyctl scale count 1 --yes
	flyctl status

redeploy:
	flyctl start
	flyctl status

turn_off:
	flyctl scale count 0 --yes
	flyctl status

scale_to_1_instance:
	flyctl scale count 1 --yes
	flyctl status



