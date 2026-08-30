import Foundation

/// The wire type for a list query's request half.
///
/// Its schema is not this repository's to keep: platform-go ships the `.proto`
/// inside the published module and the generated Swift carries that module's
/// package name, which is why the generated type is spelled
/// `Primandproper_Platform_Filtering_V1_QueryFilter`. This is what the app
/// calls it.
public typealias QueryFilterMessage = Primandproper_Platform_Filtering_V1_QueryFilter

/// The wire type for a list query's response half: what was applied, and where
/// the next page starts.
public typealias PaginationMessage = Primandproper_Platform_Filtering_V1_Pagination
