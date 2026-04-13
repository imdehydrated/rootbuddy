import { useEffect, useState } from "react";
import { USER_SETTINGS_STORAGE_KEY, defaultUserSettings, type UserSettings } from "../settings";

function loadStoredSettings(): UserSettings {
  if (typeof window === "undefined") {
    return defaultUserSettings;
  }

  try {
    const raw = window.localStorage.getItem(USER_SETTINGS_STORAGE_KEY);
    if (!raw) {
      return defaultUserSettings;
    }

    const parsed = JSON.parse(raw) as Partial<UserSettings>;
    return {
      ...defaultUserSettings,
      ...parsed
    };
  } catch (error) {
    console.warn("Failed to load RootBuddy user settings from localStorage.", error);
    return defaultUserSettings;
  }
}

export function useSettings() {
  const [settings, setSettings] = useState<UserSettings>(() => loadStoredSettings());

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    try {
      window.localStorage.setItem(USER_SETTINGS_STORAGE_KEY, JSON.stringify(settings));
    } catch (error) {
      console.warn("Failed to save RootBuddy user settings to localStorage.", error);
    }
  }, [settings]);

  function updateSetting(setting: keyof UserSettings, value: boolean) {
    setSettings((current) => ({
      ...current,
      [setting]: value
    }));
  }

  function resetSettings() {
    setSettings(defaultUserSettings);
  }

  return {
    settings,
    updateSetting,
    resetSettings
  };
}
