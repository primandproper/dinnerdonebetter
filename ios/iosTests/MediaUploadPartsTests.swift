//
//  MediaUploadPartsTests.swift
//  iosTests
//

import Foundation
@testable import ios
import PlatformClient
import Testing

@Suite("MediaUploadParts")
struct MediaUploadPartsTests {
  @Test("the header comes first, then the bytes in order")
  func testHeaderThenChunks() throws {
    let name = UUID().uuidString + ".png"
    let data = Data((0..<(150 * 1024)).map { UInt8($0 % 251) })

    let parts = MediaUploadParts.make(name: name, contentType: "image/png", data: data)

    guard case .header(let header) = parts.first?.part else {
      Issue.record("the first part is not a header")
      return
    }
    #expect(header.name == name)
    #expect(header.contentType == "image/png")
    #expect(!header.hasBelongsTo)

    var reassembled = Data()
    for part in parts.dropFirst() {
      guard case .chunk(let chunk) = part.part else {
        Issue.record("a part after the header is not a chunk")
        return
      }
      #expect(chunk.count <= 64 * 1024)
      reassembled.append(chunk)
    }
    #expect(reassembled == data)
  }

  @Test("an avatar is attached to its uploader")
  func testBelongsTo() throws {
    let userID = UUID().uuidString

    var subject = Primandproper_Platform_Mediaregistry_V1_Subject()
    subject.type = mediaRegistryUserSubjectType
    subject.id = userID

    let parts = MediaUploadParts.make(
      name: UUID().uuidString + ".jpg", contentType: "image/jpeg", data: Data([1, 2, 3]),
      belongsTo: subject)

    guard case .header(let header) = parts.first?.part else {
      Issue.record("the first part is not a header")
      return
    }
    #expect(header.belongsTo.type == "user")
    #expect(header.belongsTo.id == userID)
    #expect(parts.count == 2)
  }

  @Test("an empty upload is a header alone")
  func testEmpty() throws {
    let parts = MediaUploadParts.make(
      name: UUID().uuidString + ".png", contentType: "image/png", data: Data())

    #expect(parts.count == 1)
  }
}
