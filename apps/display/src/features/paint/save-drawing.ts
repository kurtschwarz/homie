const MUTATION = `
  mutation SaveDrawing($input: DrawingInput!, $file: Upload!) {
    saveDrawing(input: $input, file: $file) { id }
  }
`

/** Flattens the transparent canvas onto white and encodes it as JPEG. */
export async function canvasToJpeg(canvas: HTMLCanvasElement): Promise<Blob> {
  const flat = document.createElement('canvas')
  flat.width = canvas.width
  flat.height = canvas.height

  const ctx = flat.getContext('2d')
  if (!ctx) {
    throw new Error('Canvas is unavailable')
  }

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, flat.width, flat.height)
  ctx.drawImage(canvas, 0, 0)

  return new Promise((resolve, reject) => {
    flat.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('Could not encode the drawing'))),
      'image/jpeg',
      0.92,
    )
  })
}

/** Uploads a JPEG drawing using the GraphQL multipart request spec. */
export async function uploadDrawing(
  jpeg: Blob,
  size: { width: number; height: number },
  drawnAt: Date,
): Promise<void> {
  const form = new FormData()
  form.append(
    'operations',
    JSON.stringify({
      query: MUTATION,
      variables: {
        input: { ...size, drawnAt: Math.floor(drawnAt.getTime() / 1000) },
        file: null,
      },
    }),
  )
  form.append('map', JSON.stringify({ '0': ['variables.file'] }))
  form.append('0', new File([jpeg], 'drawing.jpg', { type: 'image/jpeg' }))

  const response = await fetch('/graphql', { method: 'POST', body: form })
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  const body = (await response.json()) as { errors?: Array<{ message?: string }> }
  if (body.errors?.length) throw new Error(body.errors[0]?.message ?? 'GraphQL error')
}
