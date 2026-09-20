import AsyncStorage from "@react-native-async-storage/async-storage";
import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

const pendingInvitationKey = "tm-pending-team-invitation";
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

export function clearPendingTeamInvitation() {
  return Platform.OS === "web"
    ? AsyncStorage.removeItem(pendingInvitationKey)
    : SecureStore.deleteItemAsync(pendingInvitationKey);
}
