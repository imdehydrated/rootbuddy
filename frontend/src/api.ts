import type { Action, GameState } from "./types";

const API_BASE = "http://localhost:8080/api";

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

  return json as T;
}

export async function fetchValidActions(state: GameState): Promise<Action[]> {
  const response = await postJSON<{ actions: Action[] }>("/actions/valid", { state });
  return response.actions;
}

export async function applyAction(state: GameState, action: Action): Promise<GameState> {
  const response = await postJSON<{ state: GameState }>("/actions/apply", { state, action });
  return response.state;
}

export async function resolveBattle(
  state: GameState,
  action: Action,
  attackerRoll: number,
  defenderRoll: number
): Promise<Action> {
  const response = await postJSON<{ action: Action }>("/battles/resolve", {
    state,
    action,
    attackerRoll,
    defenderRoll
  });
  return response.action;
}
