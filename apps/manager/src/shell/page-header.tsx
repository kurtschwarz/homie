import type { ReactNode } from "react";

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string;
  description: string;
  actions?: ReactNode;
}) {
  return (
    <header className="flex shrink-0 flex-col gap-4 border-b border-hm-line px-6 pb-5 pt-6 sm:flex-row sm:items-start sm:justify-between sm:gap-6 sm:px-9 sm:pb-6 sm:pt-[30px]">
      <div className="flex min-w-0 flex-col gap-2">
        <h1 className="m-0 text-hm-display font-semibold leading-[1.05] tracking-[-0.035em]">
          {title}
        </h1>
        <p className="m-0 text-hm-body text-hm-ink-2">{description}</p>
      </div>
      {actions && <div className="flex shrink-0 items-center gap-3 sm:pt-1">{actions}</div>}
    </header>
  );
}
