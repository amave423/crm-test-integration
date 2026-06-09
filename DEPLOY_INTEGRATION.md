# Развертывание CRM и модуля тестирования

Документ описывает первый запуск проекта на VPS и настройку CI/CD через GitHub Actions. Инструкция рассчитана на разработчика или администратора, которому передают проект.

## 1. Состав проекта

Проект запускается как единый сервис. Разделять CRM и модуль тестирования в production не нужно, потому что CRM использует тестирование через SSO и получает результаты обратно через API.

Компоненты:

- CRM backend: Django API, заявки, мероприятия, автоматизация, планировщик, VK-интеграция.
- CRM frontend: React + TypeScript интерфейс CRM.
- Testing backend: backend конструктора тестов.
- Testing frontend: интерфейс конструктора и прохождения тестов.
- PostgreSQL CRM: база CRM.
- PostgreSQL testing: база модуля тестирования.
- Caddy: HTTPS, frontend, reverse proxy.

## 2. Требования к VPS

Минимум для нормального production-запуска без сборки на сервере:

- 2 CPU;
- 2 GB RAM;
- 30 GB SSD;
- Ubuntu 22.04/24.04 или близкий Debian-based дистрибутив;
- открытые порты `80` и `443`.

Если собирать Docker-образы прямо на VPS, 10-15 GB диска и 1 GB RAM обычно не хватает. Production-вариант должен скачивать готовые образы из GHCR.

## 3. Домены и DNS

Основной домен CRM:

```text
meetuppoint.ru
```

Поддомен модуля тестирования:

```text
test.meetuppoint.ru
```

Обе A-записи должны вести на IP VPS:

```text
meetuppoint.ru       A  5.181.108.146
www.meetuppoint.ru   A  5.181.108.146
test.meetuppoint.ru  A  5.181.108.146
```

Проверка на сервере или локально:

```bash
nslookup meetuppoint.ru
nslookup test.meetuppoint.ru
```

Caddy сам выпустит HTTPS-сертификаты Let's Encrypt, если DNS уже указывает на VPS и порты `80/443` открыты.

## 4. Установка Docker на VPS

Подключиться к серверу:

```bash
ssh corleone@5.181.108.146
```

Установить зависимости:

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg git
```

Добавить репозиторий Docker:

```bash
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

Установить Docker:

```bash
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

Добавить пользователя в группу Docker:

```bash
sudo usermod -aG docker $USER
```

После этого нужно выйти с сервера и зайти снова.

Проверить:

```bash
docker --version
docker compose version
```

## 5. Клонирование проекта на VPS

```bash
cd ~
git clone <URL_РЕПОЗИТОРИЯ> crm-test-integration
cd ~/crm-test-integration
```

Если репозиторий приватный, настройте SSH-ключ или используйте HTTPS-доступ с token.

## 6. Файл `.env` на VPS

Файл `.env` должен лежать в корне проекта на VPS:

```text
/home/corleone/crm-test-integration/.env
```

Создать:

```bash
cd ~/crm-test-integration
cp .env.example .env
nano .env
```

Обязательные production-значения:

```env
APP_DOMAIN=meetuppoint.ru
TESTING_DOMAIN=test.meetuppoint.ru

VITE_API_BASE=
VITE_USE_MOCK=false
VITE_TESTING_URL=https://test.meetuppoint.ru

DJANGO_DEBUG=0
DJANGO_ALLOWED_HOSTS=meetuppoint.ru,5.181.108.146,localhost,127.0.0.1,backend
DJANGO_CSRF_TRUSTED_ORIGINS=https://meetuppoint.ru
DJANGO_CORS_ALLOW_ALL_ORIGINS=0
DJANGO_CORS_ALLOWED_ORIGINS=https://meetuppoint.ru,https://test.meetuppoint.ru

DB_HOST=db
DB_PORT=5432
TESTING_SERVICE_URL=https://test.meetuppoint.ru
VK_CHAT_LINK_BASE_URL=https://meetuppoint.ru
VK_BOT_FRONTEND_URL=https://meetuppoint.ru
```

Обязательно заменить заглушки на реальные секреты:

- `DJANGO_SECRET_KEY`;
- `DB_PASSWORD`;
- `TESTING_SERVICE_TOKEN`;
- `TESTING_DB_PASSWORD`;
- `TESTING_ADMIN_EMAIL`;
- `TESTING_ADMIN_PASSWORD`;
- `TESTING_JWT_SECRET`;
- `VK_GROUP_ID`;
- `VK_ACCESS_TOKEN`;
- `VK_CALLBACK_SECRET`;
- `VK_CONFIRMATION_CODE`.

Суперпользователь CRM через `.env` не создается. Его нужно создать вручную после первого запуска.

## 7. GitHub Secrets для CI/CD

Открыть в GitHub:

```text
Repository -> Settings -> Secrets and variables -> Actions -> New repository secret
```

Добавить secrets:

```text
VPS_HOST=5.181.108.146
```

```text
VPS_USER=corleone
```

```text
VPS_PROJECT_PATH=/home/corleone/crm-test-integration
```

```text
VPS_SSH_KEY=<приватный SSH-ключ целиком>
```

В GitHub поле `Name` содержит только имя секрета, например `VPS_HOST`. Поле `Secret` содержит только значение. Писать `VPS_HOST=...` в поле значения не нужно.

## 8. SSH-ключ для GitHub Actions

Создать ключ на локальном компьютере:

```powershell
ssh-keygen -t ed25519 -C "github-actions-meetuppoint" -f "$env:USERPROFILE\.ssh\github_actions_meetuppoint"
```

Публичный ключ:

```powershell
Get-Content "$env:USERPROFILE\.ssh\github_actions_meetuppoint.pub"
```

Добавить публичный ключ на VPS:

```bash
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

В `authorized_keys` публичный ключ вставляется одной строкой.

Приватный ключ:

```powershell
Get-Content "$env:USERPROFILE\.ssh\github_actions_meetuppoint" -Raw
```

Его нужно вставить в GitHub Secret `VPS_SSH_KEY` полностью, вместе со строками:

```text
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

Ключ оставляется многострочным. В одну строку его превращать не нужно.

## 9. Как работает workflow

Файл:

```text
.github/workflows/ci-cd.yml
```

Запускается:

- при push в `main`;
- при pull request в `main`, но без деплоя;
- вручную через `workflow_dispatch`.

Build job:

1. Клонирует репозиторий.
2. Создает `.env` из `.env.example` для проверки compose.
3. Вычисляет `IMAGE_PREFIX` автоматически:

```text
ghcr.io/<github_owner>/<repo_name>
```

4. Собирает и публикует образы:

```text
<IMAGE_PREFIX>-backend:<sha>
<IMAGE_PREFIX>-backend:latest
<IMAGE_PREFIX>-web:<sha>
<IMAGE_PREFIX>-web:latest
<IMAGE_PREFIX>-testing-backend:<sha>
<IMAGE_PREFIX>-testing-backend:latest
<IMAGE_PREFIX>-testing-web:<sha>
<IMAGE_PREFIX>-testing-web:latest
```

Deploy job:

1. Подключается к VPS по SSH.
2. Переходит в `VPS_PROJECT_PATH`.
3. Проверяет наличие `.env`.
4. Делает `git pull`.
5. Логинится в GHCR.
6. Экспортирует `IMAGE_PREFIX` и `IMAGE_TAG`.
7. Выполняет:

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml exec -T backend python manage.py migrate
docker compose -f docker-compose.prod.yml ps
docker image prune -f
```

## 10. Первый запуск через CI/CD

Порядок:

1. Настроить DNS.
2. Установить Docker на VPS.
3. Клонировать репозиторий на VPS.
4. Создать и заполнить `.env` на VPS.
5. Добавить GitHub Secrets.
6. Сделать push в `main` или запустить workflow вручную.
7. Проверить containers:

```bash
cd ~/crm-test-integration
docker compose -f docker-compose.prod.yml ps
```

8. Создать superuser CRM:

```bash
docker compose -f docker-compose.prod.yml exec backend python manage.py createsuperuser
```

9. Открыть:

```text
https://meetuppoint.ru/admin/
```

## 11. Ручной production-деплой без GitHub Actions

Использовать, если нужно перезапустить сервер вручную или проверить деплой без workflow.

```bash
cd ~/crm-test-integration
git pull
export IMAGE_PREFIX=ghcr.io/<github_owner>/<repo_name>
export IMAGE_TAG=latest
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml exec -T backend python manage.py migrate
docker compose -f docker-compose.prod.yml ps
```

Если образы приватные:

```bash
echo <GHCR_TOKEN> | docker login ghcr.io -u <GITHUB_USERNAME> --password-stdin
```

Token должен иметь право `read:packages`.

## 12. Частые ошибки CI/CD

### `env file .env not found`

На VPS нет `.env` в корне проекта.

Решение:

```bash
cd ~/crm-test-integration
cp .env.example .env
nano .env
```

### `required variable IMAGE_PREFIX is missing`

Команда `docker compose -f docker-compose.prod.yml ...` запущена вручную без `IMAGE_PREFIX`.

Решение:

```bash
export IMAGE_PREFIX=ghcr.io/<github_owner>/<repo_name>
export IMAGE_TAG=latest
docker compose -f docker-compose.prod.yml up -d
```

### `error from registry: denied`

Сервер не имеет доступа к GHCR-образам.

Решение:

```bash
echo <GHCR_TOKEN> | docker login ghcr.io -u <GITHUB_USERNAME> --password-stdin
```

Или сделать packages публичными в настройках GitHub Packages.

### Сайт открывается как страница хостинга или старый сайт

DNS указывает не на VPS.

Проверить:

```bash
nslookup meetuppoint.ru
```

### HTTPS не выпускается

Проверить:

- DNS указывает на VPS;
- порты `80` и `443` открыты;
- контейнер `web` запущен;
- в `.env` указаны правильные `APP_DOMAIN` и `TESTING_DOMAIN`.

Команды:

```bash
sudo ufw status
docker compose -f docker-compose.prod.yml logs --tail=100 web
```

### CSRF ошибка в Django admin

Проверить `.env`:

```env
DJANGO_CSRF_TRUSTED_ORIGINS=https://meetuppoint.ru
DJANGO_ALLOWED_HOSTS=meetuppoint.ru,5.181.108.146,localhost,127.0.0.1,backend
```

После изменения:

```bash
docker compose -f docker-compose.prod.yml up -d
```

## 13. Передача проекта другому владельцу GitHub

После transfer repository:

1. Новый владелец заново добавляет GitHub Secrets.
2. На VPS проверить, что `git pull` работает из нового remote.
3. Если используется ручной деплой, поменять `IMAGE_PREFIX`:

```bash
export IMAGE_PREFIX=ghcr.io/<new_owner>/<repo_name>
```

4. Запустить workflow из нового репозитория.
5. Проверить, что GHCR packages доступны VPS.

Workflow вычисляет `IMAGE_PREFIX` автоматически, поэтому в `.github/workflows/ci-cd.yml` обычно ничего менять не нужно, если структура имен образов осталась прежней.