struct Person: Identifiable, Equatable {
    let id: String?
    var name: String
    var emoji: String
    var swatch: String
    var isFamily: Bool

    static let defaultSwatch = "slate"
    static let swatches = ["clay", "amber", "moss", "pine", "teal", "sky", "iris", "orchid", "plum", "slate"]
}
