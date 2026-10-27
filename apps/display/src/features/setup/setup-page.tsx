import homieAnimationUrl from '@homie/ui/assets/homie-confetti.lottie?url'
import { HomieAnimation } from '@homie/ui/components/homie-animation'
import { ScanQrCode, Smartphone, Wifi } from 'lucide-react'
import QRCode from 'react-qr-code'
import { setupCodeCharacters } from './setup-code'

const MANAGE_URL = 'homie.local'

export function SetupPage({ code }: { code: string }) {
  const characters = setupCodeCharacters(code)
  const qrValue = `http://${MANAGE_URL}?code=${characters.join('')}`

  return (
    <main className="fixed inset-0 flex h-dvh w-screen flex-col items-center justify-center gap-hm-52 bg-hm-shell px-hm-screen py-hm-screen text-hm-ink">
      <header className="flex items-center gap-hm-24">
        <div className="flex size-hm-76 shrink-0 items-center justify-center overflow-hidden rounded-full bg-hm-accent-soft">
          <HomieAnimation src={homieAnimationUrl} className="size-hm-64" />
        </div>
        <div className="flex flex-col gap-hm-4">
          <h1 className="m-0 text-hm-title-lg font-semibold leading-tight tracking-[-0.035em]">
            Hi, I’m Homie!
          </h1>
          <p className="m-0 text-hm-lead text-hm-ink-2">
            Pair your phone to get me set up. Pick whichever’s easier.
          </p>
        </div>
      </header>

      <div className="flex w-full max-w-[90rem] items-stretch justify-center gap-hm-32">
        <section className="flex w-[32rem] flex-col items-center gap-hm-22 rounded-hm-28 border border-hm-line bg-hm-surface px-hm-36 py-hm-32 shadow-hm-float">
          <div className="flex w-full items-center gap-hm-12">
            <ScanQrCode aria-hidden="true" className="size-hm-26 shrink-0 text-hm-accent" />
            <h2 className="m-0 text-hm-lead font-semibold whitespace-nowrap">
              Scan with your camera
            </h2>
          </div>
          <QRCode
            aria-label={`QR code to ${MANAGE_URL}`}
            value={qrValue}
            className="size-[14rem]"
          />
          <p className="m-0 text-center text-hm-caption text-hm-ink-3">
            Opens setup with the code filled in
          </p>
        </section>

        <div aria-hidden="true" className="flex flex-col items-center gap-hm-12 py-hm-24">
          <span className="w-px flex-1 bg-hm-line" />
          <span className="text-hm-lead text-hm-ink-3">or</span>
          <span className="w-px flex-1 bg-hm-line" />
        </div>

        <section className="flex w-[38rem] flex-col gap-hm-26 rounded-hm-28 border border-hm-line bg-hm-surface px-hm-36 py-hm-32 shadow-hm-float">
          <div className="flex items-center gap-hm-12">
            <Smartphone aria-hidden="true" className="size-hm-26 shrink-0 text-hm-accent" />
            <h2 className="m-0 text-hm-lead font-semibold whitespace-nowrap">Type it in</h2>
          </div>

          <div className="flex flex-col gap-hm-8">
            <p className="m-0 text-hm-lead text-hm-ink-2">1. In any browser, go to</p>
            <strong className="whitespace-nowrap text-hm-title font-semibold leading-tight tracking-[-0.03em]">
              {MANAGE_URL}
            </strong>
          </div>

          <div className="flex flex-col gap-hm-12">
            <p className="m-0 text-hm-lead text-hm-ink-2">2. Enter this code</p>
            <fieldset className="m-0 flex items-center gap-hm-10 border-0 p-0">
              <legend className="sr-only">Setup code {code}</legend>
              {characters
                .map((character, index) => ({
                  character,
                  key: `${character}-${characters.slice(0, index).filter((previous) => previous === character).length}`,
                }))
                .map((slot) => (
                  <div
                    className="flex size-hm-52 min-w-0 flex-1 items-center justify-center rounded-hm-16 bg-hm-shell text-hm-title font-semibold"
                    key={slot.key}
                  >
                    {slot.character}
                  </div>
                ))
                .flatMap((cell, index) =>
                  index === 3
                    ? [
                        cell,
                        <span
                          aria-hidden="true"
                          className="text-hm-title-sm text-hm-ink-3"
                          key="separator"
                        >
                          –
                        </span>,
                      ]
                    : [cell],
                )}
            </fieldset>
          </div>
        </section>
      </div>

      <footer className="flex items-center gap-hm-10 pb-hm-8 text-hm-lead text-hm-ink-3">
        <Wifi aria-hidden="true" className="size-hm-20" />
        <span>Connected to Wi-Fi</span>
        <span aria-hidden="true">·</span>
        <span aria-hidden="true" className="size-hm-12 rounded-full bg-hm-accent" />
        <span>Waiting for you to pair</span>
      </footer>
    </main>
  )
}
