import AppKit
import Foundation

enum PublisherError: LocalizedError {
    case usage
    case missingConfiguration
    case invalidRepository

    var errorDescription: String? {
        switch self {
        case .usage: return "usage: BackupPublisher {configure|publish <repository-directory>}"
        case .missingConfiguration: return "missing folder permission; run configure interactively"
        case .invalidRepository: return "repository does not contain required pgBackRest metadata"
        }
    }
}

let fileManager = FileManager.default
let configurationDirectory = fileManager.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
    .appendingPathComponent("FastTourneyBackupPublisher", isDirectory: true)
let stagingBookmark = configurationDirectory.appendingPathComponent("staging.bookmark")
let destinationBookmark = configurationDirectory.appendingPathComponent("icloud-backups.bookmark")

func selectedDirectory(prompt: String) throws -> URL {
    let panel = NSOpenPanel()
    panel.message = prompt
    panel.canChooseFiles = false
    panel.canChooseDirectories = true
    panel.allowsMultipleSelection = false
    guard panel.runModal() == .OK, let url = panel.url else { throw CocoaError(.userCancelled) }
    return url
}

func saveBookmark(for url: URL, at location: URL) throws {
    let bookmark = try url.bookmarkData(options: .withSecurityScope, includingResourceValuesForKeys: nil, relativeTo: nil)
    try fileManager.createDirectory(at: configurationDirectory, withIntermediateDirectories: true)
    try bookmark.write(to: location, options: .atomic)
}

func resolveBookmark(at location: URL) throws -> URL {
    guard fileManager.fileExists(atPath: location.path) else { throw PublisherError.missingConfiguration }
    var stale = false
    let bookmark = try Data(contentsOf: location)
    let url = try URL(resolvingBookmarkData: bookmark, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &stale)
    guard !stale else { throw PublisherError.missingConfiguration }
    guard url.startAccessingSecurityScopedResource() else { throw PublisherError.missingConfiguration }
    return url
}

func metadataExists(in repository: URL) -> Bool {
    fileManager.fileExists(atPath: repository.appendingPathComponent("backup/fasttourney-prod/backup.info").path)
        && fileManager.fileExists(atPath: repository.appendingPathComponent("archive/fasttourney-prod/archive.info").path)
}

func configure() throws {
    NSApplication.shared.setActivationPolicy(.accessory)
    NSApplication.shared.activate(ignoringOtherApps: true)
    let staging = try selectedDirectory(prompt: "Selecciona el directorio local de staging de FastTourney.")
    let destination = try selectedDirectory(prompt: "Selecciona FastTourney/postgresql-backups en iCloud Drive.")
    try saveBookmark(for: staging, at: stagingBookmark)
    try saveBookmark(for: destination, at: destinationBookmark)
    print("configured security-scoped access for staging and iCloud backups")
}

func publish(repositoryPath: String) throws {
    let staging = try resolveBookmark(at: stagingBookmark)
    defer { staging.stopAccessingSecurityScopedResource() }
    let destination = try resolveBookmark(at: destinationBookmark)
    defer { destination.stopAccessingSecurityScopedResource() }

    let repository = URL(fileURLWithPath: repositoryPath).standardizedFileURL
    guard repository.path.hasPrefix(staging.standardizedFileURL.path + "/"), metadataExists(in: repository) else {
        throw PublisherError.invalidRepository
    }

    let active = destination.appendingPathComponent("prod", isDirectory: true)
    let previous = destination.appendingPathComponent(".prod.previous", isDirectory: true)
    let temporary = destination.appendingPathComponent(".prod.staging.\(UUID().uuidString)", isDirectory: true)
    try fileManager.copyItem(at: repository, to: temporary)
    guard metadataExists(in: temporary) else { throw PublisherError.invalidRepository }
    if fileManager.fileExists(atPath: previous.path) { try fileManager.removeItem(at: previous) }
    if fileManager.fileExists(atPath: active.path) { try fileManager.moveItem(at: active, to: previous) }
    try fileManager.moveItem(at: temporary, to: active)
    if fileManager.fileExists(atPath: previous.path) { try fileManager.removeItem(at: previous) }
    print("published pgBackRest replica at \(active.path)")
}

do {
    let arguments = Array(CommandLine.arguments.dropFirst())
    if arguments.isEmpty || arguments == ["configure"] {
        try configure()
    } else if arguments.count == 2, arguments[0] == "publish" {
        try publish(repositoryPath: arguments[1])
    } else {
        throw PublisherError.usage
    }
} catch {
    FileHandle.standardError.write(Data("BackupPublisher: \(error.localizedDescription)\n".utf8))
    exit(1)
}
