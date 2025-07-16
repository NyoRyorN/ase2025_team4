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

func hitAndBlow(userString, correctString string)( hits int,  blows int ){
  // Hit数とBlow数を計算
    correctStringMap := make(map[rune]int)
    for i ,char := range correctString {
      correctStringMap[char] = i
    }
    for i, char := range userString {
        if secretIndex, ok := correctStringMap[char]; ok {
            if i == secretIndex {
                hits++ // 桁と文字が両方一致
            } else {
                blows++ // 文字は一致するが桁が異なる
            }
        }
    }
    return hits, blows
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
    fmt.Printf("\nEnter your guess: ")
    _,error := fmt.Scanf("%s", &userString)
    if error != nil {
      fmt.Println("Error reading input. Please try again.")
      continue
    }
    // Check if the input is a valid n-string
    if len(userString) == len(correctString) {
      fmt.Print("Input length matches! You entered: ", userString, "\n")
      hits, blows := hitAndBlow(userString,correctString)
      fmt.Printf("\n結果: %d Hit(s), %d Blow(s)\n", hits, blows)
      if hits == len(correctString){
        timer.Stop()
        ticker.Stop()
		fmt.Println("\n🎉 Congratulations! You've guessed correctly!")
		break
      } else {
		fmt.Println("\n Wrong guess! Try again.")
	  }
    } else {
      fmt.Printf("Length mismatch. Please enter exactly %d characters.\n", len(correctString))
    }
  }
}

 
