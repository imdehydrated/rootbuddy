type GuideHelpPanelProps = {
  gamePhase: number;
  onClose: () => void;
};

export function GuideHelpPanel({ gamePhase, onClose }: GuideHelpPanelProps) {
  return (
    <section className="panel sidebar-panel">
      <div className="panel-header">
        <h2>{gamePhase === 2 ? "Review Guide" : "Quick Start"}</h2>
        <button type="button" className="secondary" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="compact-help">
        {gamePhase === 2 ? (
          <>
            <p>1. Use the Game Over panel to see how the win happened and review the final standings.</p>
            <p>2. Use Return to Setup if you want to keep this result resumable, or Clear Saved Result if you want to discard it.</p>
            <p>3. Use New Game only when you are ready to replace the finished match with a fresh setup.</p>
          </>
        ) : (
          <>
            <p>1. Click a clearing to select it, then use the Board Editor in the sidebar to place warriors, buildings, sympathy, wood, ruins, the Keep, and the Vagabond.</p>
            <p>2. Use Turn Controls in the sidebar to set the acting faction and phase, and open Advanced Turn only when you need a manual correction.</p>
            <p>3. Load or refresh actions from the Flow Guide, then apply them from Player Turn, Observed Turn, or Battle Flow.</p>
          </>
        )}
      </div>
    </section>
  );
}
