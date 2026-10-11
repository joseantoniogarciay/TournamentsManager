/** Civil values deliberately do not become instants in the reader's time zone. */
export function validCivilDate(value: string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value) || value.startsWith("0000")) return false;
  const date = new Date(`${value}T12:00:00Z`);
  return Number.isFinite(date.getTime()) && date.toISOString().slice(0, 10) === value;
}

export function validCivilTime(value: string) {
  return /^([01]\d|2[0-3]):[0-5]\d$/.test(value);
}

export function validTimeZone(value: string) {
  if (!value || value === "Local" || value.length > 100) return false;
  try {
    new Intl.DateTimeFormat("en", { timeZone: value }).format(0);
    return true;
  } catch {
    return false;
  }
}

export function validVenue(value: string) {
  return (
    Array.from(value).length <= 160 &&
    !Array.from(value).some((character) => {
      const code = character.codePointAt(0) ?? 0;
      return code < 32 || (code >= 127 && code <= 159);
    })
  );
}

export function dateToInput(value: string) {
  return value ? value.split("-").reverse().join("/") : "";
}

export function dateFromInput(value: string) {
  if (!value) return "";
  if (!/^\d{2}\/\d{2}\/\d{4}$/.test(value)) return null;
  const date = value.split("/").reverse().join("-");
  return validCivilDate(date) ? date : null;
}

export function canEditScheduling(relationship: string | null | undefined, state: string) {
  return (
    (relationship === "organizer" || relationship === "delegated") &&
    (state === "published" || state === "in_progress")
  );
}
