import Testing
@testable import PhotoSync

struct TimeoutTests {
    @Test func returnsResultWhenOperationFinishesInTime() async throws {
        let result = try await withTimeout(.seconds(1)) { 42 }
        #expect(result == 42)
    }

    @Test func returnsNilWhenOperationExceedsTimeout() async throws {
        let result = try await withTimeout(.milliseconds(50)) {
            try await Task.sleep(for: .seconds(5))
            return 42
        }
        #expect(result == nil)
    }

    @Test func earlierLookupTimeoutDoesNotCancelLaterLookup() async throws {
        let first = try await withTimeout(.milliseconds(300)) { 1 }
        #expect(first == 1)

        let second = try await withTimeout(.milliseconds(300)) {
            try await Task.sleep(for: .milliseconds(200))
            return 2
        }
        #expect(second == 2)
    }

    @Test func operationErrorsPropagate() async {
        struct Failure: Error {}
        await #expect(throws: Failure.self) {
            try await withTimeout(.seconds(1)) { () -> Int in throw Failure() }
        }
    }

    @Test func cancellationPropagates() async {
        let task = Task {
            try await withTimeout(.seconds(5)) {
                try await Task.sleep(for: .seconds(5))
                return 1
            }
        }
        task.cancel()
        await #expect(throws: CancellationError.self) {
            _ = try await task.value
        }
    }
}
