import { useEffect, useState } from "react";

import { getGetUsernameAvailabilityUrl } from "@/api/generated/usernames/usernames";

const usernamePattern = /^[a-z0-9_]{3,30}$/;
const debounceMilliseconds = 400;
const defaultAPIBaseURL = "http://127.0.0.1:8080/v1";

export type UsernameAvailabilityStatus =
  "idle" | "checking" | "available" | "unavailable" | "rate-limited" | "error";

export function useUsernameAvailability(username: string) {
  const isValid = usernamePattern.test(username);
  const [status, setStatus] = useState<UsernameAvailabilityStatus>("idle");

  useEffect(() => {
    if (!isValid) {
      setStatus("idle");
      return;
    }

    const controller = new AbortController();
    setStatus("checking");
    const timer = setTimeout(() => {
      void checkUsernameAvailability(username, controller.signal)
        .then((nextStatus) => setStatus(nextStatus))
        .catch((error: unknown) => {
          if (error instanceof Error && error.name === "AbortError") return;
          setStatus("error");
        });
    }, debounceMilliseconds);

    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [isValid, username]);

  return { isValid, status };
}

async function checkUsernameAvailability(username: string, signal: AbortSignal) {
  const baseURL = (process.env.EXPO_PUBLIC_API_BASE_URL ?? defaultAPIBaseURL).replace(/\/$/, "");
  const response = await fetch(`${baseURL}${getGetUsernameAvailabilityUrl(username)}`, {
    signal,
  });
  if (response.status === 200) {
    const body: { available: boolean } = await response.json();
    return body.available ? "available" : "unavailable";
  }
  if (response.status === 429) return "rate-limited";
  throw new Error("username availability request failed");
}
