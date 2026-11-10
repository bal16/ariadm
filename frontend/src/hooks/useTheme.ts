import { createSignal } from "solid-js";
import { useColorMode } from "@kobalte/core";

export type ThemePreference = "light" | "dark" | "system";

const THEME_STORAGE_KEY = "ariadm-theme";

function isThemePreference(value: string | null): value is ThemePreference {
  return value === "light" || value === "dark" || value === "system";
}

/**
 * Reads the persisted theme preference and keeps it in sync with Kobalte's
 * color mode. Kobalte's storage manager handles persistence.
 */
export function useTheme() {
  const { setColorMode } = useColorMode();
  const storedTheme = localStorage.getItem(THEME_STORAGE_KEY);
  const [themePref, setThemePref] = createSignal<ThemePreference>(
    isThemePreference(storedTheme) ? storedTheme : "system",
  );

  const setTheme = (theme: ThemePreference) => {
    setColorMode(theme);
    setThemePref(theme);
  };

  return { themePref, setTheme };
}
