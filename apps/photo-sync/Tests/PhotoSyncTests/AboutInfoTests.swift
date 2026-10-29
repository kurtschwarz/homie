import Testing
@testable import PhotoSync

struct AboutInfoTests {
    @Test func versionAndBuildAreCombined() {
        let about = AboutInfo(info: ["CFBundleShortVersionString": "1.0", "CFBundleVersion": "12"])
        #expect(about.versionLine == "Version 1.0 (12)")
    }

    @Test func missingBuildDropsTheParentheses() {
        let about = AboutInfo(info: ["CFBundleShortVersionString": "1.0"])
        #expect(about.versionLine == "Version 1.0")
    }

    @Test(arguments: [[:], ["CFBundleVersion": "12"], ["CFBundleShortVersionString": ""]] as [[String: String]])
    func missingShortVersionHidesTheLine(info: [String: String]) {
        #expect(AboutInfo(info: info).versionLine == nil)
    }

    @Test func noInfoDictionaryHidesTheLine() {
        let about = AboutInfo(info: nil)
        #expect(about.versionLine == nil)
        #expect(about.commit == nil)
    }

    @Test func commitComesFromTheBundle() {
        #expect(AboutInfo(info: ["HomieGitCommit": "abc1234"]).commit == "abc1234")
    }

    @Test func emptyCommitIsTreatedAsMissing() {
        #expect(AboutInfo(info: ["HomieGitCommit": ""]).commit == nil)
    }
}
