import { Globe2 } from "lucide-react";
import { HomieAnimation } from "@homie/ui/components/homie-animation";
import homieAnimationUrl from "@homie/ui/assets/homie-confetti.lottie?url";
import { setupCodeCharacters } from "./setup-code";

export function SetupPage({ code }: { code: string }) {
  const characters = setupCodeCharacters(code);

  return (
    <main className="fixed inset-0 flex h-dvh w-screen flex-col items-center justify-center bg-hm-shell px-hm-screen py-hm-screen text-hm-ink">
      <div className="flex w-full max-w-[102.5rem] flex-1 items-center justify-center gap-hm-64">
        <div className="flex w-[38%] items-center justify-center">
          <HomieAnimation src={homieAnimationUrl} className="size-[38.75rem]" />
        </div>
        <section className="flex w-[58%] flex-col justify-center gap-hm-52">
          <header className="flex flex-col gap-hm-12">
            <p className="m-0 text-hm-eyebrow-lg font-semibold uppercase tracking-[0.14em] text-hm-accent">Let’s get set up</p>
            <h1 className="m-0 text-hm-headline font-semibold leading-none tracking-[-0.045em]">Hi, I’m Homie!</h1>
          </header>

          <div className="flex items-center gap-hm-26 rounded-hm-28 border border-hm-line bg-hm-surface px-hm-36 py-hm-30 shadow-hm-float">
            <div className="flex size-hm-76 shrink-0 items-center justify-center rounded-hm-20 bg-hm-accent-soft text-hm-accent">
              <Globe2 aria-hidden="true" className="size-hm-36" />
            </div>
            <div className="flex min-w-0 flex-col gap-hm-4">
              <span className="text-hm-lead text-hm-ink-2">On your phone or computer, visit</span>
              <strong className="break-all text-hm-title-lg font-semibold leading-tight tracking-[-0.035em]">homie.local/manage</strong>
            </div>
          </div>

          <div className="flex flex-col gap-hm-18">
            <p className="m-0 text-hm-lead text-hm-ink-2">Then enter this code to pair me</p>
            <div className="flex items-center gap-hm-12" aria-label={`Setup code ${code}`}>
              {characters.map((character, index) => (
                <div className="flex h-[9.5rem] min-w-0 flex-1 items-center justify-center rounded-hm-22 border border-hm-line bg-hm-surface text-hm-headline font-semibold shadow-hm-raised" key={`${index}-${character}`}>
                  {character}
                </div>
              )).flatMap((cell, index) => index === 3 ? [cell, <span aria-hidden="true" className="text-hm-title text-hm-accent" key="separator">—</span>] : [cell])}
            </div>
          </div>
        </section>
      </div>

      <footer className="flex items-center gap-hm-12 pb-hm-8 text-hm-lead text-hm-ink-3">
        <span aria-hidden="true" className="size-hm-12 rounded-full bg-hm-accent" />
        <span>Waiting for you to pair</span>
      </footer>
    </main>
  );
}
