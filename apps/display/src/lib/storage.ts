/**
 * The slice of the Web Storage API these helpers use. `localStorage`
 * satisfies it, and tests can pass a plain in-memory object instead.
 */
export type Store = Pick<Storage, 'getItem' | 'setItem'>

/**
 * Reads a string from storage without ever throwing.
 *
 * `undefined` means "use `localStorage`"; an explicit `null` means "no
 * storage". The bare `localStorage` reference stays inside the `try`
 * because it throws when storage is blocked or missing.
 *
 * @param key - The key to read.
 * @param store - Storage to read from. Omit it to use `localStorage`, or
 * pass `null` to behave as if no storage exists.
 * @returns The stored string, or `null` when the key is unset, storage is
 * absent, or access throws.
 */
export function read(key: string, store: Store | null | undefined): string | null {
  try {
    return (store === undefined ? localStorage : store)?.getItem(key) ?? null
  } catch {
    return null
  }
}

/**
 * Writes a string to storage without ever throwing. Failures, such as a
 * full or blocked store, are ignored.
 *
 * @param key - The key to write.
 * @param value - The string to store.
 * @param store - Storage to write to. Omit it to use `localStorage`, or
 * pass `null` to skip the write.
 */
export function write(key: string, value: string, store: Store | null | undefined) {
  try {
    ;(store === undefined ? localStorage : store)?.setItem(key, value)
  } catch {}
}
