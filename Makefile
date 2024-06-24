default: img

.PHONY:img
img:
	docker build --file server/Dockerfile ./server -t example/server:latest
	docker system prune -f




