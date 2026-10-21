export function ErrorMessage({ message, label = "Settings unavailable" }: { message: string; label?: string }) {
  return <p role="alert" className="m-0 rounded-hm-16 bg-hm-accent-soft px-4 py-3 text-hm-body text-hm-ink">{label}: {message}</p>;
}
