import { Button } from '@homie/ui/components/button'
import * as React from 'react'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Field } from '@/shared/ui/field'
import { useLogin } from '../api/auth.mutations'
import { AuthCard } from './auth-card'

export function LoginPage() {
  const [username, setUsername] = React.useState('')
  const [password, setPassword] = React.useState('')
  const login = useLogin()

  // No need to clear `isPending` on success: the gate replaces this page as
  // soon as the refetched session names a user.
  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    login.mutate({ username: username.trim(), password })
  }

  return (
    <AuthCard title="Log in" description="Sign in to manage Homie.">
      <form className="flex flex-col gap-4" onSubmit={submit}>
        {login.error && <ErrorMessage label="Login failed" message={login.error.message} />}
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
        <Button type="submit" variant="solid" block disabled={login.isPending}>
          Log in
        </Button>
      </form>
    </AuthCard>
  )
}
