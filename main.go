package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/sirupsen/logrus"
)

func main() {
	lambda.Start(handler)
}

//Handle handles the step function.
func handler(ctx context.Context, event awsevents.CognitoEventUserPoolsPreSignup) (awsevents.CognitoEventUserPoolsPreSignup, error) {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	if event.TriggerSource == "PreSignUp_ExternalProvider" {
		//Set up a new session.
		conf := &aws.Config{
			Region: aws.String("eu-west-2"),
		}

		s, err := session.NewSession(conf)
		if err != nil {
			logrus.WithError(err).Info("failed to create session")
			return event, err
		}

		svc := cognitoidentityprovider.New(s)

		users, err := svc.ListUsers(&cognitoidentityprovider.ListUsersInput{
			UserPoolId: aws.String(event.UserPoolID),
			Filter:     aws.String(fmt.Sprintf(`email = "%s"`, event.Request.UserAttributes["email"])),
		})
		if err != nil {
			logrus.WithError(err).Info("failed to get users")
			return event, err
		}

		if len(users.Users) == 0 {
			logrus.Info("Found no Users")
			return event, nil
		}

		if len(users.Users) != 1 {
			err = fmt.Errorf("There was %d users returned", len(users.Users))
			logrus.WithError(err).Info("More than one user returned")
			return event, err
		}

		splitStrs := strings.Split(event.UserName, "_")

		_, err = svc.AdminLinkProviderForUser(&cognitoidentityprovider.AdminLinkProviderForUserInput{
			UserPoolId: aws.String(event.UserPoolID),
			DestinationUser: &cognitoidentityprovider.ProviderUserIdentifierType{
				ProviderName:           aws.String("Cognito"),
				ProviderAttributeValue: users.Users[0].Username,
			},
			SourceUser: &cognitoidentityprovider.ProviderUserIdentifierType{
				ProviderAttributeName:  aws.String("Cognito_Subject"),
				ProviderName:           aws.String(strings.Title(splitStrs[0])),
				ProviderAttributeValue: aws.String(splitStrs[1]),
			},
		})
		if err != nil {
			logrus.WithError(err).Info("failed link users")
			return event, err
		}
		logrus.WithField("event", event).Info("event")
		return event, errors.New("Do not create the new user.")
	} else {
		event.Response.AutoConfirmUser = true
		event.Response.AutoVerifyEmail = true
	}

	return event, nil
}
