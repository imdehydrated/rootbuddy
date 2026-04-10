import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Card } from "../types";
import { CardComponent } from "./CardComponent";

function sampleCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 7,
    deck: 0,
    name: "Anvil",
    suit: 0,
    kind: 0,
    craftingCost: {
      fox: 1,
      rabbit: 0,
      mouse: 0,
      any: 0
    },
    craftedItem: 3,
    effectID: "stand_and_deliver",
    vp: 1,
    ...overrides
  };
}

describe("CardComponent", () => {
  it("renders suit, cost, and card details", () => {
    render(<CardComponent card={sampleCard()} zoneLabel="Hand" />);

    expect(screen.getByText("Anvil")).toBeInTheDocument();
    expect(screen.getByText("Fox")).toBeInTheDocument();
    expect(screen.getByText("Stand and Deliver!")).toBeInTheDocument();
    expect(screen.getByText("Item reward: Hammer")).toBeInTheDocument();
    expect(screen.getByText("1 fox")).toBeInTheDocument();
  });

  it("supports compact rendering", () => {
    render(<CardComponent card={sampleCard()} compact />);

    expect(screen.getByText("Anvil").closest(".card-component")).toHaveClass("compact");
  });

  it("supports selectable cards", () => {
    const onSelect = vi.fn();

    render(<CardComponent card={sampleCard()} selected onSelect={onSelect} />);

    const cardButton = screen.getByRole("button");
    expect(cardButton).toHaveAttribute("aria-pressed", "true");
    expect(cardButton).toHaveClass("selected");

    fireEvent.click(cardButton);
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ id: 7 }));
  });
});
