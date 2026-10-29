import Testing
@testable import PhotoSync

struct HomieDiscoveryTests {
    @Test(arguments: [
        ("Homie · kitchen", "kitchen.local" as String?),
        ("Homie · ", nil),
        ("Something else", nil)
    ])
    func derivesHostnameFromServiceName(name: String, expected: String?) {
        #expect(HomieDiscovery.hostname(fromServiceName: name) == expected)
    }
}
