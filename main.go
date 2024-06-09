package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var templates = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/login.html",
	"templates/signup.html",
	"templates/events.html",
	"templates/user.html",
	"templates/administrator.html",
	"templates/organizer.html",
	"templates/verify.html",
	"templates/promote.html",
	"templates/sponsors.html",
	"templates/viewevents.html",
))

var client *mongo.Client

type Event struct {
	Name      string `bson:"name"`
	Date      string `bson:"date"`
	Organizer string `bson:"organizer"`
	Venue     string `bson:"venue"`
	start     string `bson:"startime"`
	end       string `bson:"endtime"`
	budget    string `bson:"budget"`
}

type User struct {
	Username  string `bson:"username"`
	Password  string `bson:"password"`
	Email     string `bson:"email"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Phone     string `bson:"phone"`
	Gender    string `bson:"gender"`
	Role      string `bson:"role"`
}

type TempUser struct {
	User
	CaptchaCode string `bson:"captcha_code"`
}

func main() {
	// MongoDB Atlas connection string
	mongoURI := "mongodb+srv://jidaar718:tRelmEXYu7NEcGFz@cluster0.j9n5kuh.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"

	// Connect to MongoDB Atlas
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	// Ensure connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB Atlas!")
	fmt.Println("Server is running")

	// Set the random seed once at the start
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/events", eventsHandler)
	http.HandleFunc("/sponsors", sponsors)
	http.HandleFunc("/promote", promote)
	http.HandleFunc("/review", review)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func generateCaptchaCode() string {
	code := rand.Intn(999999)
	return fmt.Sprintf("%06d", code)
}

func sendCaptchaEmail(email, captchaCode string) error {
	from := "userstest323@gmail.com"
	password := "Abbas@0703"
	to := email
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Verification Code\n\n" +
		"Your verification code is: " + captchaCode

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Failed to send email:", err)
	}
	return err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("eventdb").Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var events []Event
	if err = cursor.All(ctx, &events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "index.html", events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		firstname := r.FormValue("firstname")
		lastname := r.FormValue("lastname")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		gender := r.FormValue("gender")
		role := r.FormValue("role")
		password := hashPassword(r.FormValue("password"))
		repeatPassword := hashPassword(r.FormValue("repeat_password"))

		// Ensure the passwords match
		if password != repeatPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		user := User{
			Username:  email,
			Password:  password,
			Email:     email,
			FirstName: firstname,
			LastName:  lastname,
			Phone:     phone,
			Gender:    gender,
			Role:      role,
		}

		if role == "administrator" || role == "organizer" {
			// Generate CAPTCHA code
			captchaCode := generateCaptchaCode()
			// Send CAPTCHA code to user's email
			err := sendCaptchaEmail(email, captchaCode)
			if err != nil {
				http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
				return
			}

			tempUser := TempUser{
				User:        user,
				CaptchaCode: captchaCode,
			}

			collection := client.Database("eventdb").Collection("temp_users")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err = collection.InsertOne(ctx, tempUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/verify?email="+email, http.StatusSeeOther)
		} else {
			collection := client.Database("eventdb").Collection("users")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := collection.InsertOne(ctx, user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
		return
	}
	if err := templates.ExecuteTemplate(w, "signup.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if r.Method == http.MethodPost {
		captchaCode := r.FormValue("captcha_code")

		collection := client.Database("eventdb").Collection("temp_users")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var tempUser TempUser
		err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&tempUser)
		if err != nil {
			http.Error(w, "Invalid verification code", http.StatusBadRequest)
			return
		}

		if tempUser.CaptchaCode != captchaCode {
			http.Error(w, "Invalid verification code", http.StatusBadRequest)
			return
		}

		userCollection := client.Database("eventdb").Collection("users")
		_, err = userCollection.InsertOne(ctx, tempUser.User)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = collection.DeleteOne(ctx, bson.M{"email": email})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := templates.ExecuteTemplate(w, "verify.html", email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		collection := client.Database("eventdb").Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		username := r.FormValue("username")
		password := hashPassword(r.FormValue("password"))

		var user User
		err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
		if err != nil || user.Password != password {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		switch user.Role {
		case "user":
			http.Redirect(w, r, "/user", http.StatusSeeOther)
		case "administrator", "organizer":
			http.Redirect(w, r, "/verify?email="+username, http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}
	if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// func eventsHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodPost {
// 		collection := client.Database("eventdb").Collection("events")
// 		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 		defer cancel()

// 		event := Event{
// 			Name: r.FormValue("name"),
// 			Date: r.FormValue("date"),
// 		}

//			_, err := collection.InsertOne(ctx, event)
//			if err != nil {
//				http.Error(w, err.Error(), http.StatusInternalServerError)
//				return
//			}
//		}
//		if err := templates.ExecuteTemplate(w, "events.html", nil); err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//		}
//	}
func eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Extract event details from the form
		event := Event{
			Name:      r.Form.Get("name"),
			Date:      r.Form.Get("date"),
			Organizer: r.Form.Get("organizer"),
			Venue:     r.Form.Get("venue"),
			start:     r.Form.Get("startime"),
			end:       r.Form.Get("endtime"),
			budget:    r.Form.Get("budget"),
		}

		// Insert event into the database
		collection := client.Database("eventdb").Collection("events")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = collection.InsertOne(ctx, event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to events page after successful insertion
		http.Redirect(w, r, "/events", http.StatusSeeOther)
		return
	}

	// If not a POST request, just render the events page
	if err := templates.ExecuteTemplate(w, "events.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sponsors(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "sponsors.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	{
		http.Redirect(w, r, "/sponsors", http.StatusSeeOther)
		return
	}
}

func promote(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "promote.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	{
		http.Redirect(w, r, "/promote", http.StatusSeeOther)
		return
	}
}

func review(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "review.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	{
		http.Redirect(w, r, "/review", http.StatusSeeOther)
		return
	}
}
