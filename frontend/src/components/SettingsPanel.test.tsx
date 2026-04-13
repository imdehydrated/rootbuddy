import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { UserSettings } from "../settings";
import { SettingsPanel } from "./SettingsPanel";

function sampleSettings(overrides: Partial<UserSettings> = {}): UserSettings {
  return {
    showGameLog: true,
    showVPTracker: true,
    showCardTray: false,
    compactCards: false,
    ...overrides
  };
}

describe("SettingsPanel", () => {
  it("renders the current settings state", () => {
    render(
      <SettingsPanel
        settings={sampleSettings({ compactCards: true })}
        onChange={vi.fn()}
        onReset={vi.fn()}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByRole("heading", { name: "Workspace Settings" })).toBeInTheDocument();
    expect(screen.getByLabelText("Show Game Log")).toBeChecked();
    expect(screen.getByLabelText("Compact Cards")).toBeChecked();
    expect(screen.getByLabelText("Show Hand Tray")).not.toBeChecked();
  });

  it("emits change, reset, and close actions", () => {
    const onChange = vi.fn();
    const onReset = vi.fn();
    const onClose = vi.fn();

    render(
      <SettingsPanel
        settings={sampleSettings()}
        onChange={onChange}
        onReset={onReset}
        onClose={onClose}
      />
    );

    fireEvent.click(screen.getByLabelText("Compact Cards"));
    expect(onChange).toHaveBeenCalledWith("compactCards", true);

    fireEvent.click(screen.getByRole("button", { name: "Reset Defaults" }));
    expect(onReset).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
