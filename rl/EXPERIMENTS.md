# RootBuddy RL Experiments

Phase 5.3 uses repeatable train/eval presets instead of one-off commands.

The RL env currently runs with full internal hand tracking so legal action generation can see every faction's real cards. Observations are still encoded from the active faction's perspective; true hidden-hand action generation is a later engine feature.

Two-player training uses a small per-step penalty and a truncation penalty so legal but unproductive loops are not neutral. Use at least `1024` max steps for meaningful two-player evals; shorter caps are only for smoke checks.

## Starter Presets

- `smoke`: verifies the full training, checkpoint, league, and eval path quickly.
- `two-player-fast`: short Marquise vs Eyrie runs for comparing reward and PPO changes.
- `two-player-stable`: first longer two-player recipe before moving to four-player training.

Run the smoke preset:

```sh
python -m rootbuddy_rl.experiment --preset smoke
```

Run a quick tuning comparison:

```sh
python -m rootbuddy_rl.experiment --preset two-player-fast --entropy-coef 0.02 --terminal-win-bonus 30
```

## Metrics To Watch

- `entropy`: low entropy means exploration may be collapsing.
- `approx_kl`: high KL means PPO updates are too aggressive.
- `truncated`: high truncation means games are stalling or `max_steps` is too low.
- `win_rate`: large faction gaps suggest imbalance or degenerate play.
- `terminal_vp`: helps distinguish real wins from empty or stalled games.

## Initial Tuning Knobs

- Lower `--learning-rate` or `--ppo-epochs` if KL spikes.
- Raise `--entropy-coef` if entropy collapses too early.
- Raise `--max-steps` if many eval games truncate.
- Raise `--truncation-penalty` or `--step-penalty` if diagnostics show repeated late-game pass/move cycles.
- Adjust `--terminal-win-bonus` if agents farm immediate VP but fail to close games.
- Adjust `--league-opponent-fraction` if training overfits to the latest self.
