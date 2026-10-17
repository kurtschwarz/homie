import XCTest
@testable import PhotoSync

final class SaliencyTests: XCTestCase {
    func testConvertsVisionBottomLeftCoordinatesToTopLeft() {
        let rect = Saliency.topLeftRect(x: 0.2, y: 0.3, width: 0.4, height: 0.25)

        XCTAssertEqual(rect.x, 0.2, accuracy: 0.0001)
        XCTAssertEqual(rect.y, 0.45, accuracy: 0.0001)
        XCTAssertEqual(rect.width, 0.4, accuracy: 0.0001)
        XCTAssertEqual(rect.height, 0.25, accuracy: 0.0001)
    }

    func testClampsRegionToImageBounds() {
        let rect = Saliency.topLeftRect(x: 0.9, y: -0.1, width: 0.3, height: 0.4)

        XCTAssertEqual(rect.x, 0.9, accuracy: 0.0001)
        XCTAssertEqual(rect.y, 0.7, accuracy: 0.0001)
        XCTAssertEqual(rect.width, 0.1, accuracy: 0.0001)
        XCTAssertEqual(rect.height, 0.3, accuracy: 0.0001)
    }
}
