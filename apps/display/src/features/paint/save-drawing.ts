const MUTATION = `
  mutation SaveDrawing($input: DrawingInput!, $file: Upload!) {
    saveDrawing(input: $input, file: $file) { id }
  }
`

/**
 * Flattens the transparent canvas onto a white background and encodes it as
 * JPEG. JPEG has no alpha channel, so without the white fill the transparent
 * pixels would encode as black. The source canvas is left untouched.
 *
 * @param canvas - The paint canvas to encode.
 * @returns The drawing as a JPEG blob at the canvas's backing resolution.
 * @throws If a 2D context is unavailable or the browser cannot encode the image.
 */
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
      1,
    )
  })
}

/**
 * Uploads a JPEG drawing to the API's `saveDrawing` mutation using the GraphQL
 * multipart request spec. This bypasses the shared `graphql` helper, which
 * only sends JSON.
 *
 * @param jpeg - The encoded drawing, as returned by `canvasToJpeg`.
 * @param size - Pixel dimensions of the encoded image.
 * @param drawnAt - When the drawing was finished; the API stores it as the
 * photo's taken-at time.
 * @param artistId - The person who drew it, if one was chosen.
 * @throws If the request fails, the response is not OK, or it contains errors.
 */
export async function uploadDrawing(
  jpeg: Blob,
  size: { width: number; height: number },
  drawnAt: Date,
  artistId?: string,
): Promise<void> {
  const form = new FormData()
  form.append(
    'operations',
    JSON.stringify({
      query: MUTATION,
      variables: {
        input: {
          ...size,
          drawnAt: Math.floor(drawnAt.getTime() / 1000),
          artistId: artistId ?? null,
        },
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
