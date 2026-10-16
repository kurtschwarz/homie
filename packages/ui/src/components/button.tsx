"use client"

import { Button as BaseButton } from "@base-ui/react/button"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "../lib/utils"

const buttonVariants = cva(
  [
    "inline-flex shrink-0 cursor-pointer items-center justify-center whitespace-nowrap font-sans select-none",
    "[-webkit-tap-highlight-color:transparent] transition-transform duration-120 ease-out active:scale-(--hm-press-scale)",
    "focus-visible:outline-3 focus-visible:outline-offset-3 focus-visible:outline-hm-accent",
    "data-disabled:pointer-events-none data-disabled:opacity-50 [&_svg]:shrink-0",
  ],
  {
    variants: {
      variant: {
        /** Outlined button for the one way back or forward on a screen. `button` tall: 88 / 44 / 48px. */
        outline:
          "h-hm-button gap-hm-10 rounded-hm-control border-2 border-hm-line bg-hm-surface pr-hm-32 pl-hm-24 text-hm-label-lg text-hm-ink [&_svg]:size-hm-28",
        /** Text button for a jump beside a section label (Week ›). `touch` tall: 64 / 40 / 44px. */
        text: "h-hm-touch gap-hm-6 rounded-hm-control bg-transparent pr-hm-12 pl-hm-20 text-hm-label text-hm-ink-2 [&_svg]:size-hm-26",
      },
      block: { true: "w-full px-hm-24", false: "" },
    },
    defaultVariants: { variant: "outline", block: false },
  },
)

type ButtonProps = BaseButton.Props & VariantProps<typeof buttonVariants>

/**
 * Quiet button built on Base UI's Button. Put a lucide chevron inside: left and before the label for back, right
 * and after it for forward. Nothing depends on hover; pressing shrinks it to `press-scale`.
 * Use `render` to compose it with a link (set `nativeButton={false}`).
 */
function Button({ className, variant, block, ...props }: ButtonProps) {
  return <BaseButton data-slot="button" className={cn(buttonVariants({ variant, block }), className)} {...props} />
}

export { Button, buttonVariants }
export type { ButtonProps }
