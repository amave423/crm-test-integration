from django.urls import re_path

from planner.consumers import PlannerTeamDeskConsumer

websocket_urlpatterns = [
    re_path(r"^ws/planner/teams/(?P<team_id>\d+)/$", PlannerTeamDeskConsumer.as_asgi()),
]