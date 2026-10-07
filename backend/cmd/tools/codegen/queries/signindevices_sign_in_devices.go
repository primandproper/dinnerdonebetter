package main

import (
	"fmt"
)

const signInDevicesTableName = "sign_in_devices"

// buildSignInDevicesQueries are the reads and writes behind "where you're signed in": one row
// per login, recorded when a token is issued for it and read back for the logins a listing is
// about to return.
func buildSignInDevicesQueries(database string) []*Query {
	switch database {
	case postgres:
		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "UpsertSignInDevice",
					Type: ExecType,
				},
				// A refresh renews the row rather than adding one: it is the same login, and
				// what the screen wants is where it is held now. The owner is never moved —
				// a family is one person's, and a row naming somebody else is not this one.
				Content: fmt.Sprintf(`INSERT INTO %s (family_id, belongs_to_user, ip_address, user_agent, device_name, expires_at)
VALUES (sqlc.arg(family_id), sqlc.arg(belongs_to_user), sqlc.arg(ip_address), sqlc.arg(user_agent), sqlc.arg(device_name), sqlc.arg(expires_at))
ON CONFLICT (family_id) DO UPDATE SET
	ip_address = EXCLUDED.ip_address,
	user_agent = EXCLUDED.user_agent,
	device_name = EXCLUDED.device_name,
	expires_at = EXCLUDED.expires_at,
	last_seen_at = NOW()
WHERE %s.belongs_to_user = EXCLUDED.belongs_to_user;`, signInDevicesTableName, signInDevicesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetSignInDevicesForFamilies",
					Type: ManyType,
				},
				Content: fmt.Sprintf(`SELECT family_id, ip_address, user_agent, device_name, created_at, last_seen_at
FROM %s
WHERE belongs_to_user = sqlc.arg(belongs_to_user)
	AND family_id = ANY(sqlc.arg(family_ids)::TEXT[]);`, signInDevicesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetSignInDevicesForUser",
					Type: ManyType,
				},
				Content: fmt.Sprintf(`SELECT family_id, ip_address, user_agent, device_name, created_at, last_seen_at, expires_at
FROM %s
WHERE belongs_to_user = sqlc.arg(belongs_to_user)
ORDER BY created_at;`, signInDevicesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "DeleteExpiredSignInDevices",
					Type: ExecRowsType,
				},
				Content: fmt.Sprintf(`DELETE FROM %s WHERE expires_at < NOW();`, signInDevicesTableName),
			},
		}
	default:
		return nil
	}
}
