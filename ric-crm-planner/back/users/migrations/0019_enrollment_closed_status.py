from django.db import migrations


ENROLLMENT_CLOSED_STATUS = "Набор завершён"


def seed_status(apps, schema_editor):
    Status = apps.get_model("users", "Status")
    Status.objects.get_or_create(
        name=ENROLLMENT_CLOSED_STATUS,
        defaults={
            "description": "Набор участников завершен, проектант ожидает подтверждения участия в ПШ.",
            "is_positive": True,
        },
    )


def unseed_status(apps, schema_editor):
    Status = apps.get_model("users", "Status")
    Status.objects.filter(name=ENROLLMENT_CLOSED_STATUS).delete()


class Migration(migrations.Migration):
    dependencies = [
        ("users", "0018_crm_automation_attachment"),
    ]

    operations = [
        migrations.RunPython(seed_status, unseed_status),
    ]
