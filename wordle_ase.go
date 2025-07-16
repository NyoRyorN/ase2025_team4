package main
import (
    "fmt"
    "math/rand"
    "time"
)
 
const letters = "abcdefghijklmnopqrstuvwxyz"
 
func RandomString(n int) string{
    if n <= 0{
        return ""
    }
 
    rand.Seed(time.Now().UnixNano())
    b := make([]byte, n)
 
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
 
    return string(b)
}

func main() {
  fmt.Println("Welcome to Hit and Blow!")
  // var correctNumber int 
  var userString string
  correctString := RandomString(4) // This is the correct number to guess

  fmt.Println("Please guess a 4-letter string. Each letter should be a lowercase letter from 'a' to 'z'. (e.g., 'abcd').")
  for {
    fmt.Printf("Enter your guess: ")
    _,error := fmt.Scanf("%s", &userString)
    if error != nil {
      fmt.Println("Error reading input. Please try again.")
      continue
    }
    // Check if the input is a valid n-string
    if len(userString) == len(correctString) {
      fmt.Print("Input length matches! You entered: ", userString, "\n")
      break
    }
    fmt.Printf("Length mismatch. Please enter exactly %d characters.\n", len(correctString))
  }
  fmt.Println(correctString)

  fmt.Println("Congratulations! You've guessed the number correctly!")
}
 
