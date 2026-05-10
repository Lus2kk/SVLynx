build:
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/

deploy:
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	scp -r migrations root@95.182.97.36:/root/SVLynx/migrations
	ssh root@95.182.97.36 "systemctl start app"

build-deploy:
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	scp -r migrations root@95.182.97.36:/root/SVLynx/migrations
	ssh root@95.182.97.36 "systemctl start app"

logs:
	ssh root@95.182.97.36 "journalctl -u app -n 50 --no-pager"

front-build:
	cd front && npm run build

front-deploy:
	scp -r front/dist/* root@95.182.97.36:/var/www/svlynx/

full-deploy:
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	scp -r migrations root@95.182.97.36:/root/SVLynx/migrations
	ssh root@95.182.97.36 "systemctl start app"
	cd front && npm run build
	scp -r front/dist/* root@95.182.97.36:/var/www/svlynx/

voice-build:
	cd voice_server && g++ -std=c++17 -O2 -o voice_server voice_server.cpp -lpthread

voice-deploy:
	scp voice_server/voice_server root@95.182.97.36:/root/voice_server
	ssh root@95.182.97.36 "systemctl restart voice"

media-build:
	cd media_server && docker run --rm --platform linux/amd64 -v $(pwd):/src -w /src gcc:latest g++ -std=c++17 -O2 -o media_server_linux media_server.cpp -lpthread

media-deploy:
	ssh root@95.182.97.36 "systemctl stop media_server || true"
	scp media_server/media_server_linux root@95.182.97.36:/root/media_server
	ssh root@95.182.97.36 "systemctl start media_server"

media-logs:
	ssh root@95.182.97.36 "journalctl -u media_server -n 50 --no-pager"

caddy-deploy:
	scp Caddyfile root@95.182.97.36:/etc/caddy/Caddyfile
	ssh root@95.182.97.36 "systemctl reload caddy"