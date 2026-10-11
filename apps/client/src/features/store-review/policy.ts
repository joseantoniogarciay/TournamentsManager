export const reviewIntervalMs = 60 * 24 * 60 * 60 * 1000;
export const reviewStorageKey = "tournaments-manager.store-review.v1";

type ReviewState = {
  visits: number;
  visitMilestoneConsumed: boolean;
  lastAttemptAt: number | null;
};

type ReviewPorts = {
  read: () => Promise<string | null>;
  write: (state: string) => Promise<void>;
  available: () => Promise<boolean>;
  request: () => Promise<void>;
  now: () => number;
};

function parseState(raw: string | null): ReviewState {
  if (raw === null) return { visits: 0, visitMilestoneConsumed: false, lastAttemptAt: null };
  const value = JSON.parse(raw) as Partial<ReviewState> | null;
  if (
    !value ||
    !Number.isInteger(value.visits) ||
    typeof value.visits !== "number" ||
    value.visits < 0 ||
    value.visits > 10 ||
    typeof value.visitMilestoneConsumed !== "boolean" ||
    (value.lastAttemptAt !== null &&
      (typeof value.lastAttemptAt !== "number" ||
        !Number.isFinite(value.lastAttemptAt) ||
        value.lastAttemptAt < 0))
  ) {
    throw new Error("Invalid local review state");
  }
  return value as ReviewState;
}

/** One queue per installation prevents both triggers from requesting together. */
export function createReviewController(ports: ReviewPorts) {
  let queue = Promise.resolve();
  function run(operation: () => Promise<void>) {
    // Reviews are optional. Storage, platform and parsing failures stay silent.
    queue = queue.then(operation).catch(() => undefined);
    return queue;
  }

  async function attempt(state: ReviewState, isSafe: () => boolean) {
    const now = ports.now();
    if (
      !isSafe() ||
      (state.lastAttemptAt !== null && now - state.lastAttemptAt < reviewIntervalMs) ||
      !(await ports.available()) ||
      !isSafe()
    ) {
      return;
    }
    // A successful API call does not prove the dialog appeared or a review was sent.
    await ports.write(
      JSON.stringify({
        ...state,
        lastAttemptAt: now,
        visitMilestoneConsumed: state.visitMilestoneConsumed || state.visits === 10,
      }),
    );
    if (isSafe()) await ports.request();
  }

  return {
    visit(isSafe: () => boolean) {
      return run(async () => {
        const state = parseState(await ports.read());
        state.visits = Math.min(10, state.visits + 1);
        await ports.write(JSON.stringify(state));
        if (state.visits === 10 && !state.visitMilestoneConsumed) await attempt(state, isSafe);
      });
    },
    retryVisitMilestone(isSafe: () => boolean) {
      return run(async () => {
        const state = parseState(await ports.read());
        if (state.visits === 10 && !state.visitMilestoneConsumed) await attempt(state, isSafe);
      });
    },
    completed(isSafe: () => boolean) {
      return run(async () => attempt(parseState(await ports.read()), isSafe));
    },
  };
}
