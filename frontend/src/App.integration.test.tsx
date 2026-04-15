import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";
import { ACTION_TYPE } from "./labels";
import { sampleState } from "./sampleState";
import type { Action, BattlePrompt, GameState, Lobby, LobbyPlayer } from "./types";

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

  fail() {
    this.onerror?.(new Event("error"));
  }

  closeFromServer() {
    this.onclose?.({ code: 1006 } as CloseEvent);
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

function movementAction(from: number, to: number, faction = 2): Action {
  return {
    type: ACTION_TYPE.MOVEMENT,
    movement: {
      faction,
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

function buildAction(clearingID: number, faction = 2): Action {
  return {
    type: ACTION_TYPE.BUILD,
    build: {
      faction,
      clearingID,
      buildingType: 3,
      woodSources: [],
      decreeCardID: 0
    }
  };
}

function organizeAction(clearingID: number, faction = 1): Action {
  return {
    type: ACTION_TYPE.ORGANIZE,
    organize: {
      faction,
      clearingID
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

function mockCreateLobbyApi(
  createdLobby: Lobby,
  self: LobbyPlayer,
  playerToken: string,
  options: {
    battleResponse?: { prompt?: BattlePrompt; revision?: number };
    battleResolve?: { action: Action; revision?: number | null };
    applyResult?: { state: GameState; effectResult?: { effectID: string; message: string; cards: number[] }; revision?: number | null };
    validActions?: Action[];
    validRevision?: number | null;
  } = {}
) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith("/api/lobby/create")) {
      return new Response(JSON.stringify({ lobby: createdLobby, self, playerToken }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/lobby/state")) {
      return new Response(JSON.stringify({ lobby: createdLobby, self }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/game/load") && options.applyResult) {
      return new Response(
        JSON.stringify({
          state: options.applyResult.state,
          gameID: createdLobby.gameID ?? "game-123",
          revision: options.applyResult.revision ?? null
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" }
        }
      );
    }
    if (url.endsWith("/api/actions/valid")) {
      return new Response(JSON.stringify({ actions: options.validActions ?? [], revision: options.validRevision ?? null }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.includes("/api/battles/resolve") && options.battleResolve) {
      return new Response(JSON.stringify(options.battleResolve), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.endsWith("/api/actions/apply") && options.applyResult) {
      return new Response(JSON.stringify(options.applyResult), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      });
    }
    if (url.includes("/api/battles/respond")) {
      return new Response(JSON.stringify(options.battleResponse ?? { prompt: undefined, revision: null }), {
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
  return { fetchMock };
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
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

  it("updates Marquise setup legal highlights for each staged placement", async () => {
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

    const clearing1 = screen.getByRole("button", { name: "Clearing 1" });
    const clearing2 = screen.getByRole("button", { name: "Clearing 2" });
    const clearing5 = screen.getByRole("button", { name: "Clearing 5" });
    const clearing6 = screen.getByRole("button", { name: "Clearing 6" });
    const clearing9 = screen.getByRole("button", { name: "Clearing 9" });
    const clearing10 = screen.getByRole("button", { name: "Clearing 10" });

    await waitFor(() => {
      expect(clearing1).toHaveClass("setup-legal");
      expect(clearing2).toHaveClass("setup-legal");
      expect(clearing5).not.toHaveClass("setup-legal");
    });

    fireEvent.click(clearing1);

    await waitFor(() => {
      expect(clearing5).toHaveClass("setup-legal");
      expect(clearing10).toHaveClass("setup-legal");
      expect(clearing6).not.toHaveClass("setup-legal");
    });

    fireEvent.click(clearing5);

    await waitFor(() => {
      expect(clearing9).toHaveClass("setup-legal");
      expect(clearing10).toHaveClass("setup-legal");
      expect(clearing6).not.toHaveClass("setup-legal");
    });

    fireEvent.click(clearing9);

    await waitFor(() => {
      expect(clearing10).toHaveClass("setup-legal");
      expect(clearing5).not.toHaveClass("setup-legal");
      expect(clearing6).not.toHaveClass("setup-legal");
    });
  });

  it("keeps a chosen Marquise setup clearing highlighted when it remains legal for the next placement", async () => {
    const state = setupState();
    mockSetupApi(state, [
      marquiseSetupAction(1, 5, 5, 9),
      marquiseSetupAction(1, 5, 9, 5),
      marquiseSetupAction(1, 9, 5, 5)
    ]);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    await waitFor(() => expect(screen.getAllByText("Marquise Setup").length).toBeGreaterThan(0));

    fireEvent.click(screen.getByRole("button", { name: "Clearing 1" }));
    fireEvent.click(screen.getByRole("button", { name: "Clearing 5" }));

    expect(await screen.findByText("Place the workshop")).toBeInTheDocument();

    const reusedClearing = screen.getByRole("button", { name: "Clearing 5" });
    const alternateClearing = screen.getByRole("button", { name: "Clearing 9" });

    await waitFor(() => expect(reusedClearing).toHaveClass("setup-chosen", "setup-legal"));
    await waitFor(() => expect(alternateClearing).toHaveClass("setup-legal"));
  });
});

describe("App board action integration", () => {
  it("keeps routine live play board-first and only reveals correction tools through correction mode", async () => {
    const state = activeTurnState();
    mockActiveTurnApi(state, [movementAction(3, 7)], state, "Move applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();
    expect(screen.queryByText("Primary Workflow")).not.toBeInTheDocument();
    expect(screen.queryByText("Flow Guide")).not.toBeInTheDocument();
    expect(screen.queryByText("Context & Reference")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Correction Mode" }));

    expect(await screen.findByRole("heading", { name: "Correction Mode" })).toBeInTheDocument();
    expect(screen.getByText("Context & Reference")).toBeInTheDocument();
  });

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
    expect(await screen.findByText("Choose what you want to do next.")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /Move.*Pieces changed clearings/i }));
    expect(await screen.findByText("Choose a highlighted source clearing, then choose the destination.")).toBeInTheDocument();

    const sourceClearing = screen.getByRole("button", { name: "Clearing 3" });
    await waitFor(() => expect(sourceClearing).toHaveClass("highlight-source"));
    fireEvent.click(sourceClearing);

    const destinationClearing = screen.getByRole("button", { name: "Clearing 7" });
    await waitFor(() => expect(destinationClearing).toHaveClass("highlight-target"));
    fireEvent.click(destinationClearing);

    await waitFor(() => expect(appliedActions).toEqual([firstMove]));
    expect(await screen.findByText("Move applied.")).toBeInTheDocument();
  });

  it("applies an active build candidate from a board clearing click", async () => {
    const state = activeTurnState();
    const action = buildAction(7);
    const { appliedActions } = mockActiveTurnApi(state, [action], state, "Build applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /Build \/ Recruit.*Pieces, wood, or buildings/i }));
    expect(await screen.findByText("Choose the clearing where this step happens.")).toBeInTheDocument();

    const buildClearing = screen.getByRole("button", { name: "Clearing 7" });
    fireEvent.click(buildClearing);

    await waitFor(() => expect(appliedActions).toEqual([action]));
    expect(await screen.findByText("Build applied.")).toBeInTheDocument();
  });

  it("applies an active clearing-based faction action from a board clearing click", async () => {
    const state = activeTurnState({ playerFaction: 1, factionTurn: 1 });
    const action = organizeAction(9);
    const { appliedActions } = mockActiveTurnApi(state, [action], state, "Organize applied.");

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));
    fireEvent.click(screen.getByRole("button", { name: /Start Assist Game/i }));

    expect(await screen.findByText("Game created.")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /Faction Action.*faction-specific public step/i }));

    const factionActionClearing = screen.getByRole("button", { name: "Clearing 9" });
    await waitFor(() => expect(factionActionClearing).toHaveClass("highlight-affected"));
    fireEvent.click(factionActionClearing);

    await waitFor(() => expect(appliedActions).toEqual([action]));
    expect(await screen.findByText("Organize applied.")).toBeInTheDocument();
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
    await waitFor(() => expect(FakeWebSocket.instances.length).toBeGreaterThanOrEqual(1));
    expect(FakeWebSocket.instances.some((socket) => socket.url.includes("/api/ws?token=token-1"))).toBe(true);

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

  it("keeps the same websocket client alive across lobby.update messages", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: false, faction: 0, hasFaction: true });
    const createdLobby = lobby([self]);
    const updatedSelf = { ...self, isReady: true, faction: 2 };
    const updatedLobby = lobby([updatedSelf], { factions: [2] });
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
      FakeWebSocket.instances[0].open();
    });
    expect(await screen.findByText("Live")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].emit({ type: "lobby.update", lobby: updatedLobby, self: updatedSelf });
    });

    expect(await screen.findByText("Ready")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));
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
    expect(await screen.findByText("No multiplayer actions have been logged yet.")).toBeInTheDocument();
    expect(screen.queryByText("Join Code ROOT42")).not.toBeInTheDocument();
    await waitFor(() => {
      const savedMultiplayer = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
        gameID?: string;
      };
      expect(savedMultiplayer.gameID).toBe("game-123");
    });
  });

  it("loads Marquise setup targets when multiplayer game start enters setup", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true, faction: 0, hasFaction: true });
    const createdLobby = lobby([self]);
    const startedState = setupState();
    mockCreateLobbyApi(createdLobby, self, "token-1", {
      validActions: [
        marquiseSetupAction(1, 5, 10, 9),
        marquiseSetupAction(2, 6, 10, 5),
        marquiseSetupAction(3, 6, 9, 5),
        marquiseSetupAction(4, 9, 5, 10)
      ]
    });
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

    expect(await screen.findByText("Choose the Keep corner")).toBeInTheDocument();
    expect(await screen.findByText("4 legal choices")).toBeInTheDocument();
  });

  it("resumes a saved multiplayer game into the board workspace", async () => {
    const self = lobbyPlayer("Alice", { faction: 0, isHost: true, isReady: true });
    const resumedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    const savedSession = {
      playerToken: "token-1",
      displayName: "Alice",
      joinCode: "ROOT42",
      gameID: "game-123",
      savedAt: "2026-04-07T00:00:00Z"
    };
    const createdLobby = lobby([self], { state: 1, gameID: "game-123" });
    const { fetchMock } = mockCreateLobbyApi(createdLobby, self, "token-1", {
      applyResult: {
        state: resumedState,
        revision: 12
      }
    });
    window.localStorage.setItem("rootbuddy_multiplayer_session_v1", JSON.stringify(savedSession));
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();
    expect(await screen.findByText("Rejoined multiplayer game.")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/lobby/state"), expect.anything()));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/game/load"), expect.anything()));
    await waitFor(() => expect(FakeWebSocket.instances.length).toBeGreaterThanOrEqual(1));
    expect(FakeWebSocket.instances.some((socket) => socket.url.includes("/api/ws?token=token-1"))).toBe(true);
    const persistedSession = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
      gameID?: string;
      joinCode?: string;
      playerToken?: string;
    };
    expect(persistedSession).toMatchObject({
      gameID: "game-123",
      joinCode: "ROOT42",
      playerToken: "token-1"
    });
  });

  it("resumes a saved multiplayer lobby without a game into the lobby screen", async () => {
    const self = lobbyPlayer("Alice", { faction: 0, isHost: true, isReady: true });
    const savedSession = {
      playerToken: "token-1",
      displayName: "Alice",
      joinCode: "ROOT42",
      savedAt: "2026-04-08T00:00:00Z"
    };
    const createdLobby = lobby([self], { state: 0 });
    const { fetchMock } = mockCreateLobbyApi(createdLobby, self, "token-1");
    window.localStorage.setItem("rootbuddy_multiplayer_session_v1", JSON.stringify(savedSession));
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    expect(await screen.findByText("Rejoined lobby ROOT42.")).toBeInTheDocument();
    expect(screen.queryByAltText("Autumn board")).not.toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/lobby/state"), expect.anything()));
    expect(fetchMock).not.toHaveBeenCalledWith(expect.stringContaining("/api/game/load"), expect.anything());
    await waitFor(() => expect(FakeWebSocket.instances.length).toBeGreaterThanOrEqual(1));
    expect(FakeWebSocket.instances.some((socket) => socket.url.includes("/api/ws?token=token-1"))).toBe(true);
    const persistedSession = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
      gameID?: string;
      joinCode?: string;
      playerToken?: string;
    };
    expect(persistedSession).toMatchObject({
      joinCode: "ROOT42",
      playerToken: "token-1"
    });
    expect(persistedSession.gameID ?? null).toBeNull();
  });

  it("shows reconnecting state after a websocket drop and reconnects with the same multiplayer token", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    mockCreateLobbyApi(createdLobby, self, "token-1");
    vi.stubGlobal("WebSocket", FakeWebSocket);

    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Create Lobby" }));
    fireEvent.change(await screen.findByPlaceholderText("Enter your name"), { target: { value: "Alice" } });
    fireEvent.click(screen.getByRole("button", { name: "Create Lobby" }));

    expect(await screen.findByText("Join Code ROOT42")).toBeInTheDocument();
    await waitFor(() => expect(FakeWebSocket.instances.length).toBeGreaterThanOrEqual(1));

    act(() => {
      FakeWebSocket.instances[0].open();
    });
    expect(await screen.findByText("Live")).toBeInTheDocument();

    act(() => {
      FakeWebSocket.instances[0].closeFromServer();
    });

    expect(await screen.findByText("Reconnecting")).toBeInTheDocument();
    expect(await screen.findByText("Realtime connection lost. Reconnecting...")).toBeInTheDocument();

    await waitFor(() => expect(FakeWebSocket.instances.length).toBeGreaterThanOrEqual(2), { timeout: 2500 });
    expect(FakeWebSocket.instances[1].url).toContain("/api/ws?token=token-1");

    act(() => {
      FakeWebSocket.instances[1].open();
    });

    expect(await screen.findByText("Live")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Realtime connection lost. Reconnecting...")).not.toBeInTheDocument());
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

  it("applies websocket game state updates and refreshes legal actions when priority reaches the local player", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2, roundNumber: 1 });
    const pushedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 0, roundNumber: 2 });
    const legalMove = movementAction(1, 5, 0);
    const { fetchMock } = mockCreateLobbyApi(createdLobby, self, "token-1", {
      validActions: [legalMove],
      validRevision: 9
    });
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
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({
        type: "game.state",
        gameID: "game-123",
        revision: 8,
        state: pushedState
      });
    });

    expect(await screen.findByText("Choose your next move.")).toBeInTheDocument();
    expect(await screen.findByText("Choose what you want to do next.")).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/actions/valid"), expect.anything()));
    const savedMultiplayer = JSON.parse(window.localStorage.getItem("rootbuddy_multiplayer_session_v1") ?? "{}") as {
      gameID?: string;
    };
    expect(savedMultiplayer.gameID).toBe("game-123");
  });

  it("keeps the board visible and updates player presence on lobby.update during gameplay", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true, faction: 0, hasFaction: true });
    const bob = lobbyPlayer("Bob", { isHost: false, isReady: true, faction: 2, hasFaction: true });
    const createdLobby = lobby([self, bob], { state: 0 });
    const updatedLobby = lobby(
      [
        self,
        {
          ...bob,
          connected: false
        }
      ],
      { state: 1, gameID: "game-123" }
    );
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
    expect(await screen.findByLabelText("Player presence bar")).toBeInTheDocument();
    expect(screen.getAllByText("Bob").length).toBeGreaterThan(0);

    act(() => {
      FakeWebSocket.instances[0].emit({ type: "lobby.update", lobby: updatedLobby, self });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();
    expect(screen.queryByText("Join Code ROOT42")).not.toBeInTheDocument();
    expect(await screen.findByText("Away")).toBeInTheDocument();
  });

  it("renders websocket action log entries on the board during multiplayer play", async () => {
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
        state: startedState,
        actionLog: [
          {
            roundNumber: 1,
            faction: 2,
            actionType: ACTION_TYPE.MOVEMENT,
            summary: "Eyrie moved from clearing 3 to clearing 7.",
            timestamp: 1_700_000_000_000
          }
        ]
      });
    });

    expect(await screen.findByAltText("Autumn board")).toBeInTheDocument();
    expect(await screen.findByText("Eyrie moved from clearing 3 to clearing 7.")).toBeInTheDocument();
    expect(screen.getByText("Round 1")).toBeInTheDocument();
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
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

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

  it("submits a multiplayer battle response and renders the ready-to-resolve prompt", async () => {
    const self = lobbyPlayer("Alice", { isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 0, factionTurn: 2 });
    const action = battleAction(2, 0, 3);
    const battleContext = {
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
    };
    const initialPrompt: BattlePrompt = {
      gameID: "game-123",
      revision: 8,
      action,
      stage: "defender_response",
      waitingOnFaction: 0,
      battleContext,
      canUseAmbush: true
    };
    const readyPrompt: BattlePrompt = {
      ...initialPrompt,
      revision: 9,
      stage: "ready_to_resolve",
      waitingOnFaction: 2,
      defenderAmbush: false
    };
    const { fetchMock } = mockCreateLobbyApi(createdLobby, self, "token-1", {
      battleResponse: { prompt: readyPrompt, revision: 9 }
    });
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
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({ type: "battle.prompt", prompt: initialPrompt });
    });

    const submitResponseButton = await screen.findByRole("button", { name: /Submit Response/i });
    await waitFor(() => expect(submitResponseButton).not.toBeDisabled());
    act(() => {
      fireEvent.click(submitResponseButton);
    });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/battles/respond"), expect.anything()));
    const responseCall = fetchMock.mock.calls.find(([input]) => String(input).includes("/api/battles/respond"));
    expect(responseCall).toBeDefined();
    expect(JSON.parse(String((responseCall?.[1] as RequestInit | undefined)?.body ?? "{}"))).toEqual({
      gameID: "game-123",
      useAmbush: false,
      useDefenderArmorers: false,
      useSappers: false
    });
    expect(await screen.findByText("Responses are complete. Waiting on Eyrie to resolve.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Submit Response/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Resolve and Apply/i })).not.toBeInTheDocument();
  });

  it("resolves a multiplayer ready battle prompt and applies the server battle result", async () => {
    const self = lobbyPlayer("Alice", { faction: 2, isHost: true, isReady: true });
    const createdLobby = lobby([self]);
    const startedState = activeTurnState({ gameMode: 0, playerFaction: 2, factionTurn: 2 });
    const resolvedState = activeTurnState({ gameMode: 0, playerFaction: 2, factionTurn: 0, roundNumber: startedState.roundNumber + 1 });
    const action = battleAction(2, 0, 3);
    const prompt: BattlePrompt = {
      gameID: "game-123",
      revision: 9,
      action,
      stage: "ready_to_resolve",
      waitingOnFaction: 2,
      battleContext: {
        action,
        clearingSuit: 1,
        timing: [],
        attackerHasScoutingParty: false,
        canDefenderAmbush: false,
        canAttackerCounterAmbush: false,
        canAttackerArmorers: false,
        canDefenderArmorers: false,
        canAttackerBrutalTactics: false,
        canDefenderSappers: false,
        assistDefenderAmbushPromptRequired: false
      },
      defenderAmbush: false
    };
    const { fetchMock } = mockCreateLobbyApi(createdLobby, self, "token-1", {
      battleResolve: { action, revision: 10 },
      applyResult: {
        state: resolvedState,
        effectResult: { effectID: "", message: "Battle resolved on the server.", cards: [] },
        revision: 11
      }
    });
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
    await waitFor(() => expect(FakeWebSocket.instances).toHaveLength(1));

    act(() => {
      FakeWebSocket.instances[0].emit({ type: "battle.prompt", prompt });
    });

    const resolveButton = await screen.findByRole("button", { name: /Resolve and Apply/i });
    fireEvent.click(resolveButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/battles/resolve"), expect.anything()));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/actions/apply"), expect.anything()));
    const resolveCall = fetchMock.mock.calls.find(([input]) => String(input).includes("/api/battles/resolve"));
    const applyCall = fetchMock.mock.calls.find(([input]) => String(input).includes("/api/actions/apply"));
    expect(JSON.parse(String((resolveCall?.[1] as RequestInit | undefined)?.body ?? "{}"))).toMatchObject({
      gameID: "game-123",
      attackerRoll: 0,
      defenderRoll: 0
    });
    expect(JSON.parse(String((applyCall?.[1] as RequestInit | undefined)?.body ?? "{}"))).toMatchObject({
      gameID: "game-123",
      clientRevision: 10,
      action
    });
    await waitFor(() =>
      expect(screen.queryByText("Responses are complete. The server will roll battle dice during resolution.")).not.toBeInTheDocument()
    );
    expect(screen.queryByRole("button", { name: /Resolve and Apply/i })).not.toBeInTheDocument();
  });
});
