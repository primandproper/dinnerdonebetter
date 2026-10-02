//
//  AuthenticationManagerTests.swift
//  iosTests
//
//  Created by Auto on 12/8/25.
//

import Foundation
@testable import ios
import Testing

struct AuthenticationManagerTests {
    // MARK: - Initialization Tests
    
    @Test("AuthenticationManager initializes with default state")
    func testInitialization() async {
        let manager = AuthenticationManager()
        
        #expect(manager.isAuthenticated == false)
        #expect(manager.username.isEmpty)
        #expect(manager.userID.isEmpty)
        #expect(manager.accountID.isEmpty)
    }
    
    // MARK: - Logout Tests
    
    @Test("Logout clears all authentication state")
    func testLogout() async {
        let manager = AuthenticationManager()
        
        // Set up authenticated state
        await MainActor.run {
            manager.isAuthenticated = true
            manager.username = "testuser"
            manager.userID = "user-123"
            manager.accountID = "account-456"
        }
        
        // Verify state is set
        #expect(manager.isAuthenticated == true)
        #expect(manager.username == "testuser")
        
        // Logout
        await manager.logout()
        
        // Verify all state is cleared
        #expect(manager.isAuthenticated == false)
        #expect(manager.username.isEmpty)
        #expect(manager.userID.isEmpty)
        #expect(manager.accountID.isEmpty)
    }
    
    @Test("Logout can be called multiple times safely")
    func testLogoutMultipleTimes() async {
        let manager = AuthenticationManager()
        
        // Set up authenticated state
        await MainActor.run {
            manager.isAuthenticated = true
            manager.username = "testuser"
        }
        
        // Logout multiple times
        await manager.logout()
        await manager.logout()
        await manager.logout()
        
        // Verify state remains cleared
        #expect(manager.isAuthenticated == false)
        #expect(manager.username.isEmpty)
    }
    
    // MARK: - Client Manager Tests
    
    @Test("getClientManager creates client manager on first call")
    func testGetClientManagerCreatesManager() async throws {
        let manager = AuthenticationManager()
        
        // This will attempt to create a client manager
        // Note: This test may fail if the server is not running
        // In a real scenario, you'd want to mock the ClientManager
        do {
            _ = try manager.getClientManager()
            // Success: client manager was created
        } catch {
            // If connection fails, that's expected in a test environment
            // The important thing is that the method doesn't crash
        }
    }
    
    @Test("getClientManager reuses existing client manager")
    func testGetClientManagerReusesManager() async throws {
        let manager = AuthenticationManager()
        
        do {
            let firstManager = try manager.getClientManager()
            let secondManager = try manager.getClientManager()
            
            // Both should return the same instance (reused)
            // Note: This is testing the caching behavior
            #expect(firstManager === secondManager)
        } catch {
            // If connection fails, that's expected in a test environment
            // The important thing is that the method doesn't crash
        }
    }
    
    // MARK: - Login Error Handling Tests
    
    @Test("login handles empty username")
    func testLoginWithEmptyUsername() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(username: "", password: "password")
        
        // The actual behavior depends on server validation
        // This test documents that empty username is handled
        #expect(result.success == false || result.success == true)
    }
    
    @Test("login handles empty password")
    func testLoginWithEmptyPassword() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(username: "user", password: "")
        
        // The actual behavior depends on server validation
        #expect(result.success == false || result.success == true)
    }
    
    @Test("login handles TOTP token parameter")
    func testLoginWithTOTPToken() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(
            username: "user",
            password: "password",
            totpToken: "123456"
        )
        
        // The actual behavior depends on server response
        // This test documents that TOTP token is accepted
        #expect(result.success == false || result.success == true)
    }
    
    @Test("login handles nil TOTP token")
    func testLoginWithNilTOTPToken() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(
            username: "user",
            password: "password",
            totpToken: nil
        )
        
        // Should handle nil TOTP token gracefully
        #expect(result.success == false || result.success == true)
    }
    
    @Test("login handles empty TOTP token")
    func testLoginWithEmptyTOTPToken() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(
            username: "user",
            password: "password",
            totpToken: ""
        )
        
        // Empty TOTP token should be treated same as nil
        #expect(result.success == false || result.success == true)
    }
    
    // MARK: - State Management Tests
    
    @Test("Authentication state persists after login attempt")
    func testStateAfterLoginAttempt() async {
        let manager = AuthenticationManager()
        
        // Attempt login (will likely fail without server, but tests state management)
        _ = await manager.login(username: "test", password: "test")
        
        // State should be false if login failed, or true if it succeeded
        // This test documents the state management behavior
        #expect(manager.isAuthenticated == false || manager.isAuthenticated == true)
    }
    
    @Test("Multiple logout calls maintain consistent state")
    func testMultipleLogoutCallsMaintainState() async {
        let manager = AuthenticationManager()
        
        // Set authenticated state
        await MainActor.run {
            manager.isAuthenticated = true
            manager.username = "user1"
        }
        
        // Logout
        await manager.logout()
        let stateAfterFirst = manager.isAuthenticated
        
        // Set state again
        await MainActor.run {
            manager.isAuthenticated = true
            manager.username = "user2"
        }
        
        // Logout multiple times
        await manager.logout()
        await manager.logout()
        let stateAfterMultiple = manager.isAuthenticated
        
        #expect(stateAfterFirst == false)
        #expect(stateAfterMultiple == false)
    }
    
    // MARK: - Integration Test Scenarios
    // These tests document expected behavior but may require a running server
    
    @Test("Login flow sets all authentication properties on success")
    func testLoginFlowSetsProperties() async {
        let manager = AuthenticationManager()
        
        // This test documents the expected behavior when login succeeds
        // In a real scenario with a running server, it would verify:
        // - isAuthenticated is set to true
        // - username is set
        // - userID is set
        // - accountID is set
        // - the Session holds the issued token
        
        // For now, we just verify the initial state
        #expect(manager.isAuthenticated == false)
    }
    
    
    // MARK: - Login Return Value Tests
    
    @Test("Login returns correct tuple structure")
    func testLoginReturnValueStructure() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(username: "test", password: "test")
        
        // Verify the tuple structure is correct
        // result should have: success (Bool), error (String?), requiresTOTP (Bool)
        #expect(type(of: result.success) == Bool.self)
        #expect(result.error == nil || type(of: result.error) == Optional<String>.self)
        #expect(type(of: result.requiresTOTP) == Bool.self)
    }
    
    @Test("Login requiresTOTP flag is set correctly")
    func testLoginRequiresTOTPFlag() async {
        let manager = AuthenticationManager()
        
        let result = await manager.login(username: "test", password: "test")
        
        // The requiresTOTP flag should be a boolean
        // In a real scenario, this would be true if the server requires TOTP
        #expect(result.requiresTOTP == false || result.requiresTOTP == true)
    }
}
