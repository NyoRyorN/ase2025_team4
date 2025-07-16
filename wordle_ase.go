package main
import (
    "fmt"
    "math/rand"
    "time"
    "os"
    "encoding/json"
    "io"
    "net/http"
    "net/url"
    "strconv"
    "log"
)

func getCorrectString(desiredLength int)(correctString string){
  baseURL := "https://random-word-api.herokuapp.com/word"
	params := url.Values{}
	params.Add("length", strconv.Itoa(desiredLength))
	// 英語はデフォルトなので通常不要ですが、明示的に指定する場合は以下を追加
	// params.aaaAdd("lang", "en")

	fullURL := baseURL + "?" + params.Encode()

	// HTTP GETリクエストを送信
	resp, err := http.Get(fullURL)
	if err != nil {
    fmt.Println("Failed to send API request: %v\n")
		return
	}
	defer resp.Body.Close() // 関数終了時にレスポンスボディを閉じる

	// HTTPステータスコードの確認
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error response from API: Status code %d\n", resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for more details
		log.Fatalf("Response Body: %s\n", string(bodyBytes))
		return
	}

	// レスポンスボディを読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
    log.Fatalf("Failed to read response body: %v\n", err)
		return
	}
  var word []string
  json.Unmarshal(body, &word)

  return word[0]
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

func inputLength () (stringLength int){
  for {
    fmt.Printf("Enter your string length(max:10): ")
    _,err := fmt.Scanf("%d", &stringLength)
    if err != nil || stringLength < 1 || 10 < stringLength  {
			fmt.Println("No valid integer was detected. Please try again.")
			continue // エラーが発生したらループの先頭に戻り再入力を促す
    }
    return stringLength
  }
}

func main() {
  fmt.Println("Welcome to Hit and Blow!")
  // var correctNumber int 
  const totalSeconds = 30 // Total time allowed for guessing
  var userString string
  stringLength := inputLength()
  correctString := getCorrectString(stringLength) // This is the correct number to guess

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

  fmt.Println("Please guess a", stringLength ,"-letter string. Each letter should be a lowercase letter from 'a' to 'z'. (e.g., 'abcd').")
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