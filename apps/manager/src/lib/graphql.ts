export async function graphql<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const response = await fetch("/graphql", {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query, variables }),
  });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  const body = await response.json();
  if (body.errors?.length) {
    throw new Error(body.errors[0]?.message ?? "GraphQL error");
  }

  return body.data as T;
}
