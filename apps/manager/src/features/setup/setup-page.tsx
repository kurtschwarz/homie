import * as React from "react";
import { Lock } from "lucide-react";
import { HomieAnimation } from "@homie/ui/components/homie-animation";
import { Button } from "@homie/ui/components/button";
import { Callout } from "@homie/ui/components/callout";
import { CodeInput } from "../../components/code-input";
import { ErrorMessage } from "../../components/error-message";
import { Field } from "../../components/field";
import { SetupStep } from "./setup-step";
import homieAnimationUrl from "@homie/ui/assets/homie-confetti.lottie?url";

export function SetupPage({ onSetup }: { onSetup: (code: string, username: string, password: string) => Promise<void> }) {
  const [step, setStep] = React.useState<1 | 2>(1);
  const [code, setCode] = React.useState("");
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [error, setError] = React.useState<string>();
  const [submitting, setSubmitting] = React.useState(false);

  const continueToAccount = (event: React.FormEvent) => {
    event.preventDefault();
    if (code.length !== 8) return;
    setError(undefined);
    setStep(2);
  };

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (password !== confirm) {
      setError("Passwords don't match");
      return;
    }
    setSubmitting(true);
    setError(undefined);
    try {
      await onSetup(`${code.slice(0, 4)}-${code.slice(4)}`, username.trim(), password);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(message);
      if (/setup code/i.test(message)) setStep(1);
      setSubmitting(false);
    }
  };

  return <main className="flex min-h-dvh flex-col items-center justify-center gap-8 bg-hm-shell px-5 py-10 sm:gap-9 sm:px-8">
    <header className="flex flex-col items-center text-center">
      <HomieAnimation src={homieAnimationUrl} />
      <div className="-mt-14 flex flex-col gap-2 sm:-mt-[72px]">
        <h1 className="m-0 text-hm-display font-semibold text-hm-ink">Hi, I'm Homie!</h1>
        <p className="m-0 text-hm-body text-hm-ink-2">Let's get me set up. Two quick steps and you're in.</p>
      </div>
    </header>

    <div className="flex w-full max-w-2xl flex-col gap-3">
      <SetupStep number={1} title="Enter your setup code" description="Look at the Homie device screen. It's showing a setup code." state={step === 1 ? "active" : "done"} onOpen={() => { setError(undefined); setStep(1); }}>
        <form className="flex flex-col gap-4" onSubmit={continueToAccount}>
          {error && <ErrorMessage label="Setup failed" message={error} />}
          <CodeInput value={code} onChange={(value) => { setCode(value); setError(undefined); }} />
          <p className="m-0 max-w-lg text-hm-caption text-hm-ink-3">It only works once, and only someone standing at the screen can see it, so nobody else on your network can claim the admin.</p>
          <div className="flex justify-end">
            <Button type="submit" variant="solid" disabled={code.length !== 8}>Continue</Button>
          </div>
        </form>
      </SetupStep>

      <SetupStep number={2} title="Create your admin account" description="Pick the username and password you'll sign in with from now on." state={step === 2 ? "active" : "upcoming"} onOpen={() => setStep(2)}>
        <form className="flex flex-col gap-4" onSubmit={submit}>
          {error && <ErrorMessage label="Setup failed" message={error} />}
          <Field id="setup-username" label="Username" value={username} onChange={setUsername} autoComplete="username" />
          <Field id="setup-password" label="Password" type="password" value={password} onChange={setPassword} autoComplete="new-password" />
          <Field id="setup-confirm" label="Confirm password" type="password" value={confirm} onChange={setConfirm} autoComplete="new-password" />
          <Button type="submit" variant="solid" block disabled={submitting}>Create account</Button>
          <Callout lead="This account gates the whole network admin.">Anyone on the wifi can reach this page until it's set up.</Callout>
        </form>
      </SetupStep>
    </div>

    <footer className="flex items-center gap-2 text-hm-caption text-hm-ink-3">
      <Lock aria-hidden="true" className="size-4" />
      <span>Only works on your home network · homie.local</span>
    </footer>
  </main>;
}
