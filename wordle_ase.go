package main
import (
    "fmt"
    "time"
    "os"
    "encoding/json"
    "io"
    "net/http"
    "net/url"
    "strconv"
    "log"
    "strings"
    "sync"
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


func hitAndBlow(userString, correctString string) (hits int, blows int) {
	// 文字列の長さを取得
	length := len(correctString)
	// runeスライスに変換してマルチバイト文字に対応
	userRunes := []rune(userString)
	correctRunes := []rune(correctString)

	// HitまたはBlowとしてカウントされた文字のインデックスを追跡
	correctUsed := make([]bool, length)
	userUsed := make([]bool, length)

	// 第1パス：Hitを計算
	for i := 0; i < length; i++ {
		if userRunes[i] == correctRunes[i] {
			hits++
			correctUsed[i] = true
			userUsed[i] = true
		}
	}

	// 第2パス：Blowを計算
	for i := 0; i < length; i++ {
		// すでにHitとしてカウントされたユーザーの文字はスキップ
		if userUsed[i] {
			continue
		}

		for j := 0; j < length; j++ {
			// すでにHitまたはBlowとしてカウントされた正解の文字はスキップ
			if correctUsed[j] {
				continue
			}

			if userRunes[i] == correctRunes[j] {
				blows++
				correctUsed[j] = true // この正解文字はもう使えない
				break                // このユーザー文字はマッチしたので、内側のループを抜ける
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

func askForHints() bool {
	for {
		fmt.Print("ヒントを有効にしますか？ (y/n): ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(response)
		if response == "y" {
			return true
		} else if response == "n" {
			return false
		}
		fmt.Println("無効な入力です。'y' または 'n' で入力してください。")
	}
}

func displayHint(correctString string, revealed []bool) {
	fmt.Print("ヒント: ")
	for i, char := range correctString {
		if revealed[i] {
			fmt.Printf("%c ", char)
		} else {
			fmt.Print("_ ")
		}
	}
	fmt.Println()
}

func main() {
  fmt.Println("Welcome to Hit and Blow!")
  // var correctNumber int 
  const totalSeconds = 30 // Total time allowed for guessing
  var userString string
  stringLength := inputLength()
  hintsEnabled := askForHints()
  correctString := getCorrectString(stringLength) // This is the correct number to guess

  begin := time.Now()
  deadline := begin.Add(time.Duration(totalSeconds) * time.Second)

  timer := time.NewTicker(time.Duration(totalSeconds) * time.Second)
  defer timer.Stop()

 	revealedLetters := make([]bool, stringLength)
	blowCharacters := make(map[rune]bool)
	var mu sync.Mutex

	var hintInterval time.Duration
	var nextHintTime time.Time
	revealedCount := 0

	if hintsEnabled {
		hintInterval = time.Duration(totalSeconds/stringLength) * time.Second
		nextHintTime = begin.Add(hintInterval)
	}

	// ゴルーチンで残り時間更新
	go func(hintsOn bool) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
        if hintsOn && time.Now().After(nextHintTime) && revealedCount < stringLength-1 {
					mu.Lock()
					hintGiven := false
					if len(blowCharacters) > 0 {
						for charToReveal := range blowCharacters {
							for i, correctChar := range correctString {
								if charToReveal == correctChar && !revealedLetters[i] && i < stringLength-1 {
									revealedLetters[i] = true
									revealedCount++
									hintGiven = true
									delete(blowCharacters, charToReveal)
									break
								}
							}
							if hintGiven {
								break
							}
						}
					}
					if !hintGiven {
						for i := 0; i < stringLength-1; i++ {
							if !revealedLetters[i] {
								revealedLetters[i] = true
								revealedCount++
								break
							}
						}
					}
					mu.Unlock()
					nextHintTime = nextHintTime.Add(hintInterval)
				}

				rem := int(time.Until(deadline).Seconds())
				if rem < 0 {
					return
				}

        mu.Lock()
				// 1) カーソル位置を保存
				fmt.Print("\0337")
				// 2) 上の行に移動し、行全体をクリア
        if hintsOn {
					fmt.Print("\033[2A")
					fmt.Print("\033[2K")
					displayHint(correctString, revealedLetters)
				} else {
				  fmt.Print("\033[1A")  // カーソルを1行上へ
        }
				fmt.Print("\033[2K")  // 行全体クリア
				// 3) 残り時間を表示し、改行
				fmt.Printf("残り時間: %02d秒\n", rem)
				// 4) カーソル位置を復元
				fmt.Print("\0338")
        mu.Unlock()

			case <-timer.C:

				// タイムアップ描画
        mu.Lock()
				if hintsOn {
					fmt.Print("\033[2A")
					fmt.Print("\033[2K")
					displayHint(correctString, revealedLetters)
				} else {
          fmt.Print("\033[1A")
        }
				fmt.Print("\033[2K")
				fmt.Println("残り時間: 00秒")
				fmt.Println("Time's up! You didn't guess in time.")
				fmt.Println("The correct string was:", correctString)
				os.Exit(0)
			}
		}
	}(hintsEnabled)

  if hintsEnabled {
		fmt.Println()
	}
  fmt.Println()

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
      if hintsEnabled {
				mu.Lock()
				correctStringMapForCheck := make(map[rune]bool)
				for _, c := range correctString {
					correctStringMapForCheck[c] = true
				}
				for i, userChar := range userString {
					if userChar == rune(correctString[i]) && i < stringLength-1 {
						if !revealedLetters[i] {
							revealedLetters[i] = true
							revealedCount++
						}
					} else if _, exists := correctStringMapForCheck[userChar]; exists {
						blowCharacters[userChar] = true
					}
				}
				mu.Unlock()
			}
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