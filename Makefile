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