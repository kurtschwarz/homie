import type { ReactNode } from "react";
import { BrandMark } from "../shell/brand-mark";

/** The centered card setup and login share: there is no sidebar yet to sit beside. */
export function AuthCard({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return <main className="flex min-h-dvh items-center justify-center bg-hm-shell p-6">
    <div className="flex w-full max-w-md flex-col gap-6 rounded-[20px] bg-hm-surface p-8 shadow-hm-panel">
      <div className="flex items-center gap-3">
        <BrandMark />
        <span className="text-[22px] font-semibold leading-tight tracking-[-0.03em] text-hm-ink">Homie</span>
      </div>
      <div className="flex flex-col gap-2">
        <h1 className="m-0 text-hm-title-sm font-semibold text-hm-ink">{title}</h1>
        <p className="m-0 text-hm-body text-hm-ink-2">{description}</p>
      </div>
      {children}
    </div>
  </main>;
}
