import AsyncStorage from "@react-native-async-storage/async-storage";
import Constants from "expo-constants";
import { useFocusEffect } from "expo-router";
import * as StoreReview from "expo-store-review";
import { useCallback, useEffect, useRef, useState } from "react";
import { AppState, Platform } from "react-native";

import { createReviewController, reviewStorageKey } from "./policy";

const enabled =
  Platform.OS !== "web" && !__DEV__ && Constants.expoConfig?.extra?.appEnvironment === "production";

const controller = createReviewController({
  read: () => AsyncStorage.getItem(reviewStorageKey),
  write: (state) => AsyncStorage.setItem(reviewStorageKey, state),
  available: () => StoreReview.isAvailableAsync(),
  request: () => StoreReview.requestReview(),
  now: Date.now,
});

export function useStoreReview(id: string | undefined, ready: boolean, blocked: boolean) {
  const focused = useRef(false);
  const counted = useRef(false);
  const safe = useRef(false);
  const [focusRevision, setFocusRevision] = useState(0);
  safe.current = ready && !blocked;

  useFocusEffect(
    useCallback(() => {
      focused.current = true;
      counted.current = false;
      setFocusRevision((revision) => revision + 1);
      return () => {
        focused.current = false;
      };
    }, [id]),
  );

  const isSafe = useCallback(
    () => focused.current && safe.current && AppState.currentState === "active",
    [],
  );

  useEffect(() => {
    if (!enabled || !id || !ready || !focused.current) return;
    if (!counted.current) {
      counted.current = true;
      void controller.visit(isSafe);
    } else if (!blocked) {
      void controller.retryVisitMilestone(isSafe);
    }
  }, [id, ready, blocked, focusRevision, isSafe]);

  return useCallback(async () => {
    if (enabled) await controller.completed(isSafe);
  }, [isSafe]);
}
