import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { graphql } from '../../lib/graphql'
import { pollEvery } from '../../lib/query-client'
import { formatSetupCode } from './setup-code'

type SetupState =
  | { kind: 'loading' }
  | { kind: 'setup'; code: string }
  | { kind: 'ready' }
  | { kind: 'error' }

type SetupResponse = {
  health: { status: 'OK' | 'SETUP_REQUIRED' }
  setupCode: string | null
}

const QUERY = `{ health { status } setupCode }`

async function fetchSetup() {
  const data = await graphql<SetupResponse>(QUERY)
  if (!data.health) throw new Error('Setup status query failed')
  return data
}

/** Polls first-run status and retains a visible code through transient failures. */
export function useSetup(): SetupState {
  const query = useQuery({
    queryKey: ['setup'],
    queryFn: fetchSetup,
    refetchInterval: pollEvery(5000),
  })

  // A successful response can report SETUP_REQUIRED without a code (the code
  // hasn't been (re)generated yet); the setup page should stay on the last
  // code it showed rather than unmount to a blank screen.
  const lastCode = useRef<string | null>(null)
  const freshCode =
    query.data?.health.status === 'SETUP_REQUIRED' && query.data.setupCode
      ? formatSetupCode(query.data.setupCode)
      : null

  useEffect(() => {
    if (freshCode) lastCode.current = freshCode
  }, [freshCode])

  if (query.data?.health.status === 'OK') return { kind: 'ready' }
  if (freshCode) return { kind: 'setup', code: freshCode }
  if (lastCode.current) return { kind: 'setup', code: lastCode.current }
  if (query.isError) return { kind: 'error' }
  return { kind: 'loading' }
}
