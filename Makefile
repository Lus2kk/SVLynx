build:
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/

deploy:
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	ssh root@95.182.97.36 "systemctl start app"

build-deploy: # собрать и задеплоить бэкенд
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	ssh root@95.182.97.36 "systemctl start app"

logs: # посмотреть логи
	ssh root@95.182.97.36 "journalctl -u app -n 50 --no-pager"

front-build: # собирает Vue фронт в папку front/dist/
	cd front && npm run build

front-deploy: # задеплоить фронт
	scp -r front/dist/* root@95.182.97.36:/var/www/svlynx/

full-deploy: # всё сразу
	GOOS=linux GOARCH=amd64 go build -o app ./cmd/app/
	ssh root@95.182.97.36 "systemctl stop app"
	scp app root@95.182.97.36:/root/app
	ssh root@95.182.97.36 "systemctl start app"
	cd front && npm run build
	scp -r front/dist/* root@95.182.97.36:/var/www/svlynx/

