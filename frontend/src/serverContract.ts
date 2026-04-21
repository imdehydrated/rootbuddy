import type {
  Action,
  ActionLogEntry,
  BattlePrompt,
  BattleContext,
  BattleModifiers,
  EffectResult,
  GameState,
  Lobby,
  LobbyPlayer,
  SetupRequest
} from "./types";

export type ServerErrorResponse = {
  error: string;
  gameID?: string;
  revision?: number;
  state?: GameState;
};

export type ValidActionsRequestDTO = {
  state: GameState;
  gameID?: string | null;
};

export type ValidActionsResponseDTO = {
  actions: Action[];
  gameID?: string;
  revision?: number;
};

export type ApplyActionRequestDTO = {
  state: GameState;
  action: Action;
  gameID?: string | null;
  clientRevision?: number | null;
};

export type ApplyActionResponseDTO = {
  state: GameState;
  effectResult?: EffectResult;
  gameID?: string;
  revision?: number;
};

export type ResolveBattleRequestDTO = {
  state: GameState;
  action: Action;
  attackerRoll: number;
  defenderRoll: number;
  modifiers?: BattleModifiers;
  useModifiers: boolean;
  gameID?: string | null;
};

export type ResolveBattleResponseDTO = {
  action: Action;
  gameID?: string;
  revision?: number;
};

export type BattleContextRequestDTO = {
  state: GameState;
  action: Action;
  gameID?: string | null;
};

export type BattleContextResponseDTO = {
  battleContext: BattleContext;
  gameID?: string;
  revision?: number;
};

export type BattlePromptResponseDTO = {
  prompt?: BattlePrompt;
  gameID?: string;
  revision?: number;
};

export type BattleResponseRequestDTO = {
  gameID: string;
  useAmbush?: boolean;
  useDefenderArmorers?: boolean;
  useSappers?: boolean;
  useCounterAmbush?: boolean;
  useAttackerArmorers?: boolean;
  useBrutalTactics?: boolean;
};

export type SetupRequestDTO = SetupRequest;

export type SetupResponseDTO = {
  state: GameState;
  gameID?: string;
  revision?: number;
};

export type LoadGameRequestDTO = {
  gameID: string;
};

export type LoadGameResponseDTO = {
  state: GameState;
  gameID?: string;
  revision?: number;
};

export type GameLogResponseDTO = {
  entries: ActionLogEntry[];
  gameID?: string;
  revision?: number;
};

export type CreateLobbyRequestDTO = {
  displayName: string;
  factions?: number[];
  mapID?: string;
};

export type JoinLobbyRequestDTO = {
  joinCode: string;
  displayName: string;
};

export type ClaimFactionRequestDTO = {
  faction: number | null;
};

export type ReadyLobbyRequestDTO = {
  isReady: boolean;
};

export type LobbyResponseDTO = {
  lobby: Lobby;
  self: LobbyPlayer;
};

export type CreateLobbyResponseDTO = LobbyResponseDTO & {
  playerToken: string;
};

export type JoinLobbyResponseDTO = LobbyResponseDTO & {
  playerToken: string;
};

export type StartLobbyResponseDTO = LobbyResponseDTO & {
  state: GameState;
  gameID: string;
  revision?: number;
};

export type LeaveLobbyResponseDTO = {
  closed: boolean;
  lobby?: Lobby;
  self?: LobbyPlayer;
};

export type MultiplayerSocketMessage =
  | {
      type: "lobby.update";
      lobby: Lobby;
      self: LobbyPlayer;
    }
  | {
      type: "game.start";
      gameID: string;
      revision: number;
      state: GameState;
      actionLog?: ActionLogEntry[];
    }
  | {
      type: "game.state";
      gameID: string;
      revision: number;
      state: GameState;
      actionLog?: ActionLogEntry[];
    }
  | {
      type: "battle.prompt";
      prompt?: BattlePrompt;
    }
  | {
      type: "conflict";
      gameID: string;
      revision: number;
      state: GameState;
      error: string;
    }
  | {
      type: "session.error";
      error: string;
      gameID?: string;
    };
