export type AndroidFeedbackHostProps = {
  feedback: {
    id: number;
    message: string;
    kind: "network-error" | "generic-error" | "success";
  } | null;
  onDismiss: (id: number) => void;
  reducedMotion: boolean;
  topInset: number;
};
