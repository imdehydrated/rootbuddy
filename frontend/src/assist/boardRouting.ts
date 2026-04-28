import { ACTION_TYPE } from "../labels";
import type { Action, HighlightedClearing } from "../types";

export type MovementEndpoint = {
  fromClearingID: number;
  toClearingID: number;
  fromForestID: number;
  toForestID: number;
};

export type AssistMovementSource = {
  kind: "clearing" | "forest";
  id: number;
};

export type AssistActionCandidateRef = {
  actionIndex: number;
  action: Action;
};

export type AssistBattleCandidate = {
  actionIndex: number;
  action: Action;
};

export type AssistMovementCandidate = {
  actionIndex: number;
  action: Action;
  endpoints: MovementEndpoint;
};

export type AssistClearingCandidate = {
  actionIndex: number;
  action: Action;
  clearingIDs: number[];
};

export type AssistBoardCandidates = {
  battleCandidates: AssistBattleCandidate[];
  movementCandidates: AssistMovementCandidate[];
  buildRecruitCandidates: AssistClearingCandidate[];
  factionSpatialCandidates: AssistClearingCandidate[];
};

export function movementEndpoints(action: Action | null | undefined): MovementEndpoint | null {
  if (!action) {
    return null;
  }

  if (action.type === ACTION_TYPE.MOVEMENT && action.movement) {
    const hasSource = action.movement.from > 0 || action.movement.fromForestID > 0;
    const hasDestination = action.movement.to > 0 || action.movement.toForestID > 0;
    if (!hasSource || !hasDestination) {
      return null;
    }
    return {
      fromClearingID: action.movement.from,
      toClearingID: action.movement.to,
      fromForestID: action.movement.fromForestID,
      toForestID: action.movement.toForestID
    };
  }

  if (action.type === ACTION_TYPE.SLIP && action.slip) {
    const hasSource = action.slip.from > 0 || action.slip.fromForestID > 0;
    const hasDestination = action.slip.to > 0 || action.slip.toForestID > 0;
    if (!hasSource || !hasDestination) {
      return null;
    }
    return {
      fromClearingID: action.slip.from,
      toClearingID: action.slip.to,
      fromForestID: action.slip.fromForestID,
      toForestID: action.slip.toForestID
    };
  }

  return null;
}

export function movementSourceMatches(endpoint: MovementEndpoint, source: AssistMovementSource): boolean {
  return source.kind === "clearing" ? endpoint.fromClearingID === source.id : endpoint.fromForestID === source.id;
}

export function buildRecruitClearingIDs(action: Action | null | undefined): number[] {
  if (!action) {
    return [];
  }

  switch (action.type) {
    case ACTION_TYPE.BUILD:
      return action.build?.clearingID ? [action.build.clearingID] : [];
    case ACTION_TYPE.RECRUIT:
      return action.recruit?.clearingIDs ?? [];
    case ACTION_TYPE.OVERWORK:
      return action.overwork?.clearingID ? [action.overwork.clearingID] : [];
    case ACTION_TYPE.EYRIE_NEW_ROOST:
      return action.eyrieNewRoost?.clearingID ? [action.eyrieNewRoost.clearingID] : [];
    default:
      return [];
  }
}

export function factionSpatialClearingIDs(action: Action | null | undefined): number[] {
  if (!action) {
    return [];
  }

  switch (action.type) {
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return action.spreadSympathy?.clearingID ? [action.spreadSympathy.clearingID] : [];
    case ACTION_TYPE.REVOLT:
      return action.revolt?.clearingID ? [action.revolt.clearingID] : [];
    case ACTION_TYPE.ORGANIZE:
      return action.organize?.clearingID ? [action.organize.clearingID] : [];
    case ACTION_TYPE.EXPLORE:
      return action.explore?.clearingID ? [action.explore.clearingID] : [];
    case ACTION_TYPE.AID:
      return action.aid?.clearingID ? [action.aid.clearingID] : [];
    case ACTION_TYPE.STRIKE:
      return action.strike?.clearingID ? [action.strike.clearingID] : [];
    default:
      return [];
  }
}

export function buildAssistBoardCandidates(params: {
  battleCandidates: AssistActionCandidateRef[];
  movementCandidates: AssistActionCandidateRef[];
  buildRecruitCandidates: AssistActionCandidateRef[];
  factionSpatialCandidates: AssistActionCandidateRef[];
}): AssistBoardCandidates {
  return {
    battleCandidates: params.battleCandidates
      .map((candidate) => ({ actionIndex: candidate.actionIndex, action: candidate.action }))
      .filter((candidate): candidate is AssistBattleCandidate => candidate.action.type === ACTION_TYPE.BATTLE),
    movementCandidates: params.movementCandidates
      .map((candidate) => ({ actionIndex: candidate.actionIndex, action: candidate.action, endpoints: movementEndpoints(candidate.action) }))
      .filter((candidate): candidate is AssistMovementCandidate => candidate.endpoints !== null),
    buildRecruitCandidates: params.buildRecruitCandidates
      .map((candidate) => ({ actionIndex: candidate.actionIndex, action: candidate.action, clearingIDs: buildRecruitClearingIDs(candidate.action) }))
      .filter((candidate): candidate is AssistClearingCandidate => candidate.clearingIDs.length > 0),
    factionSpatialCandidates: params.factionSpatialCandidates
      .map((candidate) => ({ actionIndex: candidate.actionIndex, action: candidate.action, clearingIDs: factionSpatialClearingIDs(candidate.action) }))
      .filter((candidate): candidate is AssistClearingCandidate => candidate.clearingIDs.length > 0)
  };
}

export function assistBoardHighlights(params: AssistBoardCandidates & {
  movementSource: AssistMovementSource | null;
}): HighlightedClearing[] {
  const battleHighlights = Array.from(new Set(params.battleCandidates.map((candidate) => candidate.action.battle?.clearingID ?? 0)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));

  if (battleHighlights.length > 0) {
    return battleHighlights;
  }

  const movementHighlights =
    params.movementSource === null
      ? Array.from(new Set(params.movementCandidates.map((candidate) => candidate.endpoints.fromClearingID)))
          .filter((clearingID) => clearingID > 0)
          .map((clearingID) => ({ clearingID, role: "source" as const }))
      : Array.from(
          new Set(
            params.movementCandidates
              .filter((candidate) => movementSourceMatches(candidate.endpoints, params.movementSource as AssistMovementSource))
              .map((candidate) => candidate.endpoints.toClearingID)
          )
        )
          .filter((clearingID) => clearingID > 0)
          .map((clearingID) => ({ clearingID, role: "target" as const }));

  if (movementHighlights.length > 0) {
    return movementHighlights;
  }

  const buildRecruitHighlights = Array.from(new Set(params.buildRecruitCandidates.flatMap((candidate) => candidate.clearingIDs)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));

  if (buildRecruitHighlights.length > 0) {
    return buildRecruitHighlights;
  }

  return Array.from(new Set(params.factionSpatialCandidates.flatMap((candidate) => candidate.clearingIDs)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));
}
