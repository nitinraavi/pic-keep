package middleware

import (
	"fmt"
	"go-jwt/intializers"
	"go-jwt/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequireAuth(c *gin.Context) {

	// Get the cookie from req
	tokenString, err := c.Cookie("Authorization")
	fmt.Println("Token Found")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	// Decode and Validate it
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	fmt.Println("Token Docoded")
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the expiry time
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Find the user with token claims (sub)
		var user models.User

		result := intializers.DB.Select("id, name, email, user_identifier, bucket_name").First(&user, claims["sub"])
		if result.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "User not found")
			return
		}

		if user.DeletedAt.Valid {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "User already deleted")
			return
		}

		// Attach to req

		c.Set("user", user)

		// Continue
		c.Next()
	} else {
		fmt.Println(err)
	}

}
