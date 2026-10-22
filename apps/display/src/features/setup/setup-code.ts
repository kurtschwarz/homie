/** Split a setup code into the eight visible characters, excluding its separator. */
export function setupCodeCharacters(code: string): string[] {
  return code.replaceAll('-', '').slice(0, 8).split('')
}

/** Display the server's canonical XXXX-XXXX pairing format. */
export function formatSetupCode(code: string): string {
  const characters = setupCodeCharacters(code)
  return `${characters.slice(0, 4).join('')}-${characters.slice(4).join('')}`
}
