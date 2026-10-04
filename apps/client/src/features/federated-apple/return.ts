/** Compare the complete registered destination, including a custom scheme/host. */
export function appleReturnStatus(url: string, returnURL: string, challengeID: string) {
  const returned = new URL(url);
  const expected = new URL(returnURL);
  if (
    returned.protocol !== expected.protocol ||
    returned.host !== expected.host ||
    returned.pathname !== expected.pathname ||
    returned.username ||
    returned.password ||
    returned.hash ||
    returned.searchParams.getAll("challengeId").length !== 1 ||
    returned.searchParams.get("challengeId") !== challengeID ||
    returned.searchParams.getAll("status").length !== 1
  )
    throw new Error("Apple return mismatch");
  const status = returned.searchParams.get("status");
  if (status !== "ready" && status !== "cancelled" && status !== "failed")
    throw new Error("Apple return status invalid");
  return status;
}
