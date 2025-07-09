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
    fmt.Println("Please guess a 4-digit number (e.g., 1234):")
    // var correctNumber int
    var userNumber string
    fmt.Scanf("%s",&userNumber)
    match,_:=regexp.MatchString(`\d{4}`,userNumber)
    if match != false && len(userNumber) != 4{}
	fmt.Println(RandomString(8))
    fmt.Println("Congratulations! You've guessed the number correctly!")
}