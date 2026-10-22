import * as React from "react";

/** Crockford base32 alphabet used by the API setup-code generator. */
const SETUP_CODE_ALPHABET = "0123456789ABCDEFGHJKMNPQRSTVWXYZ";

const CODE_LENGTH = 8;
const GROUP_SIZE = 4;

function normalizeCharacter(char: string) {
  const normalized = char.toUpperCase();
  if (normalized === "I" || normalized === "L") return "1";
  if (normalized === "O") return "0";
  return SETUP_CODE_ALPHABET.includes(normalized) ? normalized : "";
}

export function CodeInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const refs = React.useRef<Array<HTMLInputElement | null>>([]);
  const chars = Array.from({ length: CODE_LENGTH }, (_, index) => value[index] ?? "");

  const updateAt = (index: number, input: string) => {
    const accepted = Array.from(input).map(normalizeCharacter).filter(Boolean);
    if (accepted.length === 0) return;

    const next = [...chars];
    let position = index;
    for (const char of accepted) {
      if (position >= CODE_LENGTH) break;
      next[position] = char;
      position += 1;
    }
    onChange(next.join(""));
    refs.current[Math.min(position, CODE_LENGTH - 1)]?.focus();
  };

  const handlePaste = (event: React.ClipboardEvent<HTMLInputElement>, index: number) => {
    const pasted = event.clipboardData.getData("text").replace(/-/g, "");
    const accepted = Array.from(pasted).map(normalizeCharacter).filter(Boolean);
    if (accepted.length === 0) return;
    event.preventDefault();
    updateAt(index, accepted.join(""));
  };

  return (
    <fieldset className="m-0 flex flex-col gap-2 border-0 p-0">
      <legend className="mb-2 text-hm-label text-hm-ink-2">Setup code</legend>
      <div className="flex items-center gap-2">
        {chars.map((char, index) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: Each input represents a fixed character slot.
          <React.Fragment key={index}>
            {index === GROUP_SIZE && (
              <span aria-hidden="true" className="text-hm-title-sm text-hm-ink-3">
                -
              </span>
            )}
            <input
              ref={(element) => {
                refs.current[index] = element;
              }}
              aria-label={`Setup code character ${index + 1} of ${CODE_LENGTH}`}
              autoComplete="off"
              inputMode="text"
              maxLength={1}
              value={char}
              onChange={(event) => updateAt(index, event.target.value)}
              onPaste={(event) => handlePaste(event, index)}
              onKeyDown={(event) => {
                if (event.key === "Backspace" && char === "" && index > 0) {
                  event.preventDefault();
                  refs.current[index - 1]?.focus();
                  const next = [...chars];
                  next[index - 1] = "";
                  onChange(next.join(""));
                } else if (event.key === "ArrowLeft" && index > 0) {
                  event.preventDefault();
                  refs.current[index - 1]?.focus();
                } else if (event.key === "ArrowRight" && index < CODE_LENGTH - 1) {
                  event.preventDefault();
                  refs.current[index + 1]?.focus();
                }
              }}
              className="h-14 w-10 rounded-hm-control border border-hm-line bg-hm-surface text-center font-mono text-hm-title-sm text-hm-ink focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-hm-accent"
            />
          </React.Fragment>
        ))}
      </div>
    </fieldset>
  );
}
