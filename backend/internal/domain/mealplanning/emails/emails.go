package email

import (
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/branding"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v7/email"

	"github.com/matcornic/hermes/v2"
)

var (
	ErrUnverifiedEmailRecipient = errors.New("missing email address verification for user")
)

// BuildMealPlanCreatedEmail builds an email notifying a user that they've been invited to join an account.
func BuildMealPlanCreatedEmail(recipient *identity.User, mealPlan *mealplanning.MealPlan, baseURL string) (*email.OutboundEmailMessage, error) {
	if recipient.EmailAddressVerifiedAt == nil {
		return nil, ErrUnverifiedEmailRecipient
	}

	isElectionMealPlan := false
	for _, event := range mealPlan.Events {
		if len(event.Options) > 1 {
			isElectionMealPlan = true
		}
	}

	buttonAction := "Check it out"
	instructions := "You can see what's up for dinner by clicking the button below"
	if isElectionMealPlan {
		instructions = "You can rank each meal in the meal plan by clicking the button below"
		buttonAction = "Submit your vote"
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.FirstName,
			Intros: []string{
				"A new meal plan has been created for your account!",
			},
			Actions: []hermes.Action{
				{
					Instructions: instructions,
					Button: hermes.Button{
						Text: buttonAction,
						Link: fmt.Sprintf("%s/meal_plans/%s", baseURL, mealPlan.ID),
					},
				},
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		ToAddress:   recipient.EmailAddress,
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     "A new meal plan has been created!",
		HTMLContent: htmlContent,
	}

	return msg, nil
}
