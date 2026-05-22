#!/bin/sh
set -e

python - <<'PY'
import os
import socket
import time

host = os.getenv("DB_HOST")
port = int(os.getenv("DB_PORT", "5432"))
if host:
    deadline = time.time() + 60
    while True:
        try:
            with socket.create_connection((host, port), timeout=3):
                break
        except OSError:
            if time.time() > deadline:
                raise
            time.sleep(1)
PY

if [ "${RUN_MIGRATIONS:-1}" = "1" ]; then
    python manage.py migrate --noinput
fi

if [ -n "${CRM_ADMIN_EMAIL:-}" ] && [ -n "${CRM_ADMIN_PASSWORD:-}" ]; then
    python - <<'PY'
import os

import django

os.environ.setdefault("DJANGO_SETTINGS_MODULE", "api.settings")
django.setup()

from django.contrib.auth import get_user_model
from users.models import Profile

email = os.environ["CRM_ADMIN_EMAIL"].strip().lower()
password = os.environ["CRM_ADMIN_PASSWORD"]
first_name = os.getenv("CRM_ADMIN_FIRST_NAME", "CRM").strip() or "CRM"
last_name = os.getenv("CRM_ADMIN_LAST_NAME", "Admin").strip() or "Admin"

User = get_user_model()
user, created = User.objects.get_or_create(
    username=email,
    defaults={
        "email": email,
        "first_name": first_name,
        "last_name": last_name,
        "is_staff": True,
        "is_superuser": True,
        "is_active": True,
    },
)

changed_fields = []
if user.email != email:
    user.email = email
    changed_fields.append("email")
if not user.is_staff:
    user.is_staff = True
    changed_fields.append("is_staff")
if not user.is_superuser:
    user.is_superuser = True
    changed_fields.append("is_superuser")
if not user.is_active:
    user.is_active = True
    changed_fields.append("is_active")
if created or os.getenv("CRM_ADMIN_RESET_PASSWORD") == "1":
    user.set_password(password)
    changed_fields.append("password")
if changed_fields:
    user.save()

Profile.objects.get_or_create(
    user=user,
    defaults={
        "surname": last_name,
        "name": first_name,
        "email": email,
        "course": 1,
    },
)

print(f"CRM admin {'created' if created else 'exists'}: {email}")
PY
fi

if [ "${RUN_COLLECTSTATIC:-1}" = "1" ]; then
    python manage.py collectstatic --noinput
fi

exec "$@"
