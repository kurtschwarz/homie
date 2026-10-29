import Foundation

/// Thrown by any step of a sync run once the server rejects the session, so the run stops and the
/// caller marks the session expired.
struct SessionExpired: Error {}

/// Rethrows `error` as `SessionExpired` when it is the server rejecting the session; otherwise returns.
func throwIfSessionExpired(_ error: Error) throws {
    if case HomieAPIError.unauthorized = error { throw SessionExpired() }
}
