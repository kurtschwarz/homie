/** Hand-rolled: packages/ui has no form Input, and the manager owns these fields. */
export function Field({ id, label, type = "text", value, onChange, autoComplete, placeholder }: {
  id: string;
  label: string;
  type?: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete?: string;
  placeholder?: string;
}) {
  return <div className="flex flex-col gap-2">
    <label htmlFor={id} className="text-hm-label text-hm-ink-2">{label}</label>
    <input id={id} type={type} autoComplete={autoComplete} placeholder={placeholder} value={value}
      onChange={(event) => onChange(event.target.value)}
      className="h-11 w-full rounded-hm-control border border-hm-line bg-hm-surface px-3 text-hm-body text-hm-ink focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-hm-accent" />
  </div>;
}
