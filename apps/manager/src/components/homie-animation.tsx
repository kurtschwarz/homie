import { DotLottieReact } from "@lottiefiles/dotlottie-react";

export function HomieAnimation() {
  return <div className="size-56 sm:size-64" role="img" aria-label="Homie">
    <DotLottieReact
      src={`${import.meta.env.BASE_URL}homie-confetti.lottie`}
      autoplay
      loop
      renderConfig={{ devicePixelRatio: typeof window === "undefined" ? 1 : window.devicePixelRatio }}
    />
  </div>;
}
