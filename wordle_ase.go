package main 
import (
 "fmt"
)

func main() {
  fmt.Println("Welcome to Hit and Blow!")
  fmt.Println("Please guess a 4-digit number (e.g., 1234):")
  // var correctNumber int 
  var userNumber string
  correctString := "qwer" // This is the correct number to guess

  for {
    fmt.Printf("Enter your guess: ")
    _,error := fmt.Scanf("%s",&userNumber)
    if error != nil {
      fmt.Println("Error reading input. Please try again.")
      continue
    }
    // Check if the input is a valid n-string
    if len(userNumber) == len(correctString) {
      fmt.Print("Input length matches! You entered: ", userNumber, "\n")
      break
    }
    fmt.Printf("Length mismatch. Please enter exactly %d characters.\n", len(correctString))
  }

  fmt.Println("Congratulations! You've guessed the number correctly!")
}
 