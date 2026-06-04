from asgiref.sync import async_to_sync
from channels.layers import get_channel_layer

from planner.serializers import TeamPlannerDeskSerializer


def planner_team_group_name(team_id: int) -> str:
    return f"planner_team_{int(team_id)}"


def broadcast_team_desk_update(desk, action: str = "updated") -> None:
    channel_layer = get_channel_layer()
    if channel_layer is None:
        return

    async_to_sync(channel_layer.group_send)(
        planner_team_group_name(desk.team_id),
        {
            "type": "desk.updated",
            "action": action,
            "teamId": int(desk.team_id),
            "desk": TeamPlannerDeskSerializer(desk).data,
        },
    )