/**
 * Executes a GraphQL query against the API and returns its data.
 *
 * @param query - The GraphQL document to send.
 * @returns The `data` field of the response.
 * @throws If the request fails, the response contains errors, or `data` is missing.
 */
export async function graphql<T>(query: string): Promise<T> {
  const response = await fetch("/graphql", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  const body = (await response.json()) as {
    data?: T;
    errors?: Array<{ message?: string }>;
  };

  if (body.errors?.length) {
    throw new Error(body.errors[0]?.message ?? "GraphQL error");
  }

  if (!body.data) {
    throw new Error("GraphQL response did not include data");
  }

  return body.data;
}
