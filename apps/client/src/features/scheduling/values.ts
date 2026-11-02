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

/** Numeric masks keep civil values local; separators never need a keyboard key. */
export function formatScheduleInput(kind: "date" | "time", input: string, previous: string) {
  const separator = kind === "date" ? "/" : ":";
  let digits = input.replace(/[^0-9]/g, "").slice(0, kind === "date" ? 8 : 4);
  const previousDigits = previous.replace(/[^0-9]/g, "");
  // Backspace over an inserted separator also removes the preceding digit.
  if (previous.endsWith(separator) && input === previous.slice(0, -1)) digits = digits.slice(0, -1);
  const groups =
    kind === "date"
      ? [digits.slice(0, 2), digits.slice(2, 4), digits.slice(4)]
      : [digits.slice(0, 2), digits.slice(2)];
  let formatted = groups[0] ?? "";
  if (digits.length >= 2) formatted += separator + groups[1];
  if (kind === "date" && digits.length >= 4) formatted += separator + groups[2];
  // Deletion must always remain possible, including while correcting a middle segment.
  if (digits.length < previousDigits.length) return formatted;
  const first = groups[0] ?? "";
  const second = groups[1] ?? "";
  const validPart = (part: string, maximum: number, minimum: number) =>
    !part ||
    (part.length === 1
      ? Number(part) <= Math.floor(maximum / 10)
      : Number(part) >= minimum && Number(part) <= maximum);
  if (
    !validPart(first, kind === "date" ? 31 : 23, kind === "date" ? 1 : 0) ||
    !validPart(second, kind === "date" ? 12 : 59, kind === "date" ? 1 : 0)
  )
    return previous;
  if (kind === "date" && digits.length === 8 && dateFromInput(formatted) === null) return previous;
  return formatted;
}
