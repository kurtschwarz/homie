import { Button } from '@homie/ui/components/button'
import { KeyRound } from 'lucide-react'
import * as React from 'react'

/** Reads a .p8 in the browser and hands up its PEM text; the file itself never leaves the page. */
export function PrivateKeyField({
  stored,
  fileName,
  onChange,
}: {
  stored: boolean
  fileName: string
  onChange: (pem: string, fileName: string) => void
}) {
  const input = React.useRef<HTMLInputElement>(null)
  const [error, setError] = React.useState('')

  async function read(file: File | undefined) {
    if (!file) return
    const pem = await file.text()
    if (!pem.includes('BEGIN PRIVATE KEY')) {
      setError('That does not look like an Apple .p8 private key.')
      return
    }
    setError('')
    onChange(pem, file.name)
  }

  const label = fileName
    ? fileName
    : stored
      ? 'Private key stored on the Pi'
      : 'No key uploaded yet'
  const action = fileName || stored ? 'Replace' : 'Choose file'

  return (
    <div className="flex flex-col gap-2">
      <span id="weatherkit-key-label" className="text-hm-label text-hm-ink-2">
        Private key (.p8)
      </span>
      <div className="flex items-center gap-3 rounded-hm-control border border-hm-line bg-hm-bg px-3 py-2">
        <KeyRound aria-hidden="true" className="size-4 text-hm-ink-2" />
        <span className="min-w-0 flex-1 truncate text-hm-body">{label}</span>
        <Button
          variant="outline"
          className="h-10 border px-3 text-hm-caption"
          aria-describedby="weatherkit-key-label"
          onClick={() => input.current?.click()}
        >
          {action}
        </Button>
        <input
          ref={input}
          type="file"
          accept=".p8"
          hidden
          onChange={(event) => {
            void read(event.target.files?.[0])
            event.target.value = ''
          }}
        />
      </div>
      {error && (
        <p role="alert" className="m-0 text-hm-caption text-red-700">
          {error}
        </p>
      )}
    </div>
  )
}
