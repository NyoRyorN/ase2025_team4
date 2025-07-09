package main 
import (
	"fmt"
	"regexp"
)

func main() {
	fmt.Println("Welcome to Hit and Blow!")
	fmt.Println("Please guess a 4-digit number (e.g., 1234):")
	// var correctNumber int 
	var userNumber string
	fmt.Scanf("%s",&userNumber)
	match,_:=regexp.MatchString(`\d{4}`,userNumber)
	if match != false && len(userNumber) != 4{}
	fmt.Println("Congratulations! You've guessed the number correctly!")
}
