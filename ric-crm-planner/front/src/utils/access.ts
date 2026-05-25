import type { Event } from "../types/event";
import type { User } from "../types/user";

export function isProjectantRole(role?: string) {
  const normalized = String(role || "").toLowerCase();
  return normalized === "student" || normalized.includes("project");
}

export function isOrganizerRole(role?: string) {
  const normalized = String(role || "").toLowerCase();
  return normalized === "organizer" || normalized.includes("admin") || normalized.includes("curator");
}

export function isGlobalOrganizer(user?: User | null) {
  return Boolean(user?.isGlobalOrganizer || user?.isSuperuser || user?.isStaff);
}

export function getManagedEventIds(user?: User | null) {
  return new Set(
    (user?.managedEventIds || [])
      .map((id) => Number(id))
      .filter((id) => Number.isFinite(id) && id > 0)
  );
}

export function canManageEvent(user: User | null | undefined, event?: Event | null) {
  if (!user || !event) return false;
  if (isGlobalOrganizer(user)) return true;

  const managedEventIds = getManagedEventIds(user);
  const eventId = Number(event.id);
  if (Number.isFinite(eventId) && managedEventIds.has(eventId)) return true;

  const userId = Number(user.id);
  const leaderId = Number(event.leader);
  if (Number.isFinite(userId) && Number.isFinite(leaderId) && leaderId === userId) return true;

  return (event.organizerIds || []).some((id) => Number(id) === userId);
}
