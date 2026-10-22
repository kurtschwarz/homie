import { Link, NavLink } from "react-router-dom";
import { useAppData } from "../app-data";
import { BrandMark } from "./brand-mark";
import { sections } from "./sections";

export function Sidebar() {
  const { state, health } = useAppData();
  const peopleCount = state.kind === "ready" ? state.settings.people.length : undefined;

  return (
    <aside className="flex shrink-0 flex-col gap-4 md:w-[264px] md:gap-10 md:px-[18px] md:pb-[22px] md:pt-[30px]">
      <Link to="/people" className="flex items-center gap-3 px-2 text-inherit no-underline">
        <BrandMark />
        <span className="flex flex-col">
          <span className="text-[22px] font-semibold leading-tight tracking-[-0.03em]">Homie</span>
          <span className="text-hm-caption text-hm-ink-3">Admin</span>
        </span>
      </Link>

      <nav
        aria-label="Admin sections"
        className="flex gap-1 overflow-x-auto md:flex-col md:overflow-visible"
      >
        {sections.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end
            className={({ isActive }) =>
              `flex h-11 shrink-0 items-center gap-3 rounded-xl px-3 text-left text-[15px] no-underline transition-colors focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-hm-accent ${isActive ? "bg-hm-surface font-semibold text-hm-ink shadow-hm-raised" : "text-hm-ink-2 hover:bg-hm-surface/70"}`
            }
          >
            {({ isActive }) => (
              <>
                <Icon
                  aria-hidden="true"
                  size={20}
                  className={isActive ? "text-hm-accent" : "text-hm-ink-3"}
                />
                <span className="md:flex-1">{label}</span>
                {label === "People" && peopleCount !== undefined && (
                  <span className="grid min-w-7 place-items-center rounded-lg bg-hm-accent px-2 text-xs font-semibold leading-6 text-hm-on-accent">
                    {peopleCount}
                  </span>
                )}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div
        className="hidden items-center gap-2 px-2 text-hm-caption text-hm-ink-3 md:mt-auto md:flex"
        aria-live="polite"
      >
        <span
          className={`size-2 rounded-full ${health.kind === "ok" ? "bg-hm-accent" : "bg-hm-ink-3"}`}
        />
        {health.kind === "connecting" && "Connecting to API…"}
        {health.kind === "error" && "API unavailable"}
        {health.kind === "ok" && `API ${health.health.status}`}
      </div>
    </aside>
  );
}
