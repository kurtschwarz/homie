import type { ActiveTimer } from "./useTimer";

type Props = {
  timer: ActiveTimer;
  addMinute: () => Promise<void>;
  togglePause: () => Promise<void>;
  cancel: () => Promise<void>;
};

export function TimerScreen({ timer, addMinute, togglePause, cancel }: Props) {
  const remaining = timer.paused
    ? timer.remainingSeconds
    : Math.max(0, timer.endsAt - Math.floor(Date.now() / 1000));
  const minutes = Math.floor(remaining / 60).toString().padStart(2, "0");
  const seconds = (remaining % 60).toString().padStart(2, "0");
  const progress = Math.min(100, Math.max(0, (remaining / timer.durationSeconds) * 100));
  const readyAt = new Date(timer.endsAt * 1000).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });

  return (
    <main className="timer-frame">
      <header className="timer-header">
        <div className="timer-title"><span>{timer.emoji}</span><strong>{timer.title}</strong></div>
        <time>{new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}</time>
      </header>
      <section className="timer-center" aria-live="off">
        <div className="timer-countdown">{minutes}:{seconds}</div>
        <div className="timer-ready">{timer.paused ? "Paused" : `Ready at ${readyAt}`}</div>
      </section>
      <footer className="timer-footer">
        <div className="timer-progress" role="progressbar" aria-valuenow={Math.round(progress)} aria-valuemin={0} aria-valuemax={100}>
          <span style={{ width: `${progress}%` }} />
        </div>
        <div className="timer-actions">
          <button onClick={() => void addMinute()}>+1 min</button>
          <button className="primary" onClick={() => void togglePause()}>{timer.paused ? "Resume" : "Pause"}</button>
          <button onClick={() => void cancel()}>Cancel</button>
        </div>
      </footer>
    </main>
  );
}
