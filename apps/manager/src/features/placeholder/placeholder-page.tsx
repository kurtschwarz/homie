import { Page } from '../../shell/page'

/** A section that has a home in the sidebar but nothing behind it yet. */
export function PlaceholderPage({ title, description }: { title: string; description: string }) {
  return (
    <Page title={title} description={description}>
      <p className="m-0 self-start rounded-hm-16 border border-hm-line bg-hm-bg px-5 py-4 text-hm-body text-hm-ink-2">
        This section is not available yet.
      </p>
    </Page>
  )
}
