import AsyncStorage from "@react-native-async-storage/async-storage";
import { color } from "@tournaments-manager/design-tokens";
import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useColorScheme } from "react-native";

export type ThemePreference = "system" | "light" | "dark";
export type ResolvedTheme = "light" | "dark";

type ThemeColors = {
  surface: { canvas: string; default: string; subtle: string };
  text: { primary: string; secondary: string; placeholder: string; inverse: string };
  indicator: { default: string };
  border: { default: string; focus: string; error: string };
  feedback: { success: string; error: string };
};

type PreferencesContextValue = {
  themePreference: ThemePreference;
  resolvedTheme: ResolvedTheme;
  colors: ThemeColors;
  setThemePreference: (theme: ThemePreference) => void;
};

const storageKey = "tournaments-manager.theme-preference";

const lightColors: ThemeColors = {
  surface: { canvas: "#F8FAFC", default: "#FFFFFF", subtle: "#F1F5F9" },
  text: { primary: "#101828", secondary: "#475467", placeholder: "#98A2B3", inverse: "#FFFFFF" },
  indicator: { default: color.brand.primary },
  border: { default: "#D0D5DD", focus: color.border.focus, error: "#D92D20" },
  feedback: { success: "#027A48", error: "#D92D20" },
};

const darkColors: ThemeColors = {
  surface: { canvas: "#101828", default: "#182230", subtle: "#1D2939" },
  text: { primary: "#F9FAFB", secondary: "#D0D5DD", placeholder: "#98A2B3", inverse: "#101828" },
  indicator: { default: color.text.inverse },
  border: { default: "#475467", focus: color.border.focus, error: "#FDA29B" },
  feedback: { success: "#6CE9A6", error: "#FDA29B" },
};

const PreferencesContext = createContext<PreferencesContextValue | null>(null);

export function PreferencesProvider({ children }: PropsWithChildren) {
  const systemTheme = useColorScheme();
  const [themePreference, setThemePreferenceState] = useState<ThemePreference>("system");
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    void AsyncStorage.getItem(storageKey)
      .then((stored) => {
        if (stored === "system" || stored === "light" || stored === "dark") {
          setThemePreferenceState(stored);
        }
      })
      .finally(() => setIsHydrated(true));
  }, []);

  const setThemePreference = (theme: ThemePreference) => {
    setThemePreferenceState(theme);
    void AsyncStorage.setItem(storageKey, theme);
  };

  const resolvedTheme: ResolvedTheme =
    themePreference === "system" ? (systemTheme === "dark" ? "dark" : "light") : themePreference;

  const value = useMemo(
    () => ({
      themePreference,
      resolvedTheme,
      colors: resolvedTheme === "dark" ? darkColors : lightColors,
      setThemePreference,
    }),
    [resolvedTheme, themePreference],
  );

  return isHydrated ? (
    <PreferencesContext.Provider value={value}>{children}</PreferencesContext.Provider>
  ) : null;
}

export function usePreferences() {
  const value = useContext(PreferencesContext);
  if (!value) throw new Error("usePreferences must be used inside PreferencesProvider");
  return value;
}
