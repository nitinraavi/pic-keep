package client

import (
	"context"
	"fmt"
	"go-jwt/models"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
)

const (
	clientID          = "************"
	userPoolID string = "*********"
	Username   string = "*********"
	Password   string = "*******"
)

func Auth(c *gin.Context) {
	var body models.User
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	// Authenticate user
	clientID := clientID
	resp, err := client.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: "USER_PASSWORD_AUTH",
		AuthParameters: map[string]string{
			"USERNAME": body.Email,
			"PASSWORD": body.Password,
		},
		ClientId: &clientID,
	})
	if err != nil {
		panic("unable to authenticate user, " + err.Error())
	}

	fmt.Println("Authentication successful!")

	// // Extract tokens
	// accessToken := resp.AuthenticationResult.AccessToken
	// idToken := resp.AuthenticationResult.IdToken
	// refreshToken := resp.AuthenticationResult.RefreshToken
	c.JSON(http.StatusOK, resp)
	// fmt.Println("Access Token:", accessToken)
	// fmt.Println("Id Token:", idToken)
	// fmt.Println("Refresh Token:", refreshToken)
}
