import { useQuery } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { ErrorMessage } from '@/shared/ui/error-message'
import { householdQueries } from '../api/household.queries'
import type { Household } from '../types'

/**
 * Renders its child only once household data has loaded. A failed background
 * refetch keeps the last good data on screen and shows the error above it.
 * Only a first load with nothing cached gives the error the whole page.
 */
export function HouseholdGate({ children }: { children: (household: Household) => ReactNode }) {
  const query = useQuery(householdQueries.settings())

  if (query.isPending) return <p className="text-hm-body text-hm-ink-2">Loading settings…</p>
  if (!query.data) return <ErrorMessage message={query.error?.message ?? 'Settings unavailable'} />

  return (
    <>
      {query.error && <ErrorMessage message={query.error.message} />}
      {children(query.data)}
    </>
  )
}
