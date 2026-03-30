import { factionLabels } from "../labels";
import type { Action, BattleContext, BattleModifiers } from "../types";

type BattleFlowPanelProps = {
  selectedBattleIndex: number | null;
  selectedBattleAction: Action | null;
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
  onResolveAndApply: () => Promise<void>;
  onClearSelection: () => void;
};

export function BattleFlowPanel({
  selectedBattleIndex,
  selectedBattleAction,
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
  onResolveAndApply,
  onClearSelection
}: BattleFlowPanelProps) {
  if (!selectedBattleAction?.battle) {
    return null;
  }

  const attackerHasScoutingParty = battleContext?.attackerHasScoutingParty ?? false;
  const canDefenderAmbush = battleContext?.canDefenderAmbush ?? false;
  const assistDefenderAmbushPromptRequired = battleContext?.assistDefenderAmbushPromptRequired ?? false;
  const canAttackerCounterAmbush = battleContext?.canAttackerCounterAmbush ?? false;
  const canAttackerArmorers = battleContext?.canAttackerArmorers ?? false;
  const canDefenderArmorers = battleContext?.canDefenderArmorers ?? false;
  const canAttackerBrutalTactics = battleContext?.canAttackerBrutalTactics ?? false;
  const canDefenderSappers = battleContext?.canDefenderSappers ?? false;

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Battle Flow</p>
      <div className="summary-stack">
        <span className="summary-label">
          {factionLabels[attackerFaction] ?? "Unknown"} vs {factionLabels[defenderFaction] ?? "Unknown"}
        </span>
        <span className="summary-line">
          Clearing {selectedBattleAction.battle.clearingID} {selectedBattleIndex !== null ? `- Action ${selectedBattleIndex + 1}` : ""}
        </span>
      </div>

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

      {assistDefenderAmbushPromptRequired ? (
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
        {!assistDefenderAmbushPromptRequired ? (
          <label className="checkbox">
            <input
              type="checkbox"
              checked={battleModifiers.defenderAmbush}
              disabled={!canDefenderAmbush}
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
              !(assistDefenderAmbushPromptRequired ? assistDefenderAmbushChoice === true : battleModifiers.defenderAmbush) ||
              !canAttackerCounterAmbush
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
            disabled={!canAttackerArmorers}
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
            disabled={!canDefenderArmorers}
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
            disabled={!canAttackerBrutalTactics}
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
            disabled={!canDefenderSappers}
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
        <button type="button" onClick={() => void onResolveAndApply()}>
          Resolve and Apply
        </button>
        <button type="button" className="secondary" onClick={onClearSelection}>
          Clear
        </button>
      </div>
    </section>
  );
}
