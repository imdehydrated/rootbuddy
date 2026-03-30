import type {
  Action,
  BattleContext,
  BattleModifiers,
  EffectResult,
  GameState,
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

export type MultiplayerSocketMessage =
  | {
      type: "lobby.update";
      lobby: Record<string, unknown>;
    }
  | {
      type: "game.start";
      gameID: string;
      revision: number;
      state: GameState;
    }
  | {
      type: "game.state";
      gameID: string;
      revision: number;
      state: GameState;
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
