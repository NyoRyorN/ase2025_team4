package main
import (
    "fmt"
    "math/rand"
    "time"
	"os"
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
  const totalSeconds = 30 // Total time allowed for guessing
  var userString string
  correctString := RandomString(4) // This is the correct number to guess

  begin := time.Now()
  deadline := begin.Add(time.Duration(totalSeconds) * time.Second)

  timer := time.NewTicker(time.Duration(totalSeconds) * time.Second)
  defer timer.Stop()

  ticker := time.NewTicker(time.Second)
  defer ticker.Stop()

  go func() {
  	<-timer.C
    fmt.Printf("\r残り時間: 00秒\n")
	fmt.Println("Time's up! You didn't guess in time.")
	fmt.Println("The correct string was:", correctString)
	os.Exit(0)
  }()

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

  for i := 0; i < len(userString); i++ {
    // 毎秒、残り時間を上書き表示
    select {
	case <-ticker.C:
      remaining := time.Until(deadline)
      sec := int(remaining.Seconds())
      if sec < 0 {
        sec = 0
      }
        fmt.Printf("\r残り時間: %02d秒", sec)
      default:
        // ティッカーをブロックせずに次へ  
	}

	if i == 2 {
        timer.Stop()
        ticker.Stop()
        fmt.Println("\n🎉 Congratulations! You've guessed correctly!")
      } else {
        fmt.Println("\n❌ Wrong guess! Try again.")
      }
      break
	}
  fmt.Println(correctString)

  fmt.Println("Congratulations! You've guessed the number correctly!")
}
 
