import Foundation

/// Runs `operation`, returning nil if it hasn't finished within `timeout`.
/// The losing task is cancelled, so no timer outlives the call.
func withTimeout<T: Sendable>(
    _ timeout: Duration,
    operation: @escaping @Sendable () async throws -> T
) async throws -> T? {
    try await withThrowingTaskGroup(of: T?.self) { group in
        group.addTask { try await operation() }
        group.addTask {
            try await Task.sleep(for: timeout)
            return nil
        }
        defer { group.cancelAll() }
        return try await group.next() ?? nil
    }
}
