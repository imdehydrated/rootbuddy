import { useCallback, useEffect, useState } from "react";
import {
  assistBoardHighlights,
  buildAssistBoardCandidates,
  movementSourceMatches,
  type AssistActionCandidateRef,
  type AssistMovementSource
} from "../assistDirector";
import { actionHeadline } from "../actionPresentation";
import type { ClearingPreviewPiece } from "../components/ClearingMarker";
import { affectedClearings } from "../gameHelpers";
import { ACTION_TYPE } from "../labels";
import type { Action, GameState } from "../types";

export type MarquiseSetupDraft = {
  keepClearingID: number | null;
  sawmillClearingID: number | null;
  workshopClearingID: number | null;
  recruiterClearingID: number | null;
};

export const emptyMarquiseSetupDraft: MarquiseSetupDraft = {
  keepClearingID: null,
  sawmillClearingID: null,
  workshopClearingID: null,
  recruiterClearingID: null
};

function marquiseSetupMatches(action: Action, draft: MarquiseSetupDraft): boolean {
  const payload = action.marquiseSetup;
  if (!payload) {
    return false;
  }
  if (draft.keepClearingID !== null && payload.keepClearingID !== draft.keepClearingID) {
    return false;
  }
  if (draft.sawmillClearingID !== null && payload.sawmillClearingID !== draft.sawmillClearingID) {
    return false;
  }
  if (draft.workshopClearingID !== null && payload.workshopClearingID !== draft.workshopClearingID) {
    return false;
  }
  if (draft.recruiterClearingID !== undefined && draft.recruiterClearingID !== null && payload.recruiterClearingID !== draft.recruiterClearingID) {
    return false;
  }
  return true;
}

function sameCandidateRefs(current: AssistActionCandidateRef[], next: AssistActionCandidateRef[]) {
  return (
    current.length === next.length &&
    current.every((candidate, index) => candidate.actionIndex === next[index].actionIndex && candidate.action === next[index].action)
  );
}

export function setupBoardPrompt(stage: number, draft: MarquiseSetupDraft): { title: string; instruction: string; detail: string } {
  switch (stage) {
    case 1:
      if (draft.keepClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Choose the Keep corner",
          detail: "Click one of the highlighted corner clearings."
        };
      }
      if (draft.sawmillClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Place the sawmill",
          detail: "Click a highlighted legal clearing. The pending building appears immediately."
        };
      }
      if (draft.workshopClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Place the workshop",
          detail: "Click a highlighted legal clearing. The pending building appears immediately."
        };
      }
      return {
        title: "Marquise Setup",
        instruction: "Place the recruiter",
        detail: "Click a highlighted legal clearing to finish Marquise setup."
      };
    case 2:
      return {
        title: "Eyrie Setup",
        instruction: "Choose the starting roost",
        detail: "Click one of the highlighted corner clearings."
      };
    case 3:
      return {
        title: "Vagabond Setup",
        instruction: "Choose the starting forest",
        detail: "Click one of the highlighted forest markers."
      };
    default:
      return {
        title: "Setup",
        instruction: "Choose a highlighted setup target",
        detail: "The board highlights the legal setup choices."
      };
  }
}

type UseBoardInteractionOptions = {
  actions: Action[];
  activeModal: "correction" | "json" | "standAndDeliver" | null;
  multiplayerToken: string | null;
  parsedState: GameState;
  previewedAction: Action | null;
  setStatus: (status: string) => void;
  setShowBoardEditor: (show: boolean) => void;
  onApply: (action: Action) => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
};

export function useBoardInteraction({
  actions,
  activeModal,
  multiplayerToken,
  parsedState,
  previewedAction,
  setStatus,
  setShowBoardEditor,
  onApply,
  onOpenBattle
}: UseBoardInteractionOptions) {
  const [selectedClearingID, setSelectedClearingID] = useState<number>(parsedState.map.clearings[0]?.id ?? 0);
  const [assistBattleCandidateRefs, setAssistBattleCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistMovementCandidateRefs, setAssistMovementCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistMovementSource, setAssistMovementSource] = useState<AssistMovementSource | null>(null);
  const [playerMovementCandidateRefs, setPlayerMovementCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerMovementSource, setPlayerMovementSource] = useState<AssistMovementSource | null>(null);
  const [assistBuildRecruitCandidateRefs, setAssistBuildRecruitCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerBuildRecruitCandidateRefs, setPlayerBuildRecruitCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistFactionSpatialCandidateRefs, setAssistFactionSpatialCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerFactionSpatialCandidateRefs, setPlayerFactionSpatialCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [marquiseSetupDraft, setMarquiseSetupDraft] = useState<MarquiseSetupDraft>(emptyMarquiseSetupDraft);
  const [eyrieSetupDraftClearingID, setEyrieSetupDraftClearingID] = useState<number | null>(null);
  const [vagabondSetupDraftForestID, setVagabondSetupDraftForestID] = useState<number | null>(null);

  useEffect(() => {
    if (parsedState.map.clearings.some((clearing) => clearing.id === selectedClearingID)) {
      return;
    }
    setSelectedClearingID(parsedState.map.clearings[0]?.id ?? 0);
  }, [parsedState, selectedClearingID]);

  useEffect(() => {
    setMarquiseSetupDraft(emptyMarquiseSetupDraft);
    setEyrieSetupDraftClearingID(null);
    setVagabondSetupDraftForestID(null);
  }, [parsedState.gamePhase, parsedState.setupStage]);

  const {
    battleCandidates: assistBattleCandidates,
    movementCandidates: assistMovementCandidates,
    buildRecruitCandidates: assistBuildRecruitCandidates,
    factionSpatialCandidates: assistFactionSpatialCandidates
  } = buildAssistBoardCandidates({
    battleCandidates: assistBattleCandidateRefs,
    movementCandidates: [...assistMovementCandidateRefs, ...playerMovementCandidateRefs],
    buildRecruitCandidates: [...assistBuildRecruitCandidateRefs, ...playerBuildRecruitCandidateRefs],
    factionSpatialCandidates: [...assistFactionSpatialCandidateRefs, ...playerFactionSpatialCandidateRefs]
  });
  const activeMovementSource = assistMovementSource ?? playerMovementSource;
  const highlightedClearings = previewedAction
    ? affectedClearings(previewedAction)
    : assistBoardHighlights({
        battleCandidates: assistBattleCandidates,
        movementCandidates: assistMovementCandidates,
        buildRecruitCandidates: assistBuildRecruitCandidates,
        factionSpatialCandidates: assistFactionSpatialCandidates,
        movementSource: activeMovementSource
      });
  const selectedClearing =
    parsedState.map.clearings.find((clearing) => clearing.id === selectedClearingID) ??
    parsedState.map.clearings[0];
  const marquiseSetupActions = actions.filter((action) => action.type === ACTION_TYPE.MARQUISE_SETUP);
  const eyrieSetupActions = actions.filter((action) => action.type === ACTION_TYPE.EYRIE_SETUP);
  const vagabondSetupActions = actions.filter((action) => action.type === ACTION_TYPE.VAGABOND_SETUP);

  const legalSetupClearingIDs =
    parsedState.gamePhase !== 0
      ? []
      : parsedState.setupStage === 1
        ? (() => {
            if (marquiseSetupDraft.keepClearingID === null) {
              return Array.from(new Set(marquiseSetupActions.map((action) => action.marquiseSetup?.keepClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }
            const filteredByKeep = marquiseSetupActions.filter((action) => marquiseSetupMatches(action, marquiseSetupDraft));
            if (marquiseSetupDraft.sawmillClearingID === null) {
              return Array.from(new Set(filteredByKeep.map((action) => action.marquiseSetup?.sawmillClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }
            const filteredBySawmill = filteredByKeep.filter((action) => marquiseSetupMatches(action, marquiseSetupDraft));
            if (marquiseSetupDraft.workshopClearingID === null) {
              return Array.from(new Set(filteredBySawmill.map((action) => action.marquiseSetup?.workshopClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }
            return Array.from(
              new Set(
                filteredBySawmill
                  .filter((action) => marquiseSetupMatches(action, marquiseSetupDraft))
                  .map((action) => action.marquiseSetup?.recruiterClearingID ?? 0)
              )
            ).filter((value) => value > 0);
          })()
        : parsedState.setupStage === 2
          ? eyrieSetupActions.map((action) => action.eyrieSetup?.clearingID ?? 0).filter((value) => value > 0)
          : [];

  const selectedSetupClearingIDs =
    parsedState.gamePhase === 0 && parsedState.setupStage === 1
      ? [
          marquiseSetupDraft.keepClearingID,
          marquiseSetupDraft.sawmillClearingID,
          marquiseSetupDraft.workshopClearingID,
          marquiseSetupDraft.recruiterClearingID
        ].filter((value): value is number => value !== null)
      : parsedState.gamePhase === 0 && parsedState.setupStage === 2 && eyrieSetupDraftClearingID !== null
        ? [eyrieSetupDraftClearingID]
        : [];

  const setupPreviewPiecesByClearing = (() => {
    const previews: Record<number, ClearingPreviewPiece[]> = {};
    const addPreview = (clearingID: number | null, piece: ClearingPreviewPiece) => {
      if (clearingID === null) {
        return;
      }
      previews[clearingID] ??= [];
      previews[clearingID].push(piece);
    };
    if (parsedState.gamePhase === 0 && parsedState.setupStage === 1) {
      addPreview(marquiseSetupDraft.keepClearingID, { kind: "keep", label: "Pending Keep", preview: true });
      addPreview(marquiseSetupDraft.sawmillClearingID, { kind: "sawmill", label: "Pending sawmill", preview: true });
      addPreview(marquiseSetupDraft.workshopClearingID, { kind: "workshop", label: "Pending workshop", preview: true });
      addPreview(marquiseSetupDraft.recruiterClearingID, { kind: "recruiter", label: "Pending recruiter", preview: true });
    }
    if (parsedState.gamePhase === 0 && parsedState.setupStage === 2) {
      addPreview(eyrieSetupDraftClearingID, { kind: "roost", label: "Pending roost", preview: true });
      addPreview(eyrieSetupDraftClearingID, { kind: "eyrie", count: 6, label: "Pending Eyrie warriors 6", preview: true });
    }
    return previews;
  })();

  const forestTargets =
    parsedState.gamePhase === 0 && parsedState.setupStage === 3
      ? parsedState.map.forests.map((forest) => ({
          forestID: forest.id,
          label: `Forest ${forest.id}`,
          legal: vagabondSetupActions.some((action) => action.vagabondSetup?.forestID === forest.id),
          selected: vagabondSetupDraftForestID === forest.id
        }))
      : assistMovementCandidates.length > 0 && parsedState.gamePhase === 1
        ? parsedState.map.forests
            .map((forest) => {
              const legal =
                activeMovementSource === null
                  ? assistMovementCandidates.some((candidate) => candidate.endpoints.fromForestID === forest.id)
                  : assistMovementCandidates.some(
                      (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toForestID === forest.id
                    );
              return {
                forestID: forest.id,
                label: `Forest ${forest.id}`,
                legal,
                selected: activeMovementSource?.kind === "forest" && activeMovementSource.id === forest.id
              };
            })
            .filter((forest) => forest.legal || forest.selected)
        : [];

  const handleAssistBattleCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistBattleCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handleAssistMovementCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistMovementCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
    setAssistMovementSource(null);
  }, []);

  const handlePlayerMovementCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerMovementCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
    setPlayerMovementSource(null);
  }, []);

  const handleAssistBuildRecruitCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistBuildRecruitCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handlePlayerBuildRecruitCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerBuildRecruitCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handleAssistFactionSpatialCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistFactionSpatialCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handlePlayerFactionSpatialCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerFactionSpatialCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  function setBoardMovementSource(source: AssistMovementSource | null) {
    if (parsedState.factionTurn === parsedState.playerFaction) {
      setPlayerMovementSource(source);
      setAssistMovementSource(null);
      return;
    }
    setAssistMovementSource(source);
    setPlayerMovementSource(null);
  }

  async function handleSetupClearingClick(clearingID: number) {
    if (assistBattleCandidates.length > 0 && parsedState.gamePhase === 1 && parsedState.factionTurn !== parsedState.playerFaction) {
      setSelectedClearingID(clearingID);
      const matchingBattles = assistBattleCandidates.filter((candidate) => candidate.action.battle?.clearingID === clearingID);
      if (matchingBattles.length === 1) {
        onOpenBattle(matchingBattles[0].actionIndex);
        setStatus(`Battle selected in clearing ${clearingID}. Resolve it from Battle Flow.`);
        return;
      }
      if (matchingBattles.length > 1) {
        setStatus(`Clearing ${clearingID} has multiple battle targets. Choose the observed defender in the Battle prompt.`);
        return;
      }
      setStatus("Choose one of the highlighted battle clearings.");
      return;
    }

    if (assistMovementCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      if (activeMovementSource === null) {
        const sourceMatches = assistMovementCandidates.filter((candidate) => candidate.endpoints.fromClearingID === clearingID);
        if (sourceMatches.length === 0) {
          setStatus("Choose one of the highlighted move source clearings.");
          return;
        }
        const uniqueClearingTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toClearingID))).filter((value) => value > 0);
        const uniqueForestTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toForestID))).filter((value) => value > 0);
        if (sourceMatches.length === 1 && uniqueClearingTargets.length === 1 && uniqueForestTargets.length === 0) {
          setStatus(`Recording movement from clearing ${clearingID} to clearing ${uniqueClearingTargets[0]}...`);
          await onApply(sourceMatches[0].action);
          return;
        }
        setBoardMovementSource({ kind: "clearing", id: clearingID });
        setStatus(`Move source selected: clearing ${clearingID}. Choose a highlighted destination.`);
        return;
      }

      const matchingMoves = assistMovementCandidates.filter(
        (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toClearingID === clearingID
      );
      if (matchingMoves.length === 1) {
        const sourceLabel = activeMovementSource.kind === "clearing" ? `clearing ${activeMovementSource.id}` : `forest ${activeMovementSource.id}`;
        setStatus(`Recording movement from ${sourceLabel} to clearing ${clearingID}...`);
        await onApply(matchingMoves[0].action);
        setBoardMovementSource(null);
        return;
      }
      if (matchingMoves.length > 1) {
        setStatus("Multiple move options match that route. Choose the exact one from the Move tray.");
        return;
      }
      if (assistMovementCandidates.some((candidate) => candidate.endpoints.fromClearingID === clearingID)) {
        setBoardMovementSource({ kind: "clearing", id: clearingID });
        setStatus(`Move source changed to clearing ${clearingID}. Choose a highlighted destination.`);
        return;
      }
      setStatus("Choose a highlighted move destination, or click another highlighted source to restart the route.");
      return;
    }

    if (assistBuildRecruitCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      const matchingCandidates = assistBuildRecruitCandidates.filter((candidate) => candidate.clearingIDs.includes(clearingID));
      if (matchingCandidates.length === 1) {
        setStatus(`Recording ${actionHeadline(matchingCandidates[0].action).toLowerCase()} at clearing ${clearingID}...`);
        await onApply(matchingCandidates[0].action);
        return;
      }
      if (matchingCandidates.length > 1) {
        setStatus(`Clearing ${clearingID} matches multiple Build / Recruit options. Choose the exact one from the tray.`);
        return;
      }
      setStatus("Choose one of the highlighted build, recruit, or overwork clearings.");
      return;
    }

    if (assistFactionSpatialCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      const matchingCandidates = assistFactionSpatialCandidates.filter((candidate) => candidate.clearingIDs.includes(clearingID));
      if (matchingCandidates.length === 1) {
        setStatus(`Recording ${actionHeadline(matchingCandidates[0].action).toLowerCase()} at clearing ${clearingID}...`);
        await onApply(matchingCandidates[0].action);
        return;
      }
      if (matchingCandidates.length > 1) {
        setStatus(`Clearing ${clearingID} matches multiple faction-action options. Choose the exact one from the tray.`);
        return;
      }
      setStatus("Choose one of the highlighted faction-action clearings.");
      return;
    }

    if (parsedState.gamePhase !== 0) {
      setSelectedClearingID(clearingID);
      if (!multiplayerToken && activeModal === "correction") {
        setShowBoardEditor(true);
        setStatus(`Selected clearing ${clearingID} for board editing.`);
        return;
      }
      setStatus(`Selected clearing ${clearingID}.`);
      return;
    }

    if (!legalSetupClearingIDs.includes(clearingID)) {
      setStatus("Choose one of the highlighted setup targets.");
      return;
    }

    if (parsedState.setupStage === 1) {
      if (marquiseSetupDraft.keepClearingID === null) {
        setMarquiseSetupDraft({ ...emptyMarquiseSetupDraft, keepClearingID: clearingID });
        setStatus("Choose the starting sawmill location.");
        return;
      }
      if (marquiseSetupDraft.sawmillClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, sawmillClearingID: clearingID });
        setStatus("Choose the starting workshop location.");
        return;
      }
      if (marquiseSetupDraft.workshopClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, workshopClearingID: clearingID });
        setStatus("Choose the starting recruiter location.");
        return;
      }

      const finalDraft = { ...marquiseSetupDraft, recruiterClearingID: clearingID };
      const action = marquiseSetupActions.find((candidate) => marquiseSetupMatches(candidate, finalDraft));
      if (!action) {
        setStatus("That building placement is not legal.");
        return;
      }
      setMarquiseSetupDraft(finalDraft);
      setStatus("Applying Marquise setup...");
      await onApply(action);
      return;
    }

    if (parsedState.setupStage === 2) {
      const action = eyrieSetupActions.find((candidate) => candidate.eyrieSetup?.clearingID === clearingID);
      if (!action) {
        setStatus("That starting clearing is not legal for the Eyrie.");
        return;
      }
      setEyrieSetupDraftClearingID(clearingID);
      setStatus("Applying Eyrie setup...");
      await onApply(action);
    }
  }

  async function handleSetupForestClick(forestID: number) {
    if (assistMovementCandidates.length > 0 && parsedState.gamePhase === 1) {
      if (activeMovementSource === null) {
        const sourceMatches = assistMovementCandidates.filter((candidate) => candidate.endpoints.fromForestID === forestID);
        if (sourceMatches.length === 0) {
          setStatus("Choose one of the highlighted move source forests.");
          return;
        }
        const uniqueClearingTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toClearingID))).filter((value) => value > 0);
        const uniqueForestTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toForestID))).filter((value) => value > 0);
        if (sourceMatches.length === 1 && uniqueClearingTargets.length === 1 && uniqueForestTargets.length === 0) {
          setStatus(`Recording movement from forest ${forestID} to clearing ${uniqueClearingTargets[0]}...`);
          await onApply(sourceMatches[0].action);
          return;
        }
        setBoardMovementSource({ kind: "forest", id: forestID });
        setStatus(`Move source selected: forest ${forestID}. Choose a highlighted destination.`);
        return;
      }

      const matchingMoves = assistMovementCandidates.filter(
        (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toForestID === forestID
      );
      if (matchingMoves.length === 1) {
        const sourceLabel = activeMovementSource.kind === "clearing" ? `clearing ${activeMovementSource.id}` : `forest ${activeMovementSource.id}`;
        setStatus(`Recording movement from ${sourceLabel} to forest ${forestID}...`);
        await onApply(matchingMoves[0].action);
        setBoardMovementSource(null);
        return;
      }
      if (matchingMoves.length > 1) {
        setStatus("Multiple move options match that forest route. Choose the exact one from the Move tray.");
        return;
      }
      if (assistMovementCandidates.some((candidate) => candidate.endpoints.fromForestID === forestID)) {
        setBoardMovementSource({ kind: "forest", id: forestID });
        setStatus(`Move source changed to forest ${forestID}. Choose a highlighted destination.`);
        return;
      }
      setStatus("Choose a highlighted move forest, or click another highlighted source to restart the route.");
      return;
    }

    if (parsedState.gamePhase !== 0 || parsedState.setupStage !== 3) {
      return;
    }

    const action = vagabondSetupActions.find((candidate) => candidate.vagabondSetup?.forestID === forestID);
    if (!action) {
      setStatus("That forest is not a legal Vagabond starting forest.");
      return;
    }
    setVagabondSetupDraftForestID(forestID);
    setStatus("Applying Vagabond setup...");
    await onApply(action);
  }

  return {
    selectedClearingID,
    selectedClearing,
    setSelectedClearingID,
    highlightedClearings,
    legalSetupClearingIDs,
    selectedSetupClearingIDs,
    setupPreviewPiecesByClearing,
    forestTargets,
    setupLegalChoiceCount: parsedState.setupStage === 3 ? forestTargets.filter((target) => target.legal).length : legalSetupClearingIDs.length,
    hasMarquiseDraftSelection:
      marquiseSetupDraft.keepClearingID !== null ||
      marquiseSetupDraft.sawmillClearingID !== null ||
      marquiseSetupDraft.workshopClearingID !== null ||
      marquiseSetupDraft.recruiterClearingID !== null,
    marquiseSetupDraft,
    setMarquiseSetupDraft,
    handleSetupClearingClick,
    handleSetupForestClick,
    handleAssistBattleCandidatesChange,
    handleAssistMovementCandidatesChange,
    handlePlayerMovementCandidatesChange,
    handleAssistBuildRecruitCandidatesChange,
    handlePlayerBuildRecruitCandidatesChange,
    handleAssistFactionSpatialCandidatesChange,
    handlePlayerFactionSpatialCandidatesChange
  };
}
