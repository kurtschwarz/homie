/** Homie's face, static: the orange-to-pink ball with the lopsided grin. */
export function BrandMark() {
  return (
    <svg width="40" height="40" viewBox="0 0 100 100" aria-hidden="true" className="shrink-0">
      <defs>
        <linearGradient id="homie-brand-warm" x1=".15" y1=".08" x2=".9" y2=".95">
          <stop offset="0" stopColor="#D45F28" />
          <stop offset=".55" stopColor="#E06C3F" />
          <stop offset="1" stopColor="#EE8AA3" />
        </linearGradient>
        <radialGradient
          id="homie-brand-shade"
          cx="50"
          cy="50"
          r="60"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0" stopColor="white" stopOpacity=".16" />
          <stop offset=".8" stopColor="#1F0900" stopOpacity="0" />
          <stop offset=".93" stopColor="#1F0900" stopOpacity=".34" />
          <stop offset="1" stopColor="#1F0900" stopOpacity=".62" />
        </radialGradient>
        <clipPath id="homie-brand-clip">
          <circle cx="50" cy="50" r="48" />
        </clipPath>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(#homie-brand-warm)" />
      <g clipPath="url(#homie-brand-clip)">
        <rect x="-40" y="-40" width="180" height="180" fill="url(#homie-brand-shade)" />
      </g>
      <ellipse cx="53" cy="41" rx="4.6" ry="7.2" transform="rotate(-14 53 41)" fill="white" />
      <ellipse cx="72" cy="33" rx="4" ry="6.4" transform="rotate(-22 72 33)" fill="white" />
      <path
        d="M44 57 Q60 71 77 56"
        fill="none"
        stroke="white"
        strokeWidth="4.6"
        strokeLinecap="round"
      />
    </svg>
  )
}
