import { useCallback, useEffect, useState } from 'react'
import { graphql } from './graphql'
import type { HealthStatus } from './use-health'

export type User = { id: string; username: string; role: 'ADMIN' }

type SessionState =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'ready'; setupRequired: boolean; user: User | null }

const SESSION_QUERY = `{ session { user { id username role } } health { status } }`

/**
 * Whether setup or login is needed, polled once and refreshed after every
 * setup, login or logout. Failures from setup/login are left to the
 * caller: a wrong code or password is a form error, not a reason to tear
 * down the whole gate. Health status is the source of truth for whether
 * first-run setup is still required.
 */
export function useSession() {
  const [state, setState] = useState<SessionState>({ kind: 'loading' })

  const load = useCallback(async () => {
    try {
      const data = await graphql<{
        session: { user: User } | null
        health: { status: HealthStatus }
      }>(SESSION_QUERY)
      setState({
        kind: 'ready',
        setupRequired: data.health.status === 'SETUP_REQUIRED',
        user: data.session?.user ?? null,
      })
    } catch (err) {
      setState({ kind: 'error', message: err instanceof Error ? err.message : String(err) })
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const setup = useCallback(
    async (code: string, username: string, password: string) => {
      await graphql(
        `mutation($code: String!, $username: String!, $password: String!) { setup(code: $code, username: $username, password: $password) { id } }`,
        { code, username, password },
      )
      await load()
    },
    [load],
  )

  const login = useCallback(
    async (username: string, password: string) => {
      await graphql(
        `mutation($username: String!, $password: String!) { login(username: $username, password: $password) { id } }`,
        {
          username,
          password,
        },
      )
      await load()
    },
    [load],
  )

  const logout = useCallback(async () => {
    await graphql(`mutation { logout }`)
    await load()
  }, [load])

  return { state, setup, login, logout }
}
