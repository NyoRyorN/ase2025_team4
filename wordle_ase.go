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
      hits, blows := hitAndBlow(userString,correctString)
      fmt.Printf("結果: %d Hit(s), %d Blow(s)\n", hits, blows)
      if hits == len(correctString){
        break
      }
    }else{
    fmt.Printf("Length mismatch. Please enter exactly %d characters.\n", len(correctString))
    }
  }
  fmt.Println(correctString)

  fmt.Println("Congratulations! You've guessed the number correctly!")
}

func hitAndBlow (userString,correctString string)( hits int,  blows int ){
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
 
