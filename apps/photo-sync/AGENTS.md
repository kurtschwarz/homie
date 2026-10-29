# Photo Sync

## Tests

- Write tests with Swift Testing (`import Testing`, `@Test`, `#expect`, `#require`). Do not import XCTest or subclass `XCTestCase`.
- Suites are structs. Setup goes in stored properties or `init`; Swift Testing creates a fresh instance per test, so state can't leak between tests. Mark a suite `@MainActor` when the code under test is.
- For isolated user defaults, use `TestDefaults.make()` as a stored property. Don't use a shared suite name and clear it by hand.
- Parameterize with `@Test(arguments:)` when the same assertion runs over several inputs, so a failing case is named in the results. Don't loop over inputs or stack sequential assertions.
- Report a failure from a branch with `Issue.record(...)`, not `XCTFail`.
- Don't disable, skip, or comment out a test to make it pass. Use `withKnownIssue` for a known external failure.
- Run the suite with `swift test` from `apps/photo-sync`.
