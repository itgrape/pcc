import type { AxiosResponse } from "axios"

interface ApiEnvelope<T> {
  data?: T
}

export function unwrapData<T>(response: AxiosResponse<unknown>): T {
  const envelope = response.data as ApiEnvelope<T>
  if (envelope?.data === undefined) {
    throw new Error("API response does not contain data")
  }
  return envelope.data
}

// Generated clients currently describe snowflake IDs as numbers, while the
// API serializes them as strings. Keeping the runtime value as a string avoids
// precision loss until the OpenAPI schema is regenerated with string IDs.
export function apiId(id: string): number {
  return id as unknown as number
}
