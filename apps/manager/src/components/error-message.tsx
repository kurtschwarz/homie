export function ErrorMessage({ message }: { message: string }) {
  return <p role="alert" className="m-0 rounded-hm-16 bg-hm-accent-soft px-4 py-3 text-hm-body text-hm-ink">Settings unavailable: {message}</p>;
}
