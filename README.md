# currency
Микросервис получает текущий курс BTC по отношению к валютам USD, EUR, GBP, RUB

Данные берутся через сервис https://bitcoinaverage.com/

По умолчанию запускается по адресу http://localhost:8888

Запросы:
- PATCH обновляет курс выбранной вылюты
    - http://localhost:8888/update/type (где type=BTCUSD,BTCEUR,BTCGBP,BTCRUB)
- GET возвращает курс текущей валюты
	- http://localhost:8888/currency/type (где type=BTCUSD,BTCEUR,BTCGBP,BTCRUB)
- GET возвращает курс всех валют
	- http://localhost:8888/currencyall
 - PATCH обновляет курс всех валют
     - http://localhost:8888/updateall

# steam
Микросервис получает список игр из магазина Steam записывает в MongoDB

По умолчанию запускается по адресу http://localhost:8099

Позволяет получить стоимость игры в USD EUR GDBP RUB BTC

Данные берутся через api: store.steampowered.com/api/

Полный список игр: http://api.steampowered.com/ISteamApps/GetAppList/v2

Запросы:
- GET возвращает сведения об игре и стоимость в различных валютах
	- http://localhost:8099/aboutgame/id (где id - уникальный номер игры в steam)
- POST возвращает сведения об игре и стоимости в выбранной валюте
	- http://localhost:8099/game
	- параметры 
	    - appid - уникальный номер игры в steam
	    - currency - тип валюты (USD, EUR, GBP, RUB, BTC)
 - DELETE обнуляет цены игры в MongoDB
     - http://localhost:8099//del/id (где id - уникальный номер игры в steam)


# Сборка Docker
docker network create -d bridge my-bridge-network

cd /currency
docker-compose up

cd /steam
docker-compose up

currency доступен по localhost:8888
steam доступен по localhost:8099
