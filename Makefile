deploy:
	flyctl deploy
	flyctl scale count 1
	flyctl status

redeploy:
	flyctl start
	flyctl status

turn_off:
	flyctl scale count 0
	flyctl status

scale_1_instance:
	flyctl scale count 1
	flyctl status



