import type { Action, BattleContext, BattleModifiers, EffectResult, GameState, SetupRequest } from "./types";
import type {
  ApplyActionRequestDTO,
  ApplyActionResponseDTO,
  BattleContextRequestDTO,
  BattleContextResponseDTO,
  LoadGameRequestDTO,
  LoadGameResponseDTO,
  ResolveBattleRequestDTO,
  ResolveBattleResponseDTO,
  ServerErrorResponse,
  SetupRequestDTO,
  SetupResponseDTO,
  ValidActionsRequestDTO,
  ValidActionsResponseDTO
} from "./serverContract";

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

  const json = (await response.json()) as ServerErrorResponse | T;
  if (!response.ok) {
    const message =
      typeof (json as ServerErrorResponse)?.error === "string"
        ? (json as ServerErrorResponse).error
        : "request failed";
    throw new Error(message);
  }

  return normalizeJSONKeys(json) as T;
}

export async function fetchValidActions(
  state: GameState,
  gameID?: string | null
): Promise<{ actions: Action[]; revision: number | null }> {
  const body: ValidActionsRequestDTO = { state, gameID };
  const response = await postJSON<ValidActionsResponseDTO>("/actions/valid", body);
  return {
    actions: response.actions,
    revision: response.revision ?? null
  };
}

export async function applyAction(
  state: GameState,
  action: Action,
  gameID?: string | null,
  clientRevision?: number | null
): Promise<{ state: GameState; effectResult: EffectResult | null; revision: number | null }> {
  const body: ApplyActionRequestDTO = { state, action, gameID, clientRevision };
  const response = await postJSON<ApplyActionResponseDTO>("/actions/apply", body);
  return {
    state: response.state,
    effectResult: response.effectResult ?? null,
    revision: response.revision ?? null
  };
}

export async function resolveBattle(
  state: GameState,
  action: Action,
  attackerRoll: number,
  defenderRoll: number,
  modifiers?: BattleModifiers,
  gameID?: string | null
): Promise<{ action: Action; revision: number | null }> {
  const body: ResolveBattleRequestDTO = {
    state,
    action,
    attackerRoll,
    defenderRoll,
    modifiers,
    useModifiers: modifiers !== undefined,
    gameID
  };
  const response = await postJSON<ResolveBattleResponseDTO>("/battles/resolve", body);
  return {
    action: response.action,
    revision: response.revision ?? null
  };
}

export async function fetchBattleContext(
  state: GameState,
  action: Action,
  gameID?: string | null
): Promise<{ battleContext: BattleContext; revision: number | null }> {
  const body: BattleContextRequestDTO = { state, action, gameID };
  const response = await postJSON<BattleContextResponseDTO>("/battles/context", body);
  return {
    battleContext: response.battleContext,
    revision: response.revision ?? null
  };
}

export async function setupGame(
  request: SetupRequest
): Promise<{ state: GameState; gameID: string | null; revision: number | null }> {
  const body: SetupRequestDTO = request;
  const response = await postJSON<SetupResponseDTO>("/game/setup", body);
  return {
    state: response.state,
    gameID: response.gameID ?? null,
    revision: response.revision ?? null
  };
}

export async function loadGame(
  gameID: string
): Promise<{ state: GameState; gameID: string | null; revision: number | null }> {
  const body: LoadGameRequestDTO = { gameID };
  const response = await postJSON<LoadGameResponseDTO>("/game/load", body);
  return {
    state: response.state,
    gameID: response.gameID ?? gameID,
    revision: response.revision ?? null
  };
}
