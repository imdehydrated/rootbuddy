import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";
import { ACTION_TYPE } from "./labels";
import { sampleState } from "./sampleState";
import type { Action, GameState, Lobby, LobbyPlayer } from "./types";

class FakeWebSocket {
  static instances: FakeWebSocket[] = [];

  readonly url: string;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent<string>) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  open() {
    this.onopen?.(new Event("open"));
  }

  emit(message: unknown) {
    this.onmessage?.({ data: JSON.stringify(message) } as MessageEvent<string>);
  }

  close() {
  }
}

function setupState(): GameState {
  const state = structuredClone(sampleState);
  state.gameMode = 1;
  state.gamePhase = 0;
  state.setupStage = 1;
  state.playerFaction = 0;
  state.factionTurn = 0;
  state.currentPhase = 0;
  state.currentStep = 0;
  state.marquise.keepClearingID = 0;
  state.map.clearings = state.map.clearings.map((clearing) => ({
    ...clearing,
    wood: 0,
    warriors: {},
    buildings: [],
    tokens: [],
    ruins: false,
    ruinItems: []
  }));
  return state;
}

function activeTurnState(overrides: Partial<GameState> = {}): GameState {
  const state = structuredClone(sampleState);
  state.gameMode = 1;
  state.gamePhase = 1;
  state.setupStage = 0;
  state.playerFaction = 2;
  state.factionTurn = 2;
  state.currentPhase = 1;
  state.currentStep = 3;
  Object.assign(state, overrides);
  return state;
}

function marquiseSetupAction(
  keepClearingID: number,
  sawmillClearingID: number,
  workshopClearingID: number,
  recruiterClearingID: number
): Action {
  return {
    type: ACTION_TYPE.MARQUISE_SETUP,
    marquiseSetup: {
      faction: 0,
      keepClearingID,
      sawmillClearingID,
      workshopClearingID,
      recruiterClearingID
    }
  };
}

function movementAction(from: number, to: number): Action {
  return {
    type: ACTION_TYPE.MOVEMENT,
    movement: {
      faction: 2,
      count: 1,
      maxCount: 1,
      from,
      to,
      fromForestID: 0,
      toForestID: 0,
      decreeCardID: 0,
      sourceEffectID: ""
    }
  };
}

function standAndDeliverAction(faction = 2, targetFaction = 0): Action {
  return {
    type: ACTION_TYPE.USE_PERSISTENT_EFFECT,
    usePersistentEffect: {
      faction,
      effectID: "stand_and_deliver",
      targetFaction,
      clearingID: 0,
      observedCardID: 0
    }
  };
}

function battleAction(faction = 2, targetFaction = 0, clearingID = 3): Action {
  return {
    type: ACTION_TYPE.BATTLE,
    battle: {
      faction,
      clearingID,
      targetFaction,
      decreeCardID: 0,
      sourceEffectID: ""
    }
  };
}

function lobbyPlayer(
  displayName: string,
  options: Partial<LobbyPlayer> = {}
): LobbyPlayer {
  return {
    displayName,
    faction: 0,
    hasFaction: true,
    isHost: false,
    isReady: false,
    connected: true,
    ...options
  };
}

function lobby(players: LobbyPlayer[], options: Partial<Lobby> = {}): Lobby {
  return {
    joinCode: "ROOT42",
    state: 0,
    players,
    factions: [0, 2],
    mapID: "autumn",
    vagabondCharacter: 0,
    eyrieLeader: 0,
    createdAt: "2026-04-07T00:00:00Z",
    ...options
  };
}

function mockSetupApi(state: GameState, actions: Action[]) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith("/api/game/setup")) {
      return new Response(JSON.stringify({ state, gameID: null, revision: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/actions/valid")) {
      return new Response(JSON.stringify({ actions, revision: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    return new Response(JSON.stringify({ error: `Unhandled test URL: ${url}` }), {
      status: 500,
      headers: { "Content-Type": "application/json" }
    });
  });

  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function mockActiveTurnApi(state: GameState, actions: Action[], appliedState: GameState = state, effectMessage = "Action applied.") {
  const appliedActions: Action[] = [];
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith("/api/game/setup")) {
      return new Response(JSON.stringify({ state, gameID: null, revision: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/actions/valid")) {
      return new Response(JSON.stringify({ actions, revision: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/actions/apply")) {
      const body = JSON.parse(String(init?.body ?? "{}")) as { action?: Action };
      if (body.action) {
        appliedActions.push(body.action);
      }
      return new Response(
        JSON.stringify({
          state: appliedState,
          effectResult: { effectID: "", message: effectMessage, cards: [] },
          revision: null
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" }
        }
      );
    }
    return new Response(JSON.stringify({ error: `Unhandled test URL: ${url}` }), {
      status: 500,
      headers: { "Content-Type": "application/json" }
    });
  });

  vi.stubGlobal("fetch", fetchMock);
  return { fetchMock, appliedActions };
}

function mockCreateLobbyApi(createdLobby: Lobby, self: LobbyPlayer, playerToken: string) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith("/api/lobby/create")) {
      return new Response(JSON.stringify({ lobby: createdLobby, self, playerToken }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    return new Response(JSON.stringify({ error: `Unhandled test URL: ${url}` }), {
      status: 500,
      headers: { "Content-Type": "application/json" }
    });
  });

  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
  window.localStorage.clear();
  FakeWebSocket.instances = [];
});

describe("App setup integration", () => {
  it("stages Marquise setup pieces immediately from board clearing clicks", async () => {
    const state = setupState();
    mockSetupApi(state, [
      marquiseSetupAction(1, 5, 10, 9),
      marquiseSetupAction(1, 5, 9, 10),
      marquiseSetupAction(1, 10, 5, 9),
      marquiseSetupAction(2, 6, 10, 5)
    ]);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    await waitFor(() => expect(screen.getAllByText("Marquise Setup").length).toBeGreaterThan(0));
    expect(await screen.findByText("Choose the Keep corner")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Clearing 1" }));

    expect(await screen.findByLabelText("Pending Keep")).toBeInTheDocument();
    expect(screen.getByText("Place the sawmill")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Clearing 5" }));

    expect(await screen.findByLabelText("Pending sawmill")).toBeInTheDocument();
    expect(screen.getByText("Place the workshop")).toBeInTheDocument();

    await waitFor(() => expect(screen.queryByText(/Unhandled test URL/i)).not.toBeInTheDocument());
  });
});

describe("App board action integration", () => {
  it("applies an active movement action from board source and destination clicks", async () => {
    const state = activeTurnState();
    const firstMove = movementAction(3, 7);
    const secondMove = movementAction(3, 8);
    const { appliedActions } = mockActiveTurnApi(state, [firstMove, secondMove], state, "Move applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();
    expect((await screen.findAllByText("Daylight / Daylight Actions")).length).toBeGreaterThan(0);

    const loadActionButtons = screen.getAllByRole("button", { name: /Load Actions/i });
    fireEvent.click(loadActionButtons[loadActionButtons.length - 1]);
    fireEvent.click(await screen.findByRole("button", { name: /Move.*Pieces changed clearings/i }));

    fireEvent.click(screen.getByRole("button", { name: "Clearing 3" }));
    expect(await screen.findByText(/Move source selected: clearing 3/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Clearing 7" }));

    await waitFor(() => expect(appliedActions).toEqual([firstMove]));
    expect(await screen.findByText("Move applied.")).toBeInTheDocument();
  });

  it("captures known Stand and Deliver stolen cards through the in-app modal before applying", async () => {
    const state = activeTurnState({ playerFaction: 0, factionTurn: 0 });
    state.otherHandCounts = { ...state.otherHandCounts, "2": 1 };
    const action = standAndDeliverAction(0, 2);
    const { appliedActions } = mockActiveTurnApi(state, [action], state, "Stand and Deliver applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();
    expect((await screen.findAllByText("Daylight / Daylight Actions")).length).toBeGreaterThan(0);

    const loadActionButtons = screen.getAllByRole("button", { name: /Load Actions/i });
    fireEvent.click(loadActionButtons[loadActionButtons.length - 1]);
    fireEvent.click(await screen.findByRole("button", { name: /Card Effect.*persistent effect/i }));
    const directStandAndDeliverButton = screen.queryByRole("button", { name: /Use Stand and Deliver! on Eyrie/i });
    fireEvent.click(directStandAndDeliverButton ?? screen.getByRole("button", { name: "Apply" }));

    expect(appliedActions).toEqual([]);
    expect(await screen.findByRole("heading", { name: "Stand and Deliver" })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Observed Card ID"), { target: { value: "41" } });

    expect(await screen.findByText(/Known card: Stand and Deliver! \(Fox\)/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Apply Stand and Deliver/i }));

    await waitFor(() =>
      expect(appliedActions).toEqual([
        {
          ...action,
          usePersistentEffect: {
            ...action.usePersistentEffect!,
            observedCardID: 41
          }
        }
      ])
    );
    expect(await screen.findByText("Stand and Deliver applied.")).toBeInTheDocument();
  });

  it("refuses Stand and Deliver when the target has no recorded cards", async () => {
    const state = activeTurnState({ playerFaction: 0, factionTurn: 0 });
    state.otherHandCounts = { ...state.otherHandCounts, "2": 0 };
    const action = standAndDeliverAction(0, 2);
    const { appliedActions } = mockActiveTurnApi(state, [action], state, "Stand and Deliver applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();
    expect((await screen.findAllByText("Daylight / Daylight Actions")).length).toBeGreaterThan(0);

    const loadActionButtons = screen.getAllByRole("button", { name: /Load Actions/i });
    fireEvent.click(loadActionButtons[loadActionButtons.length - 1]);
    fireEvent.click(await screen.findByRole("button", { name: /Card Effect.*persistent effect/i }));
    const directStandAndDeliverButton = screen.queryByRole("button", { name: /Use Stand and Deliver! on Eyrie/i });
    fireEvent.click(directStandAndDeliverButton ?? screen.getByRole("button", { name: "Apply" }));

    await waitFor(() => expect(appliedActions).toEqual([]));
    expect(screen.queryByRole("heading", { name: "Stand and Deliver" })).not.toBeInTheDocument();
    expect(await screen.findByText("Stand and Deliver cannot target a faction with no recorded cards.")).toBeInTheDocument();
  });
});

describe("App multiplayer integration", () => {
  it("creates a lobby and applies realtime lobby updates", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const updatedLobby = lobby([
      self,
      lobbyPlayer("Bob", { faction: 2, hasFaction: true, isReady: true })
    ]);
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    expect(screen.getByText("Lobby ROOT42 created.")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
    expect(FakeWebSocket.instances[0].url).toContain("/api/ws?token=token-1");

    act(() => {
      FakeWebSocket.instances[0].open();
    });
    expect(await screen.findByText("Live")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].emit({ type: "lobby.update", lobby: updatedLobby, self });
    });

    expect(await screen.findByText("2 / 2 seats filled")).toBeInTheDocument();
    expect(screen.getAllByText("Bob").length).toBeGreaterThan(0);
    const savedMultiplayer = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
      playerToken?: string;
      joinCode?: string;
    };
    expect(savedMultiplayer.playerToken).toBe("token-1");
    expect(savedMultiplayer.joinCode).toBe("ROOT42");
  });

  it("moves from lobby to board workspace on websocket game start", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    startedState.otherHandCounts = { ...startedState.otherHandCounts, "2": 3 };
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "game.start",
        gameID: "game-123",
        revision: 7,
        state: startedState
      });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();
    expect(screen.queryByText("Join Code ROOT42")).not.toBeInTheDocument();
    await waitFor(() => {
      const savedMultiplayer = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
        gameID?: string;
      };
      expect(savedMultiplayer.gameID).toBe("game-123");
    });
  });

  it("shows a conflict notice and keeps server-pushed state when websocket conflict arrives", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    const conflictState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2, roundNumber: startedState.roundNumber + 1 });
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "game.start",
        gameID: "game-123",
        revision: 7,
        state: startedState
      });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "conflict",
        gameID: "game-123",
        revision: 8,
        state: conflictState,
        error: "Server revision won."
      });
    });

    expect(await screen.findByText("Server State Updated")).toBeInTheDocument();
    expect(screen.getAllByText("Server revision won.").length).toBeGreaterThan(0);
    const savedMultiplayer = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
      gameID?: string;
    };
    expect(savedMultiplayer.gameID).toBe("game-123");
  });

  it("shows a session error notice when websocket session error arrives during play", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "game.start",
        gameID: "game-123",
        revision: 7,
        state: startedState
      });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "session.error",
        gameID: "game-123",
        error: "Session expired."
      });
    });

    expect(await screen.findByText("Session Error")).toBeInTheDocument();
    expect(screen.getAllByText("Session expired.").length).toBeGreaterThan(0);
    expect(screen.getByAltText("Autumn board")).toBeInTheDocument();
  });

  it("shows a local-player battle prompt from websocket battle prompt messages", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    const action = battleAction(2, 0, 3);
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "game.start",
        gameID: "game-123",
        revision: 7,
        state: startedState
      });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "battle.prompt",
        prompt: {
          gameID: "game-123",
          revision: 8,
          action,
          stage: "defender_response",
          waitingOnFaction: 0,
          battleContext: {
            action,
            clearingSuit: 1,
            timing: [],
            attackerHasScoutingParty: false,
            canDefenderAmbush: true,
            canAttackerCounterAmbush: false,
            canAttackerArmorers: false,
            canDefenderArmorers: false,
            canAttackerBrutalTactics: false,
            canDefenderSappers: false,
            assistDefenderAmbushPromptRequired: false
          },
          canUseAmbush: true
        }
      });
    });

    expect(await screen.findByText("Defender response required before battle resolution.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Submit Response/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Resolve and Apply/i })).not.toBeInTheDocument();
  });
});
