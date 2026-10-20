import type { ReactNode } from "react";
import { PageHeader } from "./page-header";

/** A section of the admin: fixed header over a body that scrolls on its own. */
export function Page({ title, description, actions, children }: {
  title: string;
  description: string;
  actions?: ReactNode;
  children: ReactNode;
}) {
  return <>
    <PageHeader title={title} description={description} actions={actions} />
    <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-auto px-6 py-7 sm:px-9">{children}</div>
  </>;
}
