import type { UserSettings } from "../settings";

type SettingsPanelProps = {
  settings: UserSettings;
  onChange: (setting: keyof UserSettings, value: boolean) => void;
  onReset: () => void;
  onClose: () => void;
};

const settingRows: Array<{
  key: keyof UserSettings;
  title: string;
  detail: string;
}> = [
  {
    key: "showGameLog",
    title: "Show Game Log",
    detail: "Keep the multiplayer recent-action panel visible on the board."
  },
  {
    key: "showVPTracker",
    title: "Show VP Tracker",
    detail: "Keep the always-visible score track in the top board HUD."
  },
  {
    key: "showCardTray",
    title: "Show Hand Tray",
    detail: "Keep the visible hand mounted above the action tray."
  },
  {
    key: "compactCards",
    title: "Compact Cards",
    detail: "Use the tighter card layout in board-facing hand surfaces."
  }
];

export function SettingsPanel({ settings, onChange, onReset, onClose }: SettingsPanelProps) {
  return (
    <section className="panel modal-panel settings-panel">
      <div className="panel-header">
        <h2>Workspace Settings</h2>
        <button type="button" className="secondary" onClick={onClose}>
          Close
        </button>
      </div>
      <p className="message">These settings are local to this browser and only change how the frontend workspace presents board information.</p>

      <div className="settings-list" role="group" aria-label="Workspace settings">
        {settingRows.map((row) => (
          <label key={row.key} className="settings-row">
            <div className="settings-copy">
              <strong>{row.title}</strong>
              <span className="summary-line">{row.detail}</span>
            </div>
            <input
              type="checkbox"
              aria-label={row.title}
              checked={settings[row.key]}
              onChange={(event) => onChange(row.key, event.target.checked)}
            />
          </label>
        ))}
      </div>

      <div className="sidebar-actions footer">
        <button type="button" className="secondary" onClick={onReset}>
          Reset Defaults
        </button>
      </div>
    </section>
  );
}
