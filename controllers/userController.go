package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	client "go-jwt/clients"
	"go-jwt/intializers"
	"go-jwt/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/trace"
)
var tracer = otel.Tracer("pickeep-controller")
func SignUp(c *gin.Context) {
	// Get the User details (email/pass/name from the req body)
    
	var body models.User
	var user models.User

	ctx, span := tracer.Start(c.Request.Context(), "SignUp")
	defer span.End()

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	result := intializers.DB.Unscoped().Where("email=?", body.Email).First(&user) // Check whether the User is Soft deleted ? If not it will create new user
	fmt.Println(result)
	if result.RowsAffected == 0 {
		// Generate Unique String
		userIdentifier := uuid.New().String()
		//  Hash the Password
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to Hash Password",
			})
			return
		}

		// Create the S3 Bucket for each signup user
		Bucket, err := CreateBucket(userIdentifier)
		if err != nil {
			fmt.Println("Error creating bucket:", err)
			os.Exit(1)
		}

		fmt.Println("Bucket created successfully!")

		// Create the User
		user := models.User{Email: body.Email, Name: body.Name, Password: string(hash), UserIdentifier: userIdentifier, BucketName: Bucket, MobileNumber: body.MobileNumber}
		result := intializers.DB.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "User Already Exists"})
			return
		}
		// Send Email Notification about the Account Creation

		recipientEmail := "nitin.raavi@minfytech.com"
		if err := SendEmail(recipientEmail); err != nil {
			fmt.Println("Error sending email:", err)
			return
		}

		fmt.Println("Email sent successfully")

		// Respond
		c.JSON(http.StatusOK, gin.H{
			"message": "User Created Sucessfully"})
		// c.JSON(http.StatusOK, gin.H{
		// 	"user": gin.H{
		// 		"id":         user.ID,
		// 		"email":      user.Email,
		// 		"name":       user.Name,
		// 		"bucketName": user.BucketName,
		// 	}})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User Already Exists!"})
		return
	}

}

func SignIn(c *gin.Context) {
	// get the email n pass from the req body
	var body struct {
		Email    string
		Password string
	}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// lookup user details in db
	var user models.User
	result := intializers.DB.First(&user, "email=?", body.Email)
	// Check if user is available or not
	log.Println(result)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	// verify password hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password Not Matched",
		})
	}

	// Generate JWT Token
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to Create the Token",
		})
	}
	// Pass the Generated Token as Cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)

}
func Validate(c *gin.Context) {
	// Get the user from the middleware
	fmt.Println(c)
	user, _ := c.Get("user")
	//  return the required user details on Validation
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.(models.User).ID,
			"email": user.(models.User).Email,
			"name":  user.(models.User).Name,
		}})
}

func DeleteUser(c *gin.Context) {
	// Bind the request body
	var body struct {
		DeletedConsent string
	}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Intiate S3 Client
	s3client, err := client.S3client()
	if err != nil {
		log.Println("error while configuring the client", err)
	}

	// Get the user details from the DB
	user, _ := c.Get("user")
	if user.(models.User).DeletedAt.Valid {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s :User Already Deleted", user.(models.User).Email)})
	}

	bucketName := user.(models.User).BucketName

	// List all objects in the bucket
	listObjects, err := s3client.ListObjectsV2(c, &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unable to list objects in bucket, Here Why: %s", err.Error())})
		return
	}

	// Delete all objects in the bucket
	for _, obj := range listObjects.Contents {
		_, err := s3client.DeleteObject(c, &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    obj.Key,
		})
		if err != nil {
			log.Println("Unable to delete object", err)
			c.Abort()
			return
		}
	}

	// Delete the bucket
	_, err = s3client.DeleteBucket(c, &s3.DeleteBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		log.Println("Unable to Delete the Bucket, Here Why %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bucket"})
		return
	}

	// Soft delete the user from the database
	result := intializers.DB.Delete(&user)
	if result.Error != nil {
		log.Println("Unable to Delete the User, Here Why %w", err)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, "User Not Found")
		return
	}

	// // Create a new instance of User with updated DeletedConsent field
	// updatedUser := user.(models.User)
	// updatedUser.DeleteConsent = body.DeletedConsent

	// // Save the updated user to the database
	// result = intializers.DB.Save(&updatedUser)
	// if result.Error != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"user": fmt.Sprintf("User Account with EmailID:%s Deleted Sucessfully ", user.(models.User).Email)})
}

// API handler to re-enable a soft-deleted user
func ReenableUser(c *gin.Context) {
	// Parse the request body to get the user ID (email)
	var req struct {
		Email string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the soft-deleted user by ID
	var user models.User
	result := intializers.DB.Unscoped().Where("email=?", req.Email).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("cannot retrieve user %v\n", result.Error.Error())})
		return
	}

	// Check if the user is already restored
	if !user.DeletedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already restored"})
		return
	}

	// Create the S3 Bucket for re-enabled user
	Bucket, err := CreateBucket(user.UserIdentifier)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Bucket %s re-created Sucessfully", Bucket)})
	}
	if err != nil {
		fmt.Println("Error creating bucket:", err)
		os.Exit(1)
	}

	fmt.Println("Bucket re-created successfully!")

	// Update the user's DeletedAt field to NULL to restore the user

	result = intializers.DB.Unscoped().Model(&user).Update("deleted_at", gorm.Expr("NULL"))
	log.Println(result)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Return a success message
	c.JSON(http.StatusOK, gin.H{"message": "User restored successfully"})
}

// SendEmail sends an email using the Brevo Email API
func SendEmail(recipientEmail string) error {
	// API key obtained from Brevo
	apiKey := os.Getenv("SMTP_API_KEY")

	// Read the HTML file
	htmlContent, err := os.ReadFile("./emailTemplates/accountCreated.html")
	if err != nil {
		return fmt.Errorf("error reading HTML file: %v", err)
	}
	// Create request body
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "OhMyStorage!",
			"email": "nitinraavi167@gmail.com",
		},
		"to": []map[string]string{
			{
				"email": recipientEmail,
				"name":  "",
			},
		},
		"subject":     "Account Created! Welcome to PicKeep",
		"htmlContent": string(htmlContent),
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}
	fmt.Println(string(jsonPayload))

	// Create request to Brevo Email API
	url := "https://api.brevo.com/v3/smtp/email"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add API key to request headers
	req.Header.Set("api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error: %s", resp.Status)
	}

	return nil
}
