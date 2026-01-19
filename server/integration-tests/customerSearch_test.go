package test

import (
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) TestSearchUsers_EmptySearchAllStatus() {
	// Test with empty search term and ALL status
	users, totalCount, err := dao.MasterUserRecordDao{}.SearchUsers("", "ALL", 25, 0)

	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().NotNil(users, "Users slice should not be nil")
	suite.Require().Equal(totalCount, int64(0), "Total count should be non-negative")
}

func (suite *IntegrationTestSuite) TestSearchUsers_WithTestData() {
	user1 := dao.MasterUserRecordDao{
		Id:                   uuid.New().String(),
		FirstName:            "John",
		LastName:             "Doe",
		Email:                "john.doe@example.com",
		MobileNo:             "+15551234567",
		UserStatus:           constant.ACTIVE,
		LedgerCustomerNumber: "LEDGER001",
	}

	user2 := dao.MasterUserRecordDao{
		Id:                   uuid.New().String(),
		FirstName:            "Jane",
		LastName:             "Smith",
		Email:                "jane.smith@example.com",
		MobileNo:             "+15559876543",
		UserStatus:           constant.USER_CREATED,
		LedgerCustomerNumber: "LEDGER002",
	}

	user3 := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Bob",
		LastName:   "Gohnson",
		Email:      "bob.gohnson@example.com",
		MobileNo:   "+15555555555",
		UserStatus: constant.ACTIVE,
		// No ledger customer number
	}

	err := suite.TestDB.Create(&user1).Error
	suite.Require().NoError(err, "Failed to create test user 1")

	err = suite.TestDB.Create(&user2).Error
	suite.Require().NoError(err, "Failed to create test user 2")

	err = suite.TestDB.Create(&user3).Error
	suite.Require().NoError(err, "Failed to create test user 3")

	// Test Search all users with ALL status
	users, totalCount, err := dao.MasterUserRecordDao{}.SearchUsers("", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(3), "Should find at least our 3 test users")
	suite.Require().Equal(len(users), 3, "Should return at least our 3 test users")

	// Test Filter by ACTIVE status
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("", "ACTIVE", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(2), "Should find at least 2 ACTIVE users")

	// Verify all returned users have ACTIVE status
	for _, user := range users {
		suite.Equal(constant.ACTIVE, user.UserStatus, "All returned users should have ACTIVE status")
	}

	// Test Filter by USER_CREATED status
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("", "USER_CREATED", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find at least 1 USER_CREATED user")

	// Verify all returned users have USER_CREATED status
	for _, user := range users {
		suite.Equal(constant.USER_CREATED, user.UserStatus, "All returned users should have USER_CREATED status")
	}

	// Test Search by first name (single term)
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("john", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find John Doe")
	suite.Require().Equal(users[0].Id, user1.Id, "Should find John Doe by first name")

	// Test Search by email
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("jane.smith@example.com", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find exactly 1 user")
	suite.Require().Equal(users[0].Id, user2.Id, "Should find Jane Smith by email")

	// Test Search by phone number
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("5559876543", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find exactly 1 user")
	suite.Require().Equal(users[0].Id, user2.Id, "Should find Jane Smith by email")

	// Test Search by ledger customer number
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("LEDGER001", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find exactly 1 user")
	suite.Require().Equal(users[0].Id, user1.Id, "Should find John Doe by ledger customer number")

	// Test AND search with multiple terms
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("jane smith", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find exactly 1 user")
	suite.Require().Equal(users[0].Id, user2.Id, "Should find Jane Smith by split name")

	// Test AND search that should exclude results
	users, _, err = dao.MasterUserRecordDao{}.SearchUsers("jane doe", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")

	// Should not find anyone because no user has both "jane" and "doe"
	foundJaneSmith := false
	foundJohnDoe := false
	for _, user := range users {
		if user.FirstName == "Jane" && user.LastName == "Smith" {
			foundJaneSmith = true
		}
		if user.FirstName == "John" && user.LastName == "Doe" {
			foundJohnDoe = true
		}
	}
	suite.False(foundJaneSmith, "Jane Smith should not be found when searching for 'jane doe'")
	suite.False(foundJohnDoe, "John Doe should not be found when searching for 'jane doe'")

	// Test Case insensitive search
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("JOHN", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Case insensitive search should find John")
	suite.Require().Equal(users[0].Id, user1.Id, "Should find John Doe by first name")

	// Test Complex email search should be treated as single term
	complexEmailUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Craig",
		LastName:   "Buchanan",
		Email:      "craigb+487013ac@mojotech.com",
		MobileNo:   "+15551111111",
		UserStatus: constant.ACTIVE,
	}
	err = suite.TestDB.Create(&complexEmailUser).Error
	suite.Require().NoError(err, "Failed to create complex email user")

	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("craigb+487013ac@mojotech.com", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find user by complex email")
	suite.Require().Equal(users[0].Id, complexEmailUser.Id, "Should find Craig Buchanan by email")

	// Test Formatted phone number search
	formattedPhoneUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Phone",
		LastName:   "User",
		Email:      "phone@example.com",
		MobileNo:   "+14012753374",
		UserStatus: constant.ACTIVE,
	}
	err = suite.TestDB.Create(&formattedPhoneUser).Error
	suite.Require().NoError(err, "Failed to create formatted phone user")

	// Search with formatted phone number "(401) 275-3374"
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("(401) 275-3374", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find user by formatted phone")
	suite.Require().Equal(users[0].Id, formattedPhoneUser.Id, "Should find Phone User by formatted phone number")

	// Test No results found
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("nonexistentuser", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Equal(int64(0), totalCount, "Should find no users for nonexistent search")
	suite.Equal(0, len(users), "Should return empty slice for nonexistent search")
}

func (suite *IntegrationTestSuite) TestSearchUsers_Pagination() {
	// Create multiple test users for pagination testing
	var testUsers []dao.MasterUserRecordDao
	for i := range 30 {
		user := dao.MasterUserRecordDao{
			Id:         uuid.New().String(),
			FirstName:  "TestUser",
			LastName:   "Number" + string(rune(i+48)), // Convert to character
			Email:      "testuser" + string(rune(i+48)) + "@example.com",
			MobileNo:   "+1555000000" + string(rune(i+48)),
			UserStatus: constant.ACTIVE,
		}
		testUsers = append(testUsers, user)
	}

	// Insert all test users
	for _, user := range testUsers {
		err := suite.TestDB.Create(&user).Error
		suite.Require().NoError(err, "Failed to create test user")
	}

	// Test pagination - first page
	users, totalCount, err := dao.MasterUserRecordDao{}.SearchUsers("testuser", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(30), "Should find at least 30 test users")
	suite.Require().LessOrEqual(len(users), 25, "Should return at most 25 users per page")

	// Test pagination - second page
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("testuser", "ALL", 25, 25)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(30), "Total count should be consistent")
	suite.Require().Equal(len(users), 5, "Second page should have remaining users")

	// Test with limit smaller than total results
	users, _, err = dao.MasterUserRecordDao{}.SearchUsers("testuser", "ALL", 10, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Equal(10, len(users), "Should return exactly 10 users when limit is 10")
}

func (suite *IntegrationTestSuite) TestSearchUsers_StatusFilterWithSearch() {
	// Create users with different statuses
	activeUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Active",
		LastName:   "User",
		Email:      "active@example.com",
		MobileNo:   "+15551111111",
		UserStatus: constant.ACTIVE,
	}

	createdUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Created",
		LastName:   "User",
		Email:      "created@example.com",
		MobileNo:   "+15552222222",
		UserStatus: constant.USER_CREATED,
	}

	err := suite.TestDB.Create(&activeUser).Error
	suite.Require().NoError(err, "Failed to create active user")

	err = suite.TestDB.Create(&createdUser).Error
	suite.Require().NoError(err, "Failed to create created user")

	// Search for "user" with ACTIVE status filter
	users, totalCount, err := dao.MasterUserRecordDao{}.SearchUsers("user", "ACTIVE", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find at least 1 ACTIVE user")

	// Verify only ACTIVE users are returned
	foundActive := false
	foundCreated := false
	for _, user := range users {
		if user.Email == "active@example.com" {
			foundActive = true
			suite.Equal(constant.ACTIVE, user.UserStatus, "Active user should have ACTIVE status")
		}
		if user.Email == "created@example.com" {
			foundCreated = true
		}
	}
	suite.True(foundActive, "Should find active user")
	suite.False(foundCreated, "Should not find created user when filtering by ACTIVE")

	// Search for "user" with USER_CREATED status filter
	users, totalCount, err = dao.MasterUserRecordDao{}.SearchUsers("user", "USER_CREATED", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")
	suite.Require().Equal(totalCount, int64(1), "Should find at least 1 USER_CREATED user")

	// Verify only USER_CREATED users are returned
	foundActive = false
	foundCreated = false
	for _, user := range users {
		if user.Email == "active@example.com" {
			foundActive = true
		}
		if user.Email == "created@example.com" {
			foundCreated = true
			suite.Equal(constant.USER_CREATED, user.UserStatus, "Created user should have USER_CREATED status")
		}
	}
	suite.False(foundActive, "Should not find active user when filtering by USER_CREATED")
	suite.True(foundCreated, "Should find created user")
}

func (suite *IntegrationTestSuite) TestSearchTermParsing() {
	// Test the parseSearchTerms function directly by calling SearchUsers
	// and verifying that field hints work correctly

	// Create test users with overlapping data to test field specificity
	emailUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Test",
		LastName:   "Email",
		Email:      "test.email@example.com",
		MobileNo:   "+15551234567",
		UserStatus: constant.ACTIVE,
	}

	phoneUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Test",
		LastName:   "Phone",
		Email:      "different@example.com",
		MobileNo:   "+14012753374", // This matches (401) 275-3374
		UserStatus: constant.ACTIVE,
	}

	nameUser := dao.MasterUserRecordDao{
		Id:         uuid.New().String(),
		FirstName:  "Example",
		LastName:   "User",
		Email:      "name@test.com",
		MobileNo:   "+15559876543",
		UserStatus: constant.ACTIVE,
	}

	// Insert test users
	err := suite.TestDB.Create(&emailUser).Error
	suite.Require().NoError(err, "Failed to create email test user")

	err = suite.TestDB.Create(&phoneUser).Error
	suite.Require().NoError(err, "Failed to create phone test user")

	err = suite.TestDB.Create(&nameUser).Error
	suite.Require().NoError(err, "Failed to create name test user")

	// Test Email search should only match email field
	users, _, err := dao.MasterUserRecordDao{}.SearchUsers("test.email@example.com", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")

	foundEmailUser := false
	for _, user := range users {
		if user.Email == "test.email@example.com" {
			foundEmailUser = true
			break
		}
	}
	suite.True(foundEmailUser, "Should find email user when searching by email")

	// Test Phone search should normalize and match
	users, _, err = dao.MasterUserRecordDao{}.SearchUsers("(401) 275-3374", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")

	foundCorrectPhone := false
	for _, user := range users {
		if user.MobileNo == "+14012753374" {
			foundCorrectPhone = true
			break
		}
	}
	suite.True(foundCorrectPhone, "Should find user with +14012753374 when searching for (401) 275-3374")

	// Test Phone search should operate on partial
	users, _, err = dao.MasterUserRecordDao{}.SearchUsers("275", "ALL", 25, 0)
	suite.Require().NoError(err, "SearchUsers should not return an error")

	foundCorrectPhone = false
	for _, user := range users {
		if user.MobileNo == "+14012753374" {
			foundCorrectPhone = true
			break
		}
	}
	suite.True(foundCorrectPhone, "Should find user with +14012753374 when searching for 275")
}
