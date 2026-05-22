# Deploy CRM + testing module

## 1. Prepare env

```bash
cp .env.example .env
nano .env
```

Required values:

- `APP_DOMAIN` - CRM domain, for example `crm-159-194-211-118.nip.io`.
- `TESTING_DOMAIN` - testing module domain, for example `tests-159-194-211-118.nip.io`.
- `DJANGO_ALLOWED_HOSTS` - include CRM domain.
- `DJANGO_CSRF_TRUSTED_ORIGINS` - include `https://CRM_DOMAIN`.
- `DJANGO_CORS_ALLOWED_ORIGINS` - include CRM and testing domains.
- `TESTING_SERVICE_URL=https://TESTING_DOMAIN`.
- `TESTING_SERVICE_TOKEN` and `CRM_TOKEN` must be the same shared secret. In root compose this is wired automatically through `TESTING_SERVICE_TOKEN`.
- VK settings stay in the same root `.env`.

## 2. Start services

```bash
docker compose up -d --build
```

The compose file starts:

- CRM PostgreSQL database
- CRM Django backend
- CRM automation worker
- testing PostgreSQL database
- testing Go backend
- testing React frontend
- public Caddy proxy for both domains

## 3. Create CRM superuser

```bash
docker compose exec backend python manage.py createsuperuser
```

## 4. Open services

- CRM: `https://$APP_DOMAIN`
- Testing module: `https://$TESTING_DOMAIN`

## 5. SSO flow

1. CRM creates a one-time ticket through `/api/users/integration/testing/sso-link/`.
2. Testing frontend opens `/sso?ticket=...`.
3. Testing backend calls CRM `/api/users/integration/testing/sso-exchange/` with `X-Service-Token`.
4. Testing backend creates or updates a local user.
5. CRM `organizer` becomes testing `manager`; CRM `student` becomes testing `intern`.
6. Interns are redirected to the linked test when a matching event/specialization config exists.
7. After finishing the test, testing backend sends the result back to CRM.
