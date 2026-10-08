import type { AndroidFeedbackHostProps } from "./android-feedback-host.types";

// iOS and web never import or require the Android native module.
export function AndroidFeedbackHost(_props: AndroidFeedbackHostProps) {
  void _props;
  return null;
}

export function FeedbackWindowAnchor() {
  return null;
}
