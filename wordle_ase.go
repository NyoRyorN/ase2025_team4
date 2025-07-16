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

	// ゴルーチンで残り時間更新
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rem := int(time.Until(deadline).Seconds())
				if rem < 0 {
					return
				}
				// 1) カーソル位置を保存
				fmt.Print("\0337")
				// 2) 上の行に移動し、行全体をクリア
				fmt.Print("\033[1A")  // カーソルを1行上へ
				fmt.Print("\033[2K")  // 行全体クリア
				// 3) 残り時間を表示し、改行
				fmt.Printf("残り時間: %02d秒\n", rem)
				// 4) カーソル位置を復元
				fmt.Print("\0338")
			case <-timer.C:
				// タイムアップ描画
				fmt.Print("\0337")
				fmt.Print("\033[1A")
				fmt.Print("\033[2K")
				fmt.Println("残り時間: 00秒")
				fmt.Println("Time's up! You didn't guess in time.")
				fmt.Println("The correct string was:", correctString)
				os.Exit(0)
			}
		}
	}()

  fmt.Println("Please guess a 4-letter string. Each letter should be a lowercase letter from 'a' to 'z'. (e.g., 'abcd').")
  for {
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
		fmt.Println("\n🎉 Congratulations! You've guessed correctly!")
		os.Exit(0)
      } else {
		fmt.Println("\n Wrong guess! Try again.")
	  }
    } else {
      fmt.Printf("Length mismatch. Please enter exactly %d characters.\n", len(correctString))
    }
  }
}

 
