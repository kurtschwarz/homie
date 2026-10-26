import CoreGraphics
import Foundation
import Vision

struct FocalRegionInput: Encodable {
    let x: Double
    let y: Double
    let width: Double
    let height: Double
    let source: String
}

enum Saliency {
    static func focalRegion(for image: CGImage) -> FocalRegionInput? {
        let image = downscaled(image, maximumDimension: 1024)
        let faceRequest = VNDetectFaceRectanglesRequest()
        let handler = VNImageRequestHandler(cgImage: image, orientation: .up)
        if (try? handler.perform([faceRequest])) != nil,
           let faces = faceRequest.results, !faces.isEmpty {
            let rect = faces.map(\.boundingBox).reduce(faces[0].boundingBox) { $0.union($1) }
            return normalizedRegion(rect, source: "vision-face")
        }

        let attentionRequest = VNGenerateAttentionBasedSaliencyImageRequest()
        guard (try? handler.perform([attentionRequest])) != nil,
              let rect = attentionRequest.results?.first?.salientObjects?.first?.boundingBox else { return nil }
        return normalizedRegion(rect, source: "vision-attention")
    }

    static func topLeftRect(x: Double, y: Double, width: Double, height: Double) -> (x: Double, y: Double, width: Double, height: Double) {
        let left = min(max(x, 0), 1)
        let top = min(max(1 - (y + height), 0), 1)
        let right = min(max(x + width, left), 1)
        let bottom = min(max(top + height, top), 1)
        return (left, top, right - left, bottom - top)
    }

    private static func normalizedRegion(_ rect: CGRect, source: String) -> FocalRegionInput {
        let topLeft = topLeftRect(x: rect.origin.x, y: rect.origin.y, width: rect.width, height: rect.height)
        return FocalRegionInput(x: topLeft.x, y: topLeft.y, width: topLeft.width, height: topLeft.height, source: source)
    }

    private static func downscaled(_ image: CGImage, maximumDimension: Int) -> CGImage {
        let longest = max(image.width, image.height)
        guard longest > maximumDimension else { return image }
        let scale = CGFloat(maximumDimension) / CGFloat(longest)
        let width = max(1, Int((CGFloat(image.width) * scale).rounded()))
        let height = max(1, Int((CGFloat(image.height) * scale).rounded()))
        guard let context = CGContext(data: nil, width: width, height: height, bitsPerComponent: 8,
                                      bytesPerRow: 0, space: CGColorSpaceCreateDeviceRGB(),
                                      bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue) else { return image }
        context.interpolationQuality = .high
        context.draw(image, in: CGRect(x: 0, y: 0, width: width, height: height))
        return context.makeImage() ?? image
    }
}
