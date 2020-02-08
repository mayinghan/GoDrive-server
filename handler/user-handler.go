package handler

import (
	"GoDrive/cache"
	"GoDrive/db"
	"GoDrive/utils"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

const salt = "&6ty"

// jwtKey is the key used to create the signature
var jwtKey = []byte("myhisaqt")

// Claims is a struct that is encoded to a jwt
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// LoginHandler handles user login.
func LoginHandler(c *gin.Context) {

	var userInput db.LoginInfo
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		panic(err)
	}

	fmt.Printf("%v\n", userInput)

	suc, msg, err := db.UserLogin(&userInput)

	if suc {
		//Create the expiration time (10 minutes) and the JWT claim
		expTime := time.Now().Add(10 * time.Minute)
		claims := &Claims{
			Username: userInput.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString(jwtKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  1,
				"msg":   "Internal server error: Failed to create JWT token.",
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  msg,
				"data": struct {
					Username string `json:"username"`
				}{
					Username: userInput.Username,
				},
			})
			c.SetCookie(
				"cookie",    //name
				tokenStr,    //value
				3600,        //max age
				"/",         //path
				"localhost", //domain
				true,        //secure
				true,        //httponly
			)
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   msg,
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code": 1,
			"msg":  msg,
		})
	}

	return
}

// RegisterHandler handles user registration. Method: POST
func RegisterHandler(c *gin.Context) {
	var regInput db.RegInfo
	if err := c.ShouldBindJSON(&regInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": 1,
		})

		return
	}

	fmt.Printf("%v\n", regInput)

	veriCode := regInput.Code
	rc := cache.EmailVeriPool().Get()
	code, err := redis.Uint64(rc.Do("HGET", regInput.Email, "code"))
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(500, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}
	fmt.Println(veriCode)
	if int64(code)-veriCode != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "Invalid verification code!",
		})
		return
	}
	// encrypt the password
	encryptedPwd := utils.MD5([]byte(regInput.Password + salt))
	regInput.Password = encryptedPwd
	suc, msg, err := db.UserRegister(&regInput)

	if suc {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  msg,
			"data": struct {
				Username string `json:"username"`
				Email    string `json:"email"`
			}{
				Username: regInput.Username,
				Email:    regInput.Email,
			},
		})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   msg,
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code": 1,
			"msg":  msg,
		})
	}
	return
}

// SendVerifyEmailHandler : send verify code to user email to finish registration
func SendVerifyEmailHandler(c *gin.Context) {
	type verifyEmail struct {
		Email string `json:"email" form:"email" binding:"required"`
	}

	var vrfEmail verifyEmail
	if err := c.ShouldBind(&vrfEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":   "Internal error happened",
			"code":  1,
			"error": err.Error(),
		})
		return
	}

	// get redis pool connection
	redisConn := cache.EmailVeriPool().Get()
	defer redisConn.Close()

	// check current user email
	currTimestamp := time.Now().UTC().Unix()
	storedTime, err := redis.Uint64(redisConn.Do("HGET", vrfEmail.Email, "create_at"))
	if err != nil {
		fmt.Printf("redis get previous created time failed %v\n", err)
		storedTime = 0
	}

	if storedTime != 0 && currTimestamp-int64(storedTime) < 20 {
		fmt.Println("dont send email again")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "Send request too fast! Please wait 90s to resend the code",
		})
		return
	}

	rand.Seed(currTimestamp)
	code := rand.Intn(899999) + 100000
	s := strconv.Itoa(code)
	redisConn.Do("HMSET", vrfEmail.Email, "create_at", currTimestamp, "code", code)
	// code expires after 10 min
	redisConn.Do("EXPIRE", vrfEmail.Email, 600)
	fmt.Println(s)
	err = utils.SendMail(vrfEmail.Email, s)
	if err != nil {
		panic(err)
	}
}
