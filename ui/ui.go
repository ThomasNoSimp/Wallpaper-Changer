package ui

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	signUpContainer       fyne.CanvasObject
	codeEntry             *widget.Entry
	wallpaperPaths        = []string{"ui/assets/001.jpg", "ui/assets/002.jpeg"}
	currentWallpaperIndex = 0
	wallpaperContainer    *fyne.Container
)

// SMTPServer represents the SMTP server details
type SMTPServer struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoadConfig loads SMTP server configuration from a JSON file
func LoadConfig() (SMTPServer, error) {
	var config struct {
		SMTPServer SMTPServer `json:"smtp_server"`
	}

	configFile, err := os.Open("ui/config.json")
	if err != nil {
		return SMTPServer{}, err
	}
	defer configFile.Close()

	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return SMTPServer{}, err
	}

	return config.SMTPServer, nil
}

// SetupUI sets up the user interface
func SetupUI(w fyne.Window) {
	// Initialize wallpaperContainer
	wallpaperContainer = container.NewVBox()

	// Create label
	label := widget.NewLabel("Welcome from Wallpaper Changer by Thomas!")

	// Create button
	testButton := widget.NewButton("Test", func() {
		homepage(w)
	})
	button := widget.NewButton("Sign Up", func() {
		ShowSignUpForm(w)
	})

	// Create a box layout
	box := container.NewVBox(
		label,
		button,
		testButton,
	)

	// Set content of window
	w.SetContent(box)

	// Set custom window size
	setCustomWindowSize(w)
}

// ShowSignUpForm displays the signup form
func ShowSignUpForm(w fyne.Window) {
	// Create form elements
	nameEntry := widget.NewEntry()
	emailEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry := widget.NewPasswordEntry()
	signUpButton := widget.NewButton("Sign Up", func() {
		if validateEmail(emailEntry.Text) {
			if passwordEntry.Text == confirmPasswordEntry.Text {
				if validatePassword(passwordEntry.Text) {
					// Process signup logic here
					signUpContainer.Hide()
					// Send email to the input email
					err := sendEmail(emailEntry.Text)
					if err != nil {
						log.Println("Error sending email:", err)
					} else {
						log.Println("Email sent successfully!")
						showVerificationCodeInput(w)
					}
				} else {
					// Display error message if password is invalid
					widget.NewLabel("Password must be at least 8 characters long").Show()
				}
			} else {
				// Display error message if passwords don't match
				widget.NewLabel("Passwords do not match").Show()
			}
		} else {
			// Display error message if email is invalid
			widget.NewLabel("Invalid email address").Show()
		}
	})

	// Create a form layout
	form := container.NewVBox(
		widget.NewLabel("Name:"),
		nameEntry,
		widget.NewLabel("Email:"),
		emailEntry,
		widget.NewLabel("Create a Password:"),
		passwordEntry,
		widget.NewLabel("Confirm Password:"),
		confirmPasswordEntry,
		signUpButton,
	)

	// Set content of signup container
	signUpContainer = form

	// Replace current window content with signup form
	w.SetContent(signUpContainer)
}

// showVerificationCodeInput displays the code input box
func showVerificationCodeInput(w fyne.Window) {
	codeEntry = widget.NewEntry()
	verifyButton := widget.NewButton("Verify", func() {
		if validateCode(codeEntry.Text) {
			dialog.ShowInformation("Verification Successful", "Code verified successfully!", w)
			homepage(w)
		} else {
			dialog.ShowError(errors.New("Invalid verification code"), w)
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Verification Code:"),
		codeEntry,
		verifyButton,
	)

	w.SetContent(form)
}

// setCustomWindowSize sets a custom window size
func setCustomWindowSize(w fyne.Window) {
	// Set custom window size (width, height)
	w.Resize(fyne.NewSize(float32(1000), float32(800)))
}

// validateEmail validates the format of an email address
func validateEmail(email string) bool {
	// Regular expression for email validation
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}

// validatePassword validates the format of a password
func validatePassword(password string) bool {
	// Password must be at least 8 characters long
	return len(password) >= 8
}

// Global variable to store the generated verification code
var verificationCode string

// sendEmail sends an email to the specified email address
func sendEmail(email string) error {
	// Seed the random number generator with current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random 6-digit code
	code := generateRandomCode(100000, 999999)

	// Store the generated code globally
	verificationCode = strconv.Itoa(code)

	// Load SMTP server configuration from config file
	serverConfig, err := LoadConfig()
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", serverConfig.Username, serverConfig.Password, serverConfig.Host)

	to := []string{email}
	// Concatenate the code with the email message
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Wallpaper Changer Email Verification!\r\n" +
		"\r\n" +
		"This is your unique verification code for Wallpaper Changer app. Code: " + verificationCode + "\r\n")

	err = smtp.SendMail(serverConfig.Host+":"+serverConfig.Port, auth, serverConfig.Username, to, msg)
	if err != nil {
		return err
	}
	return nil
}

// validateCode validates the entered code
func validateCode(code string) bool {
	// Compare the entered code with the generated code
	return code == verificationCode
}

// Function to generate a random number between min and max
func generateRandomCode(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// Function to set the wallpaper using AppleScript
func setWallpaperUsingAppleScript(imagePath string) error {
	// Get the full path of the image
	fullImagePath, err := getFullPath(imagePath)
	if err != nil {
		return err
	}

	// Construct the AppleScript command to set the wallpaper
	script := `
        set wallpaperImage to "` + fullImagePath + `"
        tell application "System Events"
            set picture of every desktop to wallpaperImage
        end tell
    `
	// Log the constructed AppleScript command
	log.Println("AppleScript command:", script)

	// Execute the AppleScript command using osascript
	cmd := exec.Command("osascript", "-e", script)
	err = cmd.Run()
	if err != nil {
		// Log the error for debugging
		log.Println("Error setting wallpaper:", err)
		return err
	}
	return nil
}

// Function to get the full path of a file
func getFullPath(relativePath string) (string, error) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Construct the full path by joining the current working directory and the relative path
	fullPath := filepath.Join(wd, relativePath)
	return fullPath, nil
}

// homepage function to create UI components
func homepage(w fyne.Window) {
	// Set the current wallpaper index to 0 (corresponding to 001.jpg)
	currentWallpaperIndex = 0

	// Create UI components for the homepage
	label := widget.NewLabel("Wallpaper Changer")

	// Horizontally center the label
	labelContainer := container.NewCenter(label)

	// Create Previous button
	previousButton := widget.NewButton("Previous", func() {
		// Decrement the current wallpaper index
		currentWallpaperIndex--
		if currentWallpaperIndex < 0 {
			currentWallpaperIndex = len(wallpaperPaths) - 1
		}
		// Load the previous wallpaper image
		newWallpaperImage, err := loadImageFromFileWithSize(wallpaperPaths[currentWallpaperIndex], 400, 400)
		if err != nil {
			log.Println("Error loading image:", err)
			return
		}
		// Replace the current wallpaper image with the new one
		wallpaperContainer.Objects = []fyne.CanvasObject{newWallpaperImage}
		wallpaperContainer.Refresh()
	})

	// Create Next button
	nextButton := widget.NewButton("Next", func() {
		// Increment the current wallpaper index
		currentWallpaperIndex++
		if currentWallpaperIndex >= len(wallpaperPaths) {
			currentWallpaperIndex = 0
		}
		// Load the next wallpaper image
		newWallpaperImage, err := loadImageFromFileWithSize(wallpaperPaths[currentWallpaperIndex], 400, 400)
		if err != nil {
			log.Println("Error loading image:", err)
			return
		}
		// Replace the current wallpaper image with the new one
		wallpaperContainer.Objects = []fyne.CanvasObject{newWallpaperImage}
		wallpaperContainer.Refresh()
	})

	// Create Set Wallpaper button
	setWallpaperButton := widget.NewButton("Set Wallpaper", func() {
		// Get the path of the current wallpaper image
		currentWallpaperPath := wallpaperPaths[currentWallpaperIndex]

		// Set the wallpaper using AppleScript
		err := setWallpaperUsingAppleScript(currentWallpaperPath)
		if err != nil {
			log.Println("Error setting wallpaper:", err)
			return
		}
		dialog.ShowInformation("Wallpaper Set", "The wallpaper has been set as the desktop background.", w)
	})

	// Create a horizontal box layout to contain Previous, Next, and Set Wallpaper buttons
	buttonRow := container.NewHBox(
		widget.NewLabel(""), // Empty label to create space
		previousButton,
		widget.NewLabel(""), // Empty label to create space
		nextButton,
		widget.NewLabel(""), // Empty label to create space
		setWallpaperButton,
	)

	// Create a vertical box layout to contain the label, wallpaper image, and buttons
	box := container.NewVBox(
		labelContainer,     // Use labelContainer here
		wallpaperContainer, // Use wallpaperContainer here
		container.NewCenter(container.NewHBox(buttonRow)), // Center the buttons horizontally
	)

	// Set content of the window to the homepage layout
	w.SetContent(box)
}

// Function to load an image from file with custom size
func loadImageFromFileWithSize(path string, width, height int) (fyne.CanvasObject, error) {
	// Create an image from the file
	img := canvas.NewImageFromFile(path)

	// Set the fill mode to ImageFillContain to preserve the aspect ratio and fit the image inside the specified dimensions
	img.FillMode = canvas.ImageFillContain

	// Set the size constraints on the image
	img.SetMinSize(fyne.NewSize(float32(width), float32(height)))
	img.Resize(fyne.NewSize(float32(width), float32(height)))
	return img, nil
}
