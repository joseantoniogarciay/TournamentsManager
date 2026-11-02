import { useEffect, useState } from "react";
import { TournamentIllustration } from "@/shared/ui/tournament-illustration";

let presented = false;

export function GuestWelcome() {
  const [animate] = useState(() => !presented);
  // Effects never run during static web export: SSR must not consume a visit.
  useEffect(() => {
    presented = true;
  }, []);
  return <TournamentIllustration variant="welcome" animate={animate} />;
}
