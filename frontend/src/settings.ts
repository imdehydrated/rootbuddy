export type UserSettings = {
  showGameLog: boolean;
  showVPTracker: boolean;
  showCardTray: boolean;
  compactCards: boolean;
};

export const defaultUserSettings: UserSettings = {
  showGameLog: true,
  showVPTracker: true,
  showCardTray: true,
  compactCards: false
};

export const USER_SETTINGS_STORAGE_KEY = "rootbuddy_user_settings_v1";
