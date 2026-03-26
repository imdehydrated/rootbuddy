import type { Action, BattleContext, BattleModifiers, EffectResult, GameState, SetupRequest } from "./types";

const API_BASE = "http://localhost:8080/api";

function normalizeJSONKeys(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(normalizeJSONKeys);
  }

  if (!value || typeof value !== "object") {
    return value;
  }

  const normalized: Record<string, unknown> = {};
  for (const [key, entry] of Object.entries(value as Record<string, unknown>)) {
    const normalizedKey =
      key.toUpperCase() === key ? key.toLowerCase() : `${key[0].toLowerCase()}${key.slice(1)}`;
    normalized[normalizedKey] = normalizeJSONKeys(entry);
  }

  return normalized;
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(body)
  });

  const json = await response.json();
  if (!response.ok) {
    const message = typeof json?.error === "string" ? json.error : "request failed";
    throw new Error(message);
  }

  return normalizeJSONKeys(json) as T;
}

export async function fetchValidActions(state: GameState, gameID?: string | null): Promise<Action[]> {
  const response = await postJSON<{ actions: Action[] }>("/actions/valid", { state, gameID });
  return response.actions;
}

export async function applyAction(
  state: GameState,
  action: Action,
  gameID?: string | null
): Promise<{ state: GameState; effectResult: EffectResult | null }> {
  const response = await postJSON<{ state: GameState; effectResult?: EffectResult }>("/actions/apply", { state, action, gameID });
  return {
    state: response.state,
    effectResult: response.effectResult ?? null
  };
}

export async function resolveBattle(
  state: GameState,
  action: Action,
  attackerRoll: number,
  defenderRoll: number,
  modifiers?: BattleModifiers,
  gameID?: string | null
): Promise<Action> {
  const response = await postJSON<{ action: Action }>("/battles/resolve", {
    state,
    action,
    attackerRoll,
    defenderRoll,
    modifiers,
    useModifiers: modifiers !== undefined,
    gameID
  });
  return response.action;
}

export async function fetchBattleContext(
  state: GameState,
  action: Action,
  gameID?: string | null
): Promise<BattleContext> {
  const response = await postJSON<{ battleContext: BattleContext }>("/battles/context", {
    state,
    action,
    gameID
  });
  return response.battleContext;
}

export async function setupGame(request: SetupRequest): Promise<{ state: GameState; gameID: string | null }> {
  const response = await postJSON<{ state: GameState; gameID?: string }>("/game/setup", request);
  return {
    state: response.state,
    gameID: response.gameID ?? null
  };
}
