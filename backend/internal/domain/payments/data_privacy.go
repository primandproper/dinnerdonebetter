package payments

type (
	// UserDataCollection contains the payments data disclosed to a user for GDPR/CCPA purposes.
	UserDataCollection struct {
		_ struct{} `json:"-"`

		Subscriptions       []Subscription       `json:"subscriptions,omitempty"`
		Purchases           []Purchase           `json:"purchases,omitempty"`
		PaymentTransactions []PaymentTransaction `json:"paymentTransactions,omitempty"`
	}
)
