import balloon from '@homie/ui/assets/stickers/balloon.svg?url'
import cloud from '@homie/ui/assets/stickers/cloud.svg?url'
import crown from '@homie/ui/assets/stickers/crown.svg?url'
import flower from '@homie/ui/assets/stickers/flower.svg?url'
import heart from '@homie/ui/assets/stickers/heart.svg?url'
import homie from '@homie/ui/assets/stickers/homie.svg?url'
import lightning from '@homie/ui/assets/stickers/lightning.svg?url'
import moon from '@homie/ui/assets/stickers/moon.svg?url'
import rainbow from '@homie/ui/assets/stickers/rainbow.svg?url'
import smiley from '@homie/ui/assets/stickers/smiley.svg?url'
import star from '@homie/ui/assets/stickers/star.svg?url'
import sun from '@homie/ui/assets/stickers/sun.svg?url'
import * as React from 'react'

export type StickerArt = { id: string; name: string; url: string }

export type Sticker = StickerArt & { image: HTMLImageElement }

export const STICKER_ART: StickerArt[] = [
  { id: 'star', name: 'Star', url: star },
  { id: 'heart', name: 'Heart', url: heart },
  { id: 'sun', name: 'Sun', url: sun },
  { id: 'rainbow', name: 'Rainbow', url: rainbow },
  { id: 'flower', name: 'Flower', url: flower },
  { id: 'smiley', name: 'Smiley', url: smiley },
  { id: 'homie', name: 'Homie', url: homie },
  { id: 'cloud', name: 'Cloud', url: cloud },
  { id: 'moon', name: 'Moon', url: moon },
  { id: 'lightning', name: 'Lightning', url: lightning },
  { id: 'balloon', name: 'Balloon', url: balloon },
  { id: 'crown', name: 'Crown', url: crown },
]

export const DEFAULT_STICKER_ID = 'star'

/** Sticker edge length in CSS px for each size tile. */
export const STICKER_PX = { small: 90, medium: 160, large: 260 }

async function decode({ url, ...art }: StickerArt): Promise<Sticker | null> {
  const image = new Image()
  image.src = url
  try {
    await image.decode()
    return { ...art, url, image }
  } catch {
    return null
  }
}

/** Only stickers that decoded are returned, because history replays records synchronously. */
export function useStickers() {
  const [stickers, setStickers] = React.useState<Sticker[]>([])

  React.useEffect(() => {
    let cancelled = false
    Promise.all(STICKER_ART.map(decode)).then((loaded) => {
      if (!cancelled) setStickers(loaded.filter((sticker) => sticker !== null))
    })
    return () => {
      cancelled = true
    }
  }, [])

  return stickers
}
