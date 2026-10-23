import type { ReactNode } from 'react'

/**
 * A labelled group of related controls, such as the swatches that make up one
 * colour choice. It is a `fieldset` with a `legend`, so assistive technology
 * announces the label once for the whole group instead of leaving each
 * control to explain itself. The browser's default fieldset border and
 * padding are removed; the caller owns spacing between sections.
 *
 * For a single input with its own label, use `Field` instead.
 */
export function FormSection({
  label,
  children,
}: {
  /** The visible heading, also read out as the group's name. */
  label: string
  /** The controls in the group. */
  children: ReactNode
}) {
  return (
    <fieldset className="m-0 border-0 p-0">
      <legend className="mb-2 text-hm-caption font-semibold text-hm-ink-2">{label}</legend>
      {children}
    </fieldset>
  )
}
