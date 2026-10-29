import Testing
@testable import PhotoSync

struct SaliencyTests {
    @Test(arguments: [
        (
            "converts Vision bottom-left coordinates to top-left",
            (x: 0.2, y: 0.3, width: 0.4, height: 0.25),
            (x: 0.2, y: 0.45, width: 0.4, height: 0.25)
        ),
        (
            "clamps the region to the image bounds",
            (x: 0.9, y: -0.1, width: 0.3, height: 0.4),
            (x: 0.9, y: 0.7, width: 0.1, height: 0.3)
        ),
    ])
    func topLeftRect(
        _ name: String,
        input: (x: Double, y: Double, width: Double, height: Double),
        expected: (x: Double, y: Double, width: Double, height: Double)
    ) {
        let rect = Saliency.topLeftRect(x: input.x, y: input.y, width: input.width, height: input.height)

        #expect(abs(rect.x - expected.x) < 0.0001)
        #expect(abs(rect.y - expected.y) < 0.0001)
        #expect(abs(rect.width - expected.width) < 0.0001)
        #expect(abs(rect.height - expected.height) < 0.0001)
    }
}
