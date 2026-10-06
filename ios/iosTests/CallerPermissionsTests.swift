//
//  CallerPermissionsTests.swift
//  iosTests
//
//  The caller's own controls are gated on the effective permissions GetPrincipal answers,
//  never on role names: a client that copies the server's role-to-permission table goes stale
//  the moment the server's changes.
//

import Foundation
import PlatformClient
import Testing

@testable import ios

private func principalResponse(
  granting permissions: [String]?
) -> Primandproper_Platform_Identity_V1_GetPrincipalResponse {
  var response = Primandproper_Platform_Identity_V1_GetPrincipalResponse()
  if let permissions {
    var effective = Primandproper_Platform_Identity_V1_EffectivePermissions()
    effective.permissions = permissions
    response.permissions = effective
  }
  return response
}

struct CallerPermissionsTests {
  @Test("a granted permission is allowed and an ungranted one is not")
  func grantedAndNot() {
    let permissions = CallerPermissions(
      principalResponse(granting: [CallerPermissions.Name.inviteMembers]))

    #expect(permissions.allows(CallerPermissions.Name.inviteMembers))
    #expect(!permissions.allows(CallerPermissions.Name.manageMembers))
  }

  @Test("an empty list is a caller who may do nothing here")
  func emptyListAllowsNothing() {
    let permissions = CallerPermissions(principalResponse(granting: []))

    #expect(!permissions.allows(CallerPermissions.Name.updateAccount))
  }

  @Test("an absent field is a server that did not say, which is not 'holds nothing'")
  func absentFieldIsNotANo() {
    let permissions = CallerPermissions(principalResponse(granting: nil))

    #expect(permissions.allows(CallerPermissions.Name.updateAccount))
    #expect(permissions.allows(CallerPermissions.Name.updateRecipeStatus))
  }

  @Test("nothing is allowed before the server has answered")
  func unloadedAllowsNothing() {
    #expect(!CallerPermissions.unloaded.allows(CallerPermissions.Name.updateAccount))
    #expect(!CallerPermissions.unloaded.allows(CallerPermissions.Name.updateRecipeStatus))
  }

  @Test("a role name is not a permission")
  func roleNameIsNotAPermission() {
    let permissions = CallerPermissions(
      principalResponse(granting: ["account_admin", "service_admin"]))

    #expect(!permissions.allows(CallerPermissions.Name.updateAccount))
    #expect(!permissions.allows(CallerPermissions.Name.updateRecipeStatus))
  }
}

struct RecipeListPermissionGateTests {
  @Test("the submitted-recipes toggle shows for a caller granted recipe review")
  @MainActor
  func reviewerSeesToggle() {
    let viewModel = RecipeListViewModel(authManager: AuthenticationManager())

    viewModel.permissions = CallerPermissions(
      granted: [CallerPermissions.Name.updateRecipeStatus])

    #expect(viewModel.canReviewRecipes)
  }

  @Test("the submitted-recipes toggle is hidden for everyone else, and until loaded")
  @MainActor
  func othersDoNot() {
    let viewModel = RecipeListViewModel(authManager: AuthenticationManager())
    #expect(!viewModel.canReviewRecipes)

    viewModel.permissions = CallerPermissions(
      granted: [CallerPermissions.Name.inviteMembers, CallerPermissions.Name.updateAccount])
    #expect(!viewModel.canReviewRecipes)
  }
}
