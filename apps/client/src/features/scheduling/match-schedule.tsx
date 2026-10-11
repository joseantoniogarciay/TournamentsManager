import type { Match } from "@/api/generated/models";
import { getCurrentLanguage } from "@/shared/i18n/locale";
import { Text } from "@/shared/ui";

export function MatchSchedule({
  match,
  tournamentVenue,
}: {
  match: Match;
  tournamentVenue?: string;
}) {
  if (match.state === "bye") return null;
  const venue = match.venue || tournamentVenue;
  let date = "";
  if (match.scheduledDate) {
    date = new Intl.DateTimeFormat(getCurrentLanguage(), {
      day: "numeric",
      month: "short",
      year: "numeric",
      timeZone: "UTC",
    }).format(new Date(`${match.scheduledDate}T12:00:00Z`));
  }
  const when = [date, match.scheduledTime].filter(Boolean).join(" · ");
  if (!when && !venue) return null;
  return (
    <>
      {when ? (
        <Text color="secondary">
          {when}
          {match.scheduledTime ? ` · ${match.timeZone}` : ""}
        </Text>
      ) : null}
      {venue ? <Text color="secondary">{venue}</Text> : null}
    </>
  );
}
