package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	usersTableName       = "users"
	userAvatarsTableName = "user_avatars"
	uploadedMediaTable   = "uploaded_media"

	createdByUserColumn = "created_by_user"

	usernameColumn               = "username"
	emailAddressColumn           = "email_address"
	hashedPasswordColumn         = "hashed_password"
	passwordLastChangedAtColumn  = "password_last_changed_at"
	requiresPasswordChangeColumn = "requires_password_change"
	/* #nosec G101 */
	twoFactorSecretColumn = "two_factor_secret"
	/* #nosec G101 */
	twoFactorSecretVerifiedAtColumn     = "two_factor_secret_verified_at"
	userAccountStatusColumn             = "user_account_status"
	userAccountStatusExplanationColumn  = "user_account_status_explanation"
	birthdayColumn                      = "birthday"
	emailAddressVerificationTokenColumn = "email_address_verification_token"
	emailAddressVerifiedAtColumn        = "email_address_verified_at"
	firstNameColumn                     = "first_name"
	lastNameColumn                      = "last_name"
	lastAcceptedTOSColumn               = "last_accepted_terms_of_service"
	lastAcceptedPrivacyPolicyColumn     = "last_accepted_privacy_policy"
)

func init() {
	registerTableName(usersTableName)
	registerTableName(userAvatarsTableName)
	registerTableName(uploadedMediaTable)
}

// avatarJoinColumns are the uploaded_media columns selected when joining for avatar.
var avatarJoinColumns = []string{
	"id", "storage_path", "mime_type", "created_at", "last_updated_at", "archived_at", "created_by_user",
}

func avatarJoinSelect(prefix string) []string {
	return applyToEach(avatarJoinColumns, func(_ int, s string) string {
		return fmt.Sprintf("%s.%s as %s_%s", uploadedMediaTable, s, prefix, s)
	})
}

const avatarJoinClause = `LEFT JOIN ` + userAvatarsTableName + ` ON ` + userAvatarsTableName + `.belongs_to_user = ` + usersTableName + `.id AND ` + userAvatarsTableName + `.archived_at IS NULL
	LEFT JOIN ` + uploadedMediaTable + ` ON ` + uploadedMediaTable + `.id = ` + userAvatarsTableName + `.uploaded_media_id AND ` + uploadedMediaTable + `.archived_at IS NULL`

var usersColumns = []string{
	idColumn,
	usernameColumn,
	emailAddressColumn,
	hashedPasswordColumn,
	passwordLastChangedAtColumn,
	requiresPasswordChangeColumn,
	twoFactorSecretColumn,
	twoFactorSecretVerifiedAtColumn,
	userAccountStatusColumn,
	userAccountStatusExplanationColumn,
	birthdayColumn,
	emailAddressVerificationTokenColumn,
	emailAddressVerifiedAtColumn,
	firstNameColumn,
	lastNameColumn,
	lastAcceptedTOSColumn,
	lastAcceptedPrivacyPolicyColumn,
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildUpdateAccountMembershipsQuery(ownershipColumn string, nowColumns []string) string {
	var updateQueryBuilder builq.Builder

	addendum := ""
	for _, col := range nowColumns {
		addendum = fmt.Sprintf(",\n\t%s = %s", col, currentTimeExpression)
	}

	builder := updateQueryBuilder.Addf(
		`UPDATE %s SET
	%s = %s%s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
		accountUserMembershipsTableName,
		archivedAtColumn,
		currentTimeExpression,
		addendum,
		archivedAtColumn,
		ownershipColumn,
		idColumn,
	)

	return buildRawQuery(builder)
}

func buildUserUpdateQuery(columnName string, nowColumns []string) string {
	var updateQueryBuilder builq.Builder

	addendum := ""
	for _, col := range nowColumns {
		addendum = fmt.Sprintf(",\n\t%s = %s", col, currentTimeExpression)
	}

	builder := updateQueryBuilder.Addf(
		`UPDATE %s SET
	%s = %s%s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
		usersTableName,
		columnName,
		currentTimeExpression,
		addendum,
		archivedAtColumn,
		idColumn,
		idColumn,
	)

	return buildRawQuery(builder)
}

func buildUsersQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(usersTableName, usersColumns,
				querygen.WithEntity("User", "Users"),
				querygen.WithDatabaseOwned(
					passwordLastChangedAtColumn,
					emailAddressVerifiedAtColumn,
					lastAcceptedTOSColumn,
					lastAcceptedPrivacyPolicyColumn,
				),
				querygen.WithOmitted(
					querygen.GetQuery,
					querygen.ExistsQuery,
					querygen.ListQuery,
					querygen.UpdateQuery,
				),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "AcceptPrivacyPolicyForUser",
						Type: ExecType,
					},
					Content: buildUserUpdateQuery(lastAcceptedPrivacyPolicyColumn, nil),
				},
				{
					Annotation: QueryAnnotation{
						Name: "AcceptTermsOfServiceForUser",
						Type: ExecType,
					},
					Content: buildUserUpdateQuery(lastAcceptedTOSColumn, nil),
				},
				{
					Annotation: QueryAnnotation{
						Name: "DeleteUser",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = sqlc.arg(%s);`,
						usersTableName,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAdminUserByUsername",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
	JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL AND %s.%s IS NULL
	JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
WHERE %s.%s IS NULL
	AND %s.%s = 'service_admin'
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NOT NULL;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						userRoleAssignmentsTableName, userRoleAssignmentsTableName, userIDColumn, usersTableName, idColumn, userRoleAssignmentsTableName, accountIDColumn, userRoleAssignmentsTableName, archivedAtColumn,
						userRolesTableName, userRolesTableName, idColumn, userRoleAssignmentsTableName, roleIDColumn, userRolesTableName, archivedAtColumn,
						usersTableName, archivedAtColumn,
						userRolesTableName, nameColumn,
						usersTableName, usernameColumn, usernameColumn,
						usersTableName, twoFactorSecretVerifiedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserByEmail",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, emailAddressColumn, emailAddressColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserByEmailAddressVerificationToken",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, emailAddressVerificationTokenColumn, emailAddressVerificationTokenColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserByID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserByUsername",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, usernameColumn, usernameColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetEmailVerificationTokenByUserID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						usersTableName, emailAddressVerificationTokenColumn,
						usersTableName,
						usersTableName, archivedAtColumn,
						usersTableName, emailAddressVerifiedAtColumn,
						usersTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUsers",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s,
	%s
FROM %s
	%s
WHERE %s
%s;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						querygen.FilterCountSelect(usersTableName, usersColumns, []string{}),
						querygen.TotalCountSelect(usersTableName, usersColumns, []string{}),
						usersTableName,
						avatarJoinClause,
						querygen.FilterConditions(usersTableName, usersColumns),
						querygen.CursorLimitClause(usersTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUsersForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s,
	%s
FROM %s
	%s
JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						querygen.FilterCountSelect(usersTableName, usersColumns, nil),
						querygen.TotalCountSelect(usersTableName, usersColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountUserMembershipsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						usersTableName,
						avatarJoinClause,
						accountUserMembershipsTableName,
						accountUserMembershipsTableName,
						belongsToUserColumn,
						usersTableName,
						idColumn,
						querygen.FilterConditions(usersTableName, usersColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountUserMembershipsTableName, belongsToAccountColumn, belongsToAccountColumn),
							fmt.Sprintf("%s.%s IS NULL", accountUserMembershipsTableName, archivedAtColumn),
						),
						querygen.CursorLimitClause(usersTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUsersWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(applyToEach(usersColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName,
						archivedAtColumn,
						usersTableName,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserRequiresPasswordChange",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						usersTableName, requiresPasswordChangeColumn,
						usersTableName,
						usersTableName, archivedAtColumn,
						usersTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserIDsNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL;`,
						usersTableName, idColumn,
						usersTableName,
						usersTableName, archivedAtColumn,
						usersTableName, lastIndexedAtColumn,
						usersTableName, lastIndexedAtColumn, currentTimeExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserWithUnverifiedTwoFactor",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, idColumn, idColumn,
						usersTableName, twoFactorSecretVerifiedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserWithVerifiedTwoFactor",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s
FROM %s
	%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NOT NULL;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						usersTableName,
						avatarJoinClause,
						usersTableName, archivedAtColumn,
						usersTableName, idColumn, idColumn,
						usersTableName, twoFactorSecretVerifiedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkEmailAddressAsVerified",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						emailAddressVerifiedAtColumn, currentTimeExpression,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						emailAddressVerifiedAtColumn,
						idColumn, idColumn,
						emailAddressVerificationTokenColumn, emailAddressVerificationTokenColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkEmailAddressAsUnverified",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = NULL,
	%s = %s
WHERE %s IS NULL
	AND %s IS NOT NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						emailAddressVerifiedAtColumn,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						emailAddressVerifiedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkTwoFactorSecretAsUnverified",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = NULL,
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						twoFactorSecretVerifiedAtColumn,
						twoFactorSecretColumn, twoFactorSecretColumn,
						lastUpdatedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkTwoFactorSecretAsVerified",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						twoFactorSecretVerifiedAtColumn, currentTimeExpression,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchUsersByUsername",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s,
	%s
FROM %s
	%s
WHERE %s
%s;`,
						strings.Join(applyToEach(usersColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", usersTableName, s)
						}), ",\n\t"),
						strings.Join(avatarJoinSelect("avatar"), ",\n\t"),
						querygen.FilterCountSelect(usersTableName, usersColumns, []string{}),
						querygen.TotalCountSelect(usersTableName, usersColumns, []string{}),
						usersTableName,
						avatarJoinClause,
						querygen.FilterConditions(usersTableName, usersColumns,
							fmt.Sprintf("%s.%s %s", usersTableName, usernameColumn, buildILIKEForArgument(usernameColumn)),
						),
						querygen.CursorLimitClause(usersTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserDetails",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						firstNameColumn, firstNameColumn,
						lastNameColumn, lastNameColumn,
						birthdayColumn, birthdayColumn,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserEmailAddress",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = NULL,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						emailAddressColumn, emailAddressColumn,
						emailAddressVerifiedAtColumn,
						lastUpdatedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserPassword",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = FALSE,
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						hashedPasswordColumn, hashedPasswordColumn,
						requiresPasswordChangeColumn,
						passwordLastChangedAtColumn, currentTimeExpression,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserTwoFactorSecret",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = NULL,
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						twoFactorSecretVerifiedAtColumn,
						twoFactorSecretColumn, twoFactorSecretColumn,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserUsername",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						usersTableName,
						usernameColumn, usernameColumn,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
