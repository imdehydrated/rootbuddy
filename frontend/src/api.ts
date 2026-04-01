import type { Action, BattleContext, BattleModifiers, BattlePrompt, EffectResult, GameState, Lobby, LobbyPlayer, SetupRequest } from "./types";
import type {
  ApplyActionRequestDTO,
  ApplyActionResponseDTO,
  BattleContextRequestDTO,
  BattleContextResponseDTO,
  BattlePromptResponseDTO,
  BattleResponseRequestDTO,
  ClaimFactionRequestDTO,
  CreateLobbyRequestDTO,
  CreateLobbyResponseDTO,
  LoadGameRequestDTO,
  LoadGameResponseDTO,
  LobbyResponseDTO,
  JoinLobbyRequestDTO,
  JoinLobbyResponseDTO,
  LeaveLobbyResponseDTO,
  ReadyLobbyRequestDTO,
  ResolveBattleRequestDTO,
  ResolveBattleResponseDTO,
  ServerErrorResponse,
  StartLobbyResponseDTO,
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

type RequestOptions = {
  playerToken?: string | null;
};

function requestHeaders(options?: RequestOptions): HeadersInit {
  const headers: Record<string, string> = {
    "Content-Type": "application/json"
  };
  if (options?.playerToken) {
    headers["X-Player-Token"] = options.playerToken;
  }
  return headers;
}

async function postJSON<T>(path: string, body: unknown, options?: RequestOptions): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: requestHeaders(options),
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

async function getJSON<T>(path: string, options?: RequestOptions): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "GET",
    headers: options?.playerToken ? { "X-Player-Token": options.playerToken } : undefined
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

function lobbyResult(response: LobbyResponseDTO): { lobby: Lobby; self: LobbyPlayer } {
  return {
    lobby: response.lobby,
    self: response.self
  };
}

export async function fetchValidActions(
  state: GameState,
  gameID?: string | null,
  playerToken?: string | null
): Promise<{ actions: Action[]; revision: number | null }> {
  const body: ValidActionsRequestDTO = { state, gameID };
  const response = await postJSON<ValidActionsResponseDTO>("/actions/valid", body, { playerToken });
  return {
    actions: response.actions,
    revision: response.revision ?? null
  };
}

export async function applyAction(
  state: GameState,
  action: Action,
  gameID?: string | null,
  clientRevision?: number | null,
  playerToken?: string | null
): Promise<{ state: GameState; effectResult: EffectResult | null; revision: number | null }> {
  const body: ApplyActionRequestDTO = { state, action, gameID, clientRevision };
  const response = await postJSON<ApplyActionResponseDTO>("/actions/apply", body, { playerToken });
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
  gameID?: string | null,
  playerToken?: string | null
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
  const response = await postJSON<ResolveBattleResponseDTO>("/battles/resolve", body, { playerToken });
  return {
    action: response.action,
    revision: response.revision ?? null
  };
}

export async function fetchBattleContext(
  state: GameState,
  action: Action,
  gameID?: string | null,
  playerToken?: string | null
): Promise<{ battleContext: BattleContext; revision: number | null }> {
  const body: BattleContextRequestDTO = { state, action, gameID };
  const response = await postJSON<BattleContextResponseDTO>("/battles/context", body, { playerToken });
  return {
    battleContext: response.battleContext,
    revision: response.revision ?? null
  };
}

export async function openBattle(
  state: GameState,
  action: Action,
  gameID: string,
  playerToken?: string | null
): Promise<{ prompt: BattlePrompt | null; revision: number | null }> {
  const body: BattleContextRequestDTO = { state, action, gameID };
  const response = await postJSON<BattlePromptResponseDTO>("/battles/open", body, { playerToken });
  return {
    prompt: response.prompt ?? null,
    revision: response.revision ?? null
  };
}

export async function fetchBattleSession(
  gameID: string,
  playerToken?: string | null
): Promise<{ prompt: BattlePrompt | null; revision: number | null }> {
  const response = await getJSON<BattlePromptResponseDTO>(`/battles/session?gameID=${encodeURIComponent(gameID)}`, { playerToken });
  return {
    prompt: response.prompt ?? null,
    revision: response.revision ?? null
  };
}

export async function submitBattleResponse(
  request: BattleResponseRequestDTO,
  playerToken?: string | null
): Promise<{ prompt: BattlePrompt | null; revision: number | null }> {
  const response = await postJSON<BattlePromptResponseDTO>("/battles/respond", request, { playerToken });
  return {
    prompt: response.prompt ?? null,
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
  gameID: string,
  playerToken?: string | null
): Promise<{ state: GameState; gameID: string | null; revision: number | null }> {
  const body: LoadGameRequestDTO = { gameID };
  const response = await postJSON<LoadGameResponseDTO>("/game/load", body, { playerToken });
  return {
    state: response.state,
    gameID: response.gameID ?? gameID,
    revision: response.revision ?? null
  };
}

export async function createLobby(request: CreateLobbyRequestDTO): Promise<{ lobby: Lobby; self: LobbyPlayer; playerToken: string }> {
  const response = await postJSON<CreateLobbyResponseDTO>("/lobby/create", request);
  return {
    ...lobbyResult(response),
    playerToken: response.playerToken
  };
}

export async function joinLobby(request: JoinLobbyRequestDTO): Promise<{ lobby: Lobby; self: LobbyPlayer; playerToken: string }> {
  const response = await postJSON<JoinLobbyResponseDTO>("/lobby/join", request);
  return {
    ...lobbyResult(response),
    playerToken: response.playerToken
  };
}

export async function fetchLobbyState(playerToken: string): Promise<{ lobby: Lobby; self: LobbyPlayer }> {
  const response = await getJSON<LobbyResponseDTO>("/lobby/state", { playerToken });
  return lobbyResult(response);
}

export async function claimLobbyFaction(playerToken: string, faction: number | null): Promise<{ lobby: Lobby; self: LobbyPlayer }> {
  const body: ClaimFactionRequestDTO = { faction };
  const response = await postJSON<LobbyResponseDTO>("/lobby/claim-faction", body, { playerToken });
  return lobbyResult(response);
}

export async function setLobbyReady(playerToken: string, isReady: boolean): Promise<{ lobby: Lobby; self: LobbyPlayer }> {
  const body: ReadyLobbyRequestDTO = { isReady };
  const response = await postJSON<LobbyResponseDTO>("/lobby/ready", body, { playerToken });
  return lobbyResult(response);
}

export async function startLobby(playerToken: string): Promise<{ lobby: Lobby; self: LobbyPlayer; state: GameState; gameID: string; revision: number | null }> {
  const response = await postJSON<StartLobbyResponseDTO>("/lobby/start", {}, { playerToken });
  return {
    ...lobbyResult(response),
    state: response.state,
    gameID: response.gameID,
    revision: response.revision ?? null
  };
}

export async function leaveLobby(playerToken: string): Promise<{ closed: boolean; lobby: Lobby | null; self: LobbyPlayer | null }> {
  const response = await postJSON<LeaveLobbyResponseDTO>("/lobby/leave", {}, { playerToken });
  return {
    closed: response.closed,
    lobby: response.lobby ?? null,
    self: response.self ?? null
  };
}
