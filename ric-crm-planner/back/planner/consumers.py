from channels.db import database_sync_to_async
from channels.generic.websocket import AsyncJsonWebsocketConsumer
from django.db.models import Q

from planner.models import PlannerWorkspaceState, TeamPlannerDesk
from planner.realtime import planner_team_group_name
from planner.serializers import TeamPlannerDeskSerializer
from users.models import Event
from users.permissions import has_curator_or_admin_role


def _to_int(value):
    try:
        if value is None:
            return None
        return int(value)
    except (TypeError, ValueError):
        return None


def _member_ids_from_team(team):
    if not isinstance(team, dict):
        return []
    member_ids = team.get("memberIds", team.get("member_ids", []))
    if not isinstance(member_ids, list):
        return []
    return [_to_int(member_id) for member_id in member_ids]


def _team_matches_user(team, user_id):
    if not isinstance(team, dict):
        return False
    return user_id in _member_ids_from_team(team) or user_id == _to_int(team.get("curatorId", team.get("curator_id")))


def _team_event_id(team):
    if not isinstance(team, dict):
        return None
    return _to_int(team.get("eventId", team.get("event_id")))


@database_sync_to_async
def _can_access_team(user, team_id: int) -> bool:
    if not user or not user.is_authenticated:
        return False

    if has_curator_or_admin_role(user):
        return True

    user_id = _to_int(user.id)
    desk = TeamPlannerDesk.objects.filter(team_id=team_id).first()
    if desk:
        member_ids = desk.member_ids if isinstance(desk.member_ids, list) else []
        if user_id in [_to_int(member_id) for member_id in member_ids]:
            return True
        if user_id == _to_int(desk.curator_id):
            return True

    workspace = PlannerWorkspaceState.objects.order_by("id").first()
    teams = workspace.teams if workspace and isinstance(workspace.teams, list) else []
    team = next((item for item in teams if isinstance(item, dict) and _to_int(item.get("id")) == team_id), None)
    if _team_matches_user(team, user_id):
        return True

    event_id = _team_event_id(team)
    if event_id:
        return Event.objects.filter(pk=event_id).filter(Q(leader=user) | Q(organizers=user)).exists()

    return False


@database_sync_to_async
def _get_team_desk_payload(team_id: int) -> dict:
    desk, _ = TeamPlannerDesk.objects.get_or_create(team_id=team_id)
    return TeamPlannerDeskSerializer(desk).data


class PlannerTeamDeskConsumer(AsyncJsonWebsocketConsumer):
    async def connect(self):
        self.team_id = int(self.scope["url_route"]["kwargs"]["team_id"])
        self.group_name = planner_team_group_name(self.team_id)

        if not await _can_access_team(self.scope.get("user"), self.team_id):
            await self.close(code=4403)
            return

        await self.channel_layer.group_add(self.group_name, self.channel_name)
        await self.accept()
        await self.send_json(
            {
                "type": "desk.snapshot",
                "teamId": self.team_id,
                "desk": await _get_team_desk_payload(self.team_id),
            }
        )

    async def disconnect(self, close_code):
        if hasattr(self, "group_name"):
            await self.channel_layer.group_discard(self.group_name, self.channel_name)

    async def receive_json(self, content, **kwargs):
        if content.get("type") == "ping":
            await self.send_json({"type": "pong", "teamId": self.team_id})

    async def desk_updated(self, event):
        await self.send_json(
            {
                "type": "desk.updated",
                "teamId": event.get("teamId", self.team_id),
                "action": event.get("action", "updated"),
                "desk": event.get("desk"),
            }
        )