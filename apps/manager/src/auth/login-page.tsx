import { Button } from "@homie/ui/components/button";
import * as React from "react";
import { ErrorMessage } from "../components/error-message";
import { Field } from "../components/field";
import { AuthCard } from "./auth-card";

export function LoginPage({
  onLogin,
}: {
  onLogin: (username: string, password: string) => Promise<void>;
}) {
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [error, setError] = React.useState<string>();
  const [submitting, setSubmitting] = React.useState(false);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(undefined);
    try {
      await onLogin(username.trim(), password);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setSubmitting(false);
    }
  };

  return (
    <AuthCard title="Log in" description="Sign in to manage Homie.">
      <form className="flex flex-col gap-4" onSubmit={submit}>
        {error && <ErrorMessage label="Login failed" message={error} />}
        <Field
          id="login-username"
          label="Username"
          value={username}
          onChange={setUsername}
          autoComplete="username"
        />
        <Field
          id="login-password"
          label="Password"
          type="password"
          value={password}
          onChange={setPassword}
          autoComplete="current-password"
        />
        <Button type="submit" variant="solid" block disabled={submitting}>
          Log in
        </Button>
      </form>
    </AuthCard>
  );
}
