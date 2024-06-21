default: run

.PHONY:img
img:
	docker build --force-rm . -t example/exapp:latest
	docker system prune -f




