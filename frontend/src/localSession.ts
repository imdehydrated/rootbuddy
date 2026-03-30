import type { GameState } from "./types";

const STORAGE_KEY = "rootbuddy_saved_game_v1";

export type SavedSession = {
  state: GameState;
  gameID: string | null;
  revision: number | null;
  savedAt: string;
};

export function loadSavedSession(): SavedSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }

    const parsed = JSON.parse(raw) as Partial<SavedSession>;
    if (!parsed || typeof parsed !== "object" || !parsed.state) {
      return null;
    }

    return {
      state: parsed.state,
      gameID: parsed.gameID ?? null,
      revision: parsed.revision ?? null,
      savedAt: parsed.savedAt ?? ""
    };
  } catch {
    return null;
  }
}

export function saveSavedSession(session: SavedSession): boolean {
  if (typeof window === "undefined") {
    return false;
  }

  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
    return true;
  } catch {
    return false;
  }
}

export function clearSavedSession(): void {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
  }
}
