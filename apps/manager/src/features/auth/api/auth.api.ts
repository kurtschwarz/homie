import { graphql } from '@/shared/lib/graphql'
import type { Credentials, SessionResponse, SetupInput } from '../types'

const SESSION_QUERY = `{ session { user { id username role } } health { status } }`

export const authApi = {
  session: () => graphql<SessionResponse>(SESSION_QUERY),

  setup: (input: SetupInput) =>
    graphql(
      `mutation($code: String!, $username: String!, $password: String!) { setup(code: $code, username: $username, password: $password) { id } }`,
      input,
    ),

  login: (input: Credentials) =>
    graphql(
      `mutation($username: String!, $password: String!) { login(username: $username, password: $password) { id } }`,
      input,
    ),

  logout: () => graphql(`mutation { logout }`),
}
