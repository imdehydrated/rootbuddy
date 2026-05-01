import { factionLabels } from "../labels";
import type { Action, BattleContext, BattleModifiers, BattlePrompt } from "../types";

type BattleFlowPanelProps = {
  surface?: "sidebar" | "modal";
  selectedBattleIndex: number | null;
  selectedBattleAction: Action | null;
  multiplayerBattlePrompt: BattlePrompt | null;
  multiplayerPerspectiveFaction: number | null;
  multiplayerSubmitting: boolean;
  attackerFaction: number;
  defenderFaction: number;
  attackerRoll: string;
  defenderRoll: string;
  battleModifiers: BattleModifiers;
  battleContext: BattleContext | null;
  assistDefenderAmbushChoice: boolean | null;
  onSetAttackerRoll: (value: string) => void;
  onSetDefenderRoll: (value: string) => void;
  onSetBattleModifiers: (updater: (current: BattleModifiers) => BattleModifiers) => void;
  onSetAssistDefenderAmbushChoice: (value: boolean | null) => void;
  onSubmitMultiplayerResponse: () => Promise<void>;
  onResolveAndApply: () => Promise<void>;
  onClearSelection: () => void;
};

type EffectOption = {
  key: string;
  label: string;
  owner: string;
  available: boolean;
  selected: boolean;
  disabled: boolean;
  status: string;
  onToggle?: () => void;
};

export function BattleFlowPanel({
  surface = "sidebar",
  selectedBattleIndex,
  selectedBattleAction,
  multiplayerBattlePrompt,
  multiplayerPerspectiveFaction,
  multiplayerSubmitting,
  attackerFaction,
  defenderFaction,
  attackerRoll,
  defenderRoll,
  battleModifiers,
  battleContext,
  assistDefenderAmbushChoice,
  onSetAttackerRoll,
  onSetDefenderRoll,
  onSetBattleModifiers,
  onSetAssistDefenderAmbushChoice,
  onSubmitMultiplayerResponse,
  onResolveAndApply,
  onClearSelection
}: BattleFlowPanelProps) {
  const battleAction = multiplayerBattlePrompt?.action ?? selectedBattleAction;
  const effectiveContext = multiplayerBattlePrompt?.battleContext ?? battleContext;
  if (!battleAction?.battle) {
    return null;
  }

  const attackerHasScoutingParty = effectiveContext?.attackerHasScoutingParty ?? false;
  const canDefenderAmbush = effectiveContext?.canDefenderAmbush ?? false;
  const assistDefenderAmbushPromptRequired = effectiveContext?.assistDefenderAmbushPromptRequired ?? false;
  const canAttackerCounterAmbush = effectiveContext?.canAttackerCounterAmbush ?? false;
  const canAttackerArmorers = effectiveContext?.canAttackerArmorers ?? false;
  const canDefenderArmorers = effectiveContext?.canDefenderArmorers ?? false;
  const canAttackerBrutalTactics = effectiveContext?.canAttackerBrutalTactics ?? false;
  const canDefenderSappers = effectiveContext?.canDefenderSappers ?? false;
  const defenderCanAmbush = multiplayerBattlePrompt?.canUseAmbush ?? canDefenderAmbush;
  const defenderCanArmorers = multiplayerBattlePrompt?.canUseDefenderArmorers ?? canDefenderArmorers;
  const defenderCanSappers = multiplayerBattlePrompt?.canUseSappers ?? canDefenderSappers;
  const attackerCanCounterAmbush = multiplayerBattlePrompt?.canUseCounterAmbush ?? canAttackerCounterAmbush;
  const attackerCanArmorers = multiplayerBattlePrompt?.canUseAttackerArmorers ?? canAttackerArmorers;
  const attackerCanBrutalTactics = multiplayerBattlePrompt?.canUseBrutalTactics ?? canAttackerBrutalTactics;
  const multiplayerStage = multiplayerBattlePrompt?.stage ?? null;
  const multiplayerWaitingLabel = multiplayerBattlePrompt
    ? factionLabels[multiplayerBattlePrompt.waitingOnFaction] ?? "another player"
    : "";
  const isDefenderPrompt = multiplayerStage === "defender_response";
  const isAttackerPrompt = multiplayerStage === "attacker_response";
  const isReadyPrompt = multiplayerStage === "ready_to_resolve";
  const isWaitingPrompt = multiplayerStage === "waiting_defender" || multiplayerStage === "waiting_attacker";
  const localPlayerOwnsPrompt =
    multiplayerBattlePrompt !== null && multiplayerBattlePrompt.waitingOnFaction === multiplayerPerspectiveFaction;
  const localPlayerCanResolve =
    multiplayerBattlePrompt !== null && isReadyPrompt && multiplayerPerspectiveFaction === attackerFaction;
  const defenderAmbushToggleDisabled =
    assistDefenderAmbushPromptRequired || !defenderCanAmbush || (!!multiplayerBattlePrompt && !isDefenderPrompt);
  const counterAmbushToggleDisabled =
    (multiplayerBattlePrompt !== null && !isAttackerPrompt) ||
    !(assistDefenderAmbushPromptRequired ? assistDefenderAmbushChoice === true : battleModifiers.defenderAmbush) ||
    !attackerCanCounterAmbush;
  const effectOptions: EffectOption[] = [
    {
      key: "scouting-party",
      label: "Scouting Party",
      owner: factionLabels[attackerFaction] ?? "Attacker",
      available: attackerHasScoutingParty,
      selected: attackerHasScoutingParty,
      disabled: true,
      status: "Passive effect"
    },
    {
      key: "defender-ambush",
      label: "Ambush",
      owner: factionLabels[defenderFaction] ?? "Defender",
      available: defenderCanAmbush,
      selected: battleModifiers.defenderAmbush,
      disabled: defenderAmbushToggleDisabled,
      status: battleModifiers.defenderAmbush ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          defenderAmbush: !current.defenderAmbush,
          attackerCounterAmbush: !current.defenderAmbush ? current.attackerCounterAmbush : false
        }))
    },
    {
      key: "attacker-counter-ambush",
      label: "Counter-Ambush",
      owner: factionLabels[attackerFaction] ?? "Attacker",
      available: attackerCanCounterAmbush,
      selected: battleModifiers.attackerCounterAmbush,
      disabled: counterAmbushToggleDisabled,
      status: battleModifiers.attackerCounterAmbush ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          attackerCounterAmbush: !current.attackerCounterAmbush
        }))
    },
    {
      key: "attacker-armorers",
      label: "Armorers",
      owner: factionLabels[attackerFaction] ?? "Attacker",
      available: attackerCanArmorers,
      selected: battleModifiers.attackerUsesArmorers,
      disabled: !attackerCanArmorers || (!!multiplayerBattlePrompt && !isAttackerPrompt),
      status: battleModifiers.attackerUsesArmorers ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          attackerUsesArmorers: !current.attackerUsesArmorers
        }))
    },
    {
      key: "defender-armorers",
      label: "Armorers",
      owner: factionLabels[defenderFaction] ?? "Defender",
      available: defenderCanArmorers,
      selected: battleModifiers.defenderUsesArmorers,
      disabled: !defenderCanArmorers || (!!multiplayerBattlePrompt && !isDefenderPrompt),
      status: battleModifiers.defenderUsesArmorers ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          defenderUsesArmorers: !current.defenderUsesArmorers
        }))
    },
    {
      key: "attacker-brutal-tactics",
      label: "Brutal Tactics",
      owner: factionLabels[attackerFaction] ?? "Attacker",
      available: attackerCanBrutalTactics,
      selected: battleModifiers.attackerUsesBrutalTactics,
      disabled: !attackerCanBrutalTactics || (!!multiplayerBattlePrompt && !isAttackerPrompt),
      status: battleModifiers.attackerUsesBrutalTactics ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          attackerUsesBrutalTactics: !current.attackerUsesBrutalTactics
        }))
    },
    {
      key: "defender-sappers",
      label: "Sappers",
      owner: factionLabels[defenderFaction] ?? "Defender",
      available: defenderCanSappers,
      selected: battleModifiers.defenderUsesSappers,
      disabled: !defenderCanSappers || (!!multiplayerBattlePrompt && !isDefenderPrompt),
      status: battleModifiers.defenderUsesSappers ? "Selected for this battle" : "Click to use",
      onToggle: () =>
        onSetBattleModifiers((current) => ({
          ...current,
          defenderUsesSappers: !current.defenderUsesSappers
        }))
    }
  ].filter((effect) => (effect.key === "scouting-party" ? effect.available : effect.available || effect.selected));
  const chosenEffects = effectOptions.filter((effect) => effect.selected);
  const promptHasRolls = multiplayerBattlePrompt?.attackerRoll !== undefined && multiplayerBattlePrompt?.defenderRoll !== undefined;

  return (
    <section className={`panel ${surface === "modal" ? "battle-event-panel" : "sidebar-panel"}`}>
      <p className="eyebrow">Battle Event</p>
      <div className="battle-event-hero">
        <div className="battle-event-header">
          <div className="summary-stack">
            <span className="summary-label">Battle</span>
            <strong>
              {factionLabels[attackerFaction] ?? "Unknown"} vs {factionLabels[defenderFaction] ?? "Unknown"}
            </strong>
          </div>
          <span className="battle-clearing-pill">
            Clearing {battleAction.battle.clearingID}
            {selectedBattleIndex !== null ? ` • Action ${selectedBattleIndex + 1}` : ""}
          </span>
        </div>
        <span className="summary-line">
          Resolve the confrontation here, then return to the board flow.
        </span>
      </div>

      {effectOptions.length > 0 ? (
        <div className="summary-stack battle-section">
          <span className="summary-label">Effect Cards</span>
          <div className="battle-effect-grid">
            {effectOptions.map((effect) => (
              <button
                key={effect.key}
                type="button"
                className={`battle-effect-card toggle ${effect.selected ? "selected" : ""} ${effect.disabled ? "disabled" : ""}`}
                disabled={effect.disabled}
                onClick={() => effect.onToggle?.()}
              >
                <span className="summary-label">{effect.owner}</span>
                <strong>{effect.label}</strong>
                <span className="summary-line">{effect.status}</span>
              </button>
            ))}
          </div>
        </div>
      ) : null}

      {multiplayerBattlePrompt ? (
        <div className="summary-stack battle-section battle-prompt-card">
          <span className="summary-label">Multiplayer Prompt</span>
          {isWaitingPrompt ? (
            <span className="summary-line">Waiting on {multiplayerWaitingLabel} to respond.</span>
          ) : null}
          {isDefenderPrompt ? (
            <span className="summary-line">
              {promptHasRolls ? "Dice are rolled. Defender after-roll response required." : "Defender response required before battle resolution."}
            </span>
          ) : null}
          {isAttackerPrompt ? (
            <span className="summary-line">
              {promptHasRolls ? "Dice are rolled. Attacker after-roll response required." : "Attacker follow-up response required before battle resolution."}
            </span>
          ) : null}
          {isReadyPrompt ? (
            <span className="summary-line">
              {localPlayerCanResolve
                ? "Responses are complete. The server will roll battle dice during resolution."
                : `Responses are complete. Waiting on ${factionLabels[attackerFaction] ?? "the attacker"} to resolve.`}
            </span>
          ) : null}
          {chosenEffects.length > 0 ? (
            <div className="known-card-pill-list">
              {chosenEffects.map((effect) => (
                <span key={`chosen-${effect.owner}-${effect.label}`} className="known-card-pill">
                  {effect.owner}: {effect.label}
                </span>
              ))}
            </div>
          ) : null}
          {promptHasRolls ? (
            <div className="known-card-pill-list">
              <span className="known-card-pill">Attacker roll: {multiplayerBattlePrompt.attackerRoll}</span>
              <span className="known-card-pill">Defender roll: {multiplayerBattlePrompt.defenderRoll}</span>
            </div>
          ) : null}
        </div>
      ) : (
        <div className="resolve-grid battle-section battle-roll-grid">
          <label>
            <span>Attacker Roll</span>
            <input type="number" min="0" max="3" value={attackerRoll} onChange={(event) => onSetAttackerRoll(event.target.value)} />
          </label>
          <label>
            <span>Defender Roll</span>
            <input type="number" min="0" max="3" value={defenderRoll} onChange={(event) => onSetDefenderRoll(event.target.value)} />
          </label>
        </div>
      )}

      {assistDefenderAmbushPromptRequired && !multiplayerBattlePrompt ? (
        <div className="summary-stack battle-section battle-prompt-card">
          <span className="summary-label">Assist Prompt</span>
          <span className="summary-line">Did {factionLabels[defenderFaction] ?? "the defender"} play an Ambush?</span>
          <div className="sidebar-actions">
            <button
              type="button"
              className={assistDefenderAmbushChoice === true ? "" : "secondary"}
              onClick={() => {
                onSetAssistDefenderAmbushChoice(true);
                onSetBattleModifiers((current) => ({ ...current, defenderAmbush: true }));
              }}
            >
              Yes
            </button>
            <button
              type="button"
              className={assistDefenderAmbushChoice === false ? "" : "secondary"}
              onClick={() => {
                onSetAssistDefenderAmbushChoice(false);
                onSetBattleModifiers((current) => ({
                  ...current,
                  defenderAmbush: false,
                  attackerCounterAmbush: false
                }));
              }}
            >
              No
            </button>
          </div>
        </div>
      ) : null}

      {attackerHasScoutingParty ? (
        <p className="message battle-inline-note">
          Attacker has Scouting Party, so defender ambushes are ignored.
        </p>
      ) : null}

      <div className="sidebar-actions footer battle-event-actions">
        {multiplayerBattlePrompt ? (
          <>
            {(isDefenderPrompt || isAttackerPrompt) && localPlayerOwnsPrompt ? (
              <button type="button" onClick={() => void onSubmitMultiplayerResponse()} disabled={multiplayerSubmitting}>
                Submit Response
              </button>
            ) : null}
            {localPlayerCanResolve ? (
              <button type="button" onClick={() => void onResolveAndApply()} disabled={multiplayerSubmitting}>
                Resolve and Apply
              </button>
            ) : null}
            <button type="button" className="secondary" onClick={onClearSelection} disabled>
              Active Prompt
            </button>
          </>
        ) : (
          <>
            <button type="button" onClick={() => void onResolveAndApply()}>
              Resolve and Apply
            </button>
            <button type="button" className="secondary" onClick={onClearSelection}>
              Clear
            </button>
          </>
        )}
      </div>
    </section>
  );
}
