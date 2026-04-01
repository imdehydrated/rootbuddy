import { factionLabels } from "../labels";
import type { Action, BattleContext, BattleModifiers, BattlePrompt } from "../types";

type BattleFlowPanelProps = {
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

export function BattleFlowPanel({
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

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Battle Flow</p>
      <div className="summary-stack">
        <span className="summary-label">
          {factionLabels[attackerFaction] ?? "Unknown"} vs {factionLabels[defenderFaction] ?? "Unknown"}
        </span>
        <span className="summary-line">
          Clearing {battleAction.battle.clearingID} {selectedBattleIndex !== null ? `- Action ${selectedBattleIndex + 1}` : ""}
        </span>
      </div>

      {multiplayerBattlePrompt ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Multiplayer Prompt</span>
          {isWaitingPrompt ? (
            <span className="summary-line">Waiting on {multiplayerWaitingLabel} to respond.</span>
          ) : null}
          {isDefenderPrompt ? (
            <span className="summary-line">Defender response required before battle resolution.</span>
          ) : null}
          {isAttackerPrompt ? (
            <span className="summary-line">Attacker follow-up response required before battle resolution.</span>
          ) : null}
          {isReadyPrompt ? (
            <span className="summary-line">
              {localPlayerCanResolve
                ? "Responses are complete. The server will roll battle dice during resolution."
                : `Responses are complete. Waiting on ${factionLabels[attackerFaction] ?? "the attacker"} to resolve.`}
            </span>
          ) : null}
        </div>
      ) : (
        <div className="resolve-grid" style={{ marginTop: "0.9rem" }}>
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
        <div className="summary-stack" style={{ marginTop: "1rem" }}>
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

      <div className="control-grid" style={{ marginTop: "1rem" }}>
        {!assistDefenderAmbushPromptRequired && (!multiplayerBattlePrompt || isDefenderPrompt) ? (
          <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.defenderAmbush}
              disabled={!defenderCanAmbush || (!!multiplayerBattlePrompt && !isDefenderPrompt)}
              onChange={(event) =>
                onSetBattleModifiers((current) => ({
                  ...current,
                  defenderAmbush: event.target.checked,
                  attackerCounterAmbush: event.target.checked ? current.attackerCounterAmbush : false
                }))
              }
            />
            Defender Ambush
          </label>
        ) : null}
        <label className="checkbox">
          <input
            type="checkbox"
            checked={battleModifiers.attackerCounterAmbush}
            disabled={
              (multiplayerBattlePrompt !== null && !isAttackerPrompt) ||
              !(assistDefenderAmbushPromptRequired ? assistDefenderAmbushChoice === true : battleModifiers.defenderAmbush) ||
              !attackerCanCounterAmbush
            }
            onChange={(event) =>
              onSetBattleModifiers((current) => ({
                ...current,
                attackerCounterAmbush: event.target.checked
              }))
            }
          />
          Attacker Counter-Ambush
        </label>
        <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.attackerUsesArmorers}
            disabled={!attackerCanArmorers || (!!multiplayerBattlePrompt && !isAttackerPrompt)}
            onChange={(event) =>
              onSetBattleModifiers((current) => ({
                ...current,
                attackerUsesArmorers: event.target.checked
              }))
            }
          />
          Attacker Armorers
        </label>
        <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.defenderUsesArmorers}
            disabled={!defenderCanArmorers || (!!multiplayerBattlePrompt && !isDefenderPrompt)}
            onChange={(event) =>
              onSetBattleModifiers((current) => ({
                ...current,
                defenderUsesArmorers: event.target.checked
              }))
            }
          />
          Defender Armorers
        </label>
        <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.attackerUsesBrutalTactics}
            disabled={!attackerCanBrutalTactics || (!!multiplayerBattlePrompt && !isAttackerPrompt)}
            onChange={(event) =>
              onSetBattleModifiers((current) => ({
                ...current,
                attackerUsesBrutalTactics: event.target.checked
              }))
            }
          />
          Attacker Brutal Tactics
        </label>
        <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.defenderUsesSappers}
            disabled={!defenderCanSappers || (!!multiplayerBattlePrompt && !isDefenderPrompt)}
            onChange={(event) =>
              onSetBattleModifiers((current) => ({
                ...current,
                defenderUsesSappers: event.target.checked
              }))
            }
          />
          Defender Sappers
        </label>
      </div>

      {attackerHasScoutingParty ? (
        <p className="message" style={{ marginTop: "0.8rem" }}>
          Attacker has Scouting Party, so defender ambushes are ignored.
        </p>
      ) : null}

      <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
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
