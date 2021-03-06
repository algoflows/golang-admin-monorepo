package controllers

import (
	"strconv"
	"time"

	"go-admin/cmd/database"
	"go-admin/cmd/models"
	"go-admin/cmd/util"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

/*
	Register User
	desc: We take body/data from fiber context and store in map
				We use context.BodyParser to generate struct from stringyfied data
				Then we check if both passwords match
				Then hash the password using bcrypt.GenerateFromPassword passing in []byte
				Then create new user model from models.User struct
				Save new user into database using DB.Create(&user)
				Return newly created user

*/

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if data["password"] != data["password_confirm"] {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "Password doesn't match",
		})
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		FirstName: data["first_name"],
		LastName:  data["last_name"],
		Email:     data["email"],
		Password:  password,
	}

	database.DB.Create(&user)

	return c.JSON(user)
}

/*
 	Login User
	desc: We take the data map[string]string (key value string string)
				We use c.BodyParser(&data) to turn data into usable struct
				Then initialise empty user struct
				Then we check if user exists using email
				Then CompareHashandPassword using bcrypt package
				Then generate JWT using util.GenerateJwt convert userId to string
				Then construct cookie
				Then set cookie to context
				Return token to client

*/

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user models.User
	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.Id == 0 {
		c.Status(404)
		return c.JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.Status(401)
		return c.JSON(fiber.Map{
			"message": "Incorrect password",
		})
	}

	token, err := util.GenerateJwt(strconv.Itoa(int(user.Id)))

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	// Cookie sets a cookie by passing a cookie struct
	c.Cookie(&cookie)

	return c.JSON(token)
}

/*
 	User Claims - get generate token with user claims and return user
	desc: We take the jwt cookie from fiber context
				We pass cookie, jwt.StandardClaims struct, anon func with token to
				parseWithClaims function to get back token
				Then we check if claims and type cast token.claims.(*Claims)
				Then we filter where id == claims.Issuer return 1st result and store in user
				Then return c.JSON(user)
*/

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	id, _ := util.ParseJwt(cookie)

	var user models.User

	database.DB.Where("id = ?", id).First(&user)

	return c.JSON(user)
}

/*
	Logout User
	desc: We reset cookie values because we are not able to remove the cookie
				Set cookie date value to expired by 1 hour
				set c.cookie(&cookie)
*/

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}
