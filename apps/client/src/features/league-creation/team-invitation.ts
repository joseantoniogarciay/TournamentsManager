import AsyncStorage from "@react-native-async-storage/async-storage";
import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

const pendingInvitationKey = "tm-pending-team-invitation";
const pendingInvitationDraftKey = "tm-pending-team-invitation-draft";
const invitationTokenPattern = /^[A-Za-z0-9_-]{43}$/;

export function isTeamInvitationToken(value: string | null | undefined): value is string {
  return typeof value === "string" && invitationTokenPattern.test(value);
}

export async function rememberPendingTeamInvitation(token: string) {
  if (!isTeamInvitationToken(token)) return false;
  if (Platform.OS === "web") await AsyncStorage.setItem(pendingInvitationKey, token);
  else await SecureStore.setItemAsync(pendingInvitationKey, token);
  return true;
}

export async function getPendingTeamInvitation() {
  const token =
    Platform.OS === "web"
      ? await AsyncStorage.getItem(pendingInvitationKey)
      : await SecureStore.getItemAsync(pendingInvitationKey);
  return isTeamInvitationToken(token) ? token : null;
}

export async function rememberPendingTeamInvitationName(token: string, name: string) {
  if (!isTeamInvitationToken(token)) return;
  const draft = JSON.stringify({ token, name });
  if (Platform.OS === "web") await AsyncStorage.setItem(pendingInvitationDraftKey, draft);
  else await SecureStore.setItemAsync(pendingInvitationDraftKey, draft);
}

export async function getPendingTeamInvitationName(token: string) {
  const raw =
    Platform.OS === "web"
      ? await AsyncStorage.getItem(pendingInvitationDraftKey)
      : await SecureStore.getItemAsync(pendingInvitationDraftKey);
  if (!raw) return null;
  try {
    const draft: unknown = JSON.parse(raw);
    if (
      typeof draft === "object" &&
      draft !== null &&
      "token" in draft &&
      draft.token === token &&
      "name" in draft &&
      typeof draft.name === "string"
    ) {
      return draft.name;
    }
  } catch {
    return null;
  }
  return null;
}

export async function clearPendingTeamInvitation() {
  return Platform.OS === "web"
    ? Promise.all([
        AsyncStorage.removeItem(pendingInvitationKey),
        AsyncStorage.removeItem(pendingInvitationDraftKey),
      ])
    : Promise.all([
        SecureStore.deleteItemAsync(pendingInvitationKey),
        SecureStore.deleteItemAsync(pendingInvitationDraftKey),
      ]);
}
