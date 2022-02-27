docker.rabbitmq:
	docker run --rm -d \
		 --name dev-rabbitmq \
		 --hostname dev-rabbitmq \
		 -v ${HOME}/dev-rabbitmq:/var/lib/rabbitmq \
		 -v ${PWD}/configs/definitions.json:/opt/definitions.json:ro \
		 -v ${PWD}/configs/rabbitmq.config:/etc/rabbitmq/rabbitmq.config:ro \
		 -p 5672:5672 \
		 -p 15672:15672 \
		 rabbitmq:3-management
