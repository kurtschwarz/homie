/** Where the bearer token lives. The API log prints it on first boot. */
export const TOKEN_KEY = "homie-auth-token";

export async function graphql<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const token = localStorage.getItem(TOKEN_KEY);
  const response = await fetch("/graphql", {
    method: "POST",
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify({ query, variables }),
  });
  if (response.status === 401) throw new Error("API token missing or invalid");
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  const body = await response.json();
  if (body.errors?.length) throw new Error(body.errors[0]?.message ?? "GraphQL error");
  return body.data as T;
}
