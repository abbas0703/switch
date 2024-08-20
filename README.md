# switch
# Event Management System

This is a web-based Event Management System built using Go and MongoDB. The system allows users to sign up, log in, and manage events. It also includes functionalities for verifying users via email and handling roles such as administrator, organizer, and user.

## Features

- **User Authentication**: Secure signup and login using SHA-256 hashed passwords.
- **Role-based Access Control**: Different access levels for administrators, organizers, and general users.
- **Event Management**: Users can create and view events. Administrators and organizers have additional permissions.
- **Email Verification**: Sends a CAPTCHA code to the user's email during signup for verification.
- **MongoDB Integration**: Stores user and event data in a MongoDB Atlas database.

## Prerequisites

- Go 1.16 or higher
- MongoDB Atlas account
- An SMTP server (e.g., Gmail) for sending verification emails

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/event-management-system.git
    cd event-management-system
    ```

2. Install dependencies:

    - Go modules are managed automatically. Just run the Go server and dependencies will be downloaded automatically.

3. Set up MongoDB Atlas:

    - Replace the `mongoURI` variable in the `main.go` file with your MongoDB Atlas connection string.

4. Configure the SMTP server:

    - In the `sendCaptchaEmail` function, replace the `from`, `password`, `smtpHost`, and `smtpPort` variables with your SMTP server details.

5. Run the server:

    ```bash
    go run main.go
    ```

6. Access the application:

    - Open your web browser and navigate to `http://localhost:8080`.

## Folder Structure

```plaintext
.
├── templates
│   ├── index.html
│   ├── login.html
│   ├── signup.html
│   ├── events.html
│   ├── user.html
│   ├── administrator.html
│   ├── organizer.html
│   ├── verify.html
│   ├── promote.html
│   ├── sponsors.html
│   └── viewevents.html
├── static
│   └── (static files such as CSS, JavaScript, images)
├── main.go
└── README.md

