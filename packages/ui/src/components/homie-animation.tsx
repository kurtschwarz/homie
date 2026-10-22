"use client";

import * as React from "react";
import { DotLottieReact } from "@lottiefiles/dotlottie-react";
import { cn } from "../lib/utils";

/** The looping Homie mascot animation used by first-run setup screens. */
export function HomieAnimation({ src, className, ...props }: React.ComponentProps<"div"> & { src: string }) {
  return (
    <div data-slot="homie-animation" className={cn("size-56 sm:size-64", className)} role="img" aria-label="Homie" {...props}>
      <DotLottieReact
        src={src}
        autoplay
        loop
        renderConfig={{ devicePixelRatio: typeof window === "undefined" ? 1 : window.devicePixelRatio }}
      />
    </div>
  );
}
