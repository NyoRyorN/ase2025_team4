package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

type GameSession struct {
	CorrectString   string    `json:"correctString"`
	StringLength    int       `json:"stringLength"`
	StartTime       time.Time `json:"startTime"`
	GameStarted     bool      `json:"gameStarted"`
	GameOver        bool      `json:"gameOver"`
	Won             bool      `json:"won"`
	Guesses         []Guess   `json:"guesses"`
	HintsEnabled    bool      `json:"hintsEnabled"`
	RevealedLetters []bool    `json:"revealedLetters"`
	BlowCharacters  map[rune]bool `json:"blowCharacters"`
	RevealedCount   int       `json:"revealedCount"`
	NextHintTime    time.Time `json:"nextHintTime"`
}

type Guess struct {
	Input string `json:"input"`
	Hits  int    `json:"hits"`
	Blows int    `json:"blows"`
	// Time  string `json:"time"`
}

type GameRequest struct {
	Action       string `json:"action"`
	StringLength int    `json:"stringLength,omitempty"`
	Guess        string `json:"guess,omitempty"`
	HintsEnabled bool   `json:"hintsEnabled,omitempty"`
}

type GameResponse struct {
	Success       bool         `json:"success"`
	Message       string       `json:"message,omitempty"`
	GameSession   *GameSession `json:"gameSession,omitempty"`
	TimeRemaining int          `json:"timeRemaining,omitempty"`
	HintDisplay   string       `json:"hintDisplay,omitempty"`
}

var currentGame *GameSession
var mu sync.Mutex

func getCorrectString(desiredLength int) string {
	baseURL := "https://random-word-api.herokuapp.com/word"
	params := url.Values{}
	params.Add("length", strconv.Itoa(desiredLength))

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Failed to send API request: %v\n", err)
		return RandomString(desiredLength) // フォールバックとしてランダム文字列を使用
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response from API: Status code %d\n", resp.StatusCode)
		return RandomString(desiredLength) // フォールバックとしてランダム文字列を使用
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return RandomString(desiredLength) // フォールバックとしてランダム文字列を使用
	}

	var words []string
	err = json.Unmarshal(body, &words)
	if err != nil || len(words) == 0 {
		fmt.Printf("Failed to parse API response: %v\n", err)
		return RandomString(desiredLength) // フォールバックとしてランダム文字列を使用
	}

	return strings.ToLower(words[0])
}

func RandomString(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func hitAndBlow(userString, correctString string) (hits int, blows int) {
	length := len(correctString)
	userRunes := []rune(userString)
	correctRunes := []rune(correctString)

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
		if userUsed[i] {
			continue
		}

		for j := 0; j < length; j++ {
			if correctUsed[j] {
				continue
			}

			if userRunes[i] == correctRunes[j] {
				blows++
				correctUsed[j] = true
				break
			}
		}
	}

	return hits, blows
}

func startNewGame(length int, hintsEnabled bool) *GameSession {
	correctString := getCorrectString(length)
	fmt.Printf("新しいゲーム開始: 文字数=%d, 正解=%s, ヒント=%v\n", length, correctString, hintsEnabled)
	
	game := &GameSession{
		CorrectString:   correctString,
		StringLength:    length,
		StartTime:       time.Now(),
		GameStarted:     true,
		GameOver:        false,
		Won:             false,
		Guesses:         make([]Guess, 0),
		HintsEnabled:    hintsEnabled,
		RevealedLetters: make([]bool, length),
		BlowCharacters:  make(map[rune]bool),
		RevealedCount:   0,
	}

	if hintsEnabled {
		hintInterval := time.Duration(30/length) * time.Second
		game.NextHintTime = game.StartTime.Add(hintInterval)
		
		// ヒント更新のゴルーチンを開始
		go updateHints(game)
	}

	return game
}

func updateHints(game *GameSession) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for !game.GameOver {
		select {
		case <-ticker.C:
			if time.Now().After(game.NextHintTime) && game.RevealedCount < game.StringLength-1 {
				mu.Lock()
				hintGiven := false
				
				// まずBlow文字から優先してヒントを出す
				for charToReveal := range game.BlowCharacters {
					for i, correctChar := range game.CorrectString {
						if charToReveal == correctChar && !game.RevealedLetters[i] && i < game.StringLength-1 {
							game.RevealedLetters[i] = true
							game.RevealedCount++
							hintGiven = true
							delete(game.BlowCharacters, charToReveal)
							break
						}
					}
					if hintGiven {
						break
					}
				}

				// Blow文字がない場合は順番に公開
				if !hintGiven {
					for i := 0; i < game.StringLength-1; i++ {
						if !game.RevealedLetters[i] {
							game.RevealedLetters[i] = true
							game.RevealedCount++
							break
						}
					}
				}

				hintInterval := time.Duration(30/game.StringLength) * time.Second
				game.NextHintTime = game.NextHintTime.Add(hintInterval)
				mu.Unlock()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (gs *GameSession) isTimeUp() bool {
	return time.Since(gs.StartTime).Seconds() >= 30
}

func (gs *GameSession) getRemainingTime() int {
	remaining := 30 - int(time.Since(gs.StartTime).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (gs *GameSession) getHintDisplay() string {
	if !gs.HintsEnabled {
		return ""
	}
	
	var hint strings.Builder
	for i, char := range gs.CorrectString {
		if gs.RevealedLetters[i] {
			hint.WriteRune(char)
			hint.WriteString(" ")
		} else {
			hint.WriteString("_ ")
		}
	}
	return hint.String()
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req GameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var resp GameResponse

	switch req.Action {
	case "start":
		if req.StringLength < 1 || req.StringLength > 10 {
			resp = GameResponse{Success: false, Message: "文字数は1-10の間で選択してください"}
		} else {
			currentGame = startNewGame(req.StringLength, req.HintsEnabled)
			resp = GameResponse{
				Success:       true,
				Message:       fmt.Sprintf("ゲームを開始しました"),
				GameSession:   currentGame,
				TimeRemaining: currentGame.getRemainingTime(),
				HintDisplay:   currentGame.getHintDisplay(),
			}
		}

	case "guess":
		if currentGame == nil || !currentGame.GameStarted {
			resp = GameResponse{Success: false, Message: "ゲームが開始されていません"}
		} else if currentGame.GameOver {
			resp = GameResponse{Success: false, Message: "ゲームは終了しています"}
		} else if currentGame.isTimeUp() {
			currentGame.GameOver = true
			currentGame.Won = false
			fmt.Printf("時間切れ: 正解は %s でした\n", currentGame.CorrectString)
			resp = GameResponse{
				Success:       false,
				Message:       fmt.Sprintf("時間切れです！正解は: %s", currentGame.CorrectString),
				GameSession:   currentGame,
				TimeRemaining: 0,
			}
		} else {
			userInput := req.Guess

			if len(userInput) != currentGame.StringLength {
				resp = GameResponse{
					Success:       false,
					Message:       fmt.Sprintf("長さが違います。%d文字で入力してください。", currentGame.StringLength),
					GameSession:   currentGame,
					TimeRemaining: currentGame.getRemainingTime(),
					HintDisplay:   currentGame.getHintDisplay(),
				}
			} else {
				valid := true
				for _, char := range userInput {
					if char < 'a' || char > 'z' {
						valid = false
						break
					}
				}

				if !valid {
					resp = GameResponse{
						Success:       false,
						Message:       "小文字のアルファベット (a-z) のみ使用してください。",
						GameSession:   currentGame,
						TimeRemaining: currentGame.getRemainingTime(),
						HintDisplay:   currentGame.getHintDisplay(),
					}
				} else {
					// ヒント機能のための処理
					if currentGame.HintsEnabled {
						mu.Lock()
						correctStringMapForCheck := make(map[rune]bool)
						for _, c := range currentGame.CorrectString {
							correctStringMapForCheck[c] = true
						}
						for i, userChar := range userInput {
							if userChar == rune(currentGame.CorrectString[i]) && i < currentGame.StringLength-1 {
								if !currentGame.RevealedLetters[i] {
									currentGame.RevealedLetters[i] = true
									currentGame.RevealedCount++
								}
							} else if _, exists := correctStringMapForCheck[userChar]; exists {
								currentGame.BlowCharacters[userChar] = true
							}
						}
						mu.Unlock()
					}

					hits, blows := hitAndBlow(userInput, currentGame.CorrectString)
					guess := Guess{
						Input: userInput,
						Hits:  hits,
						Blows: blows,
						// Time:  time.Now().Format("15:04:05"),
					}
					currentGame.Guesses = append(currentGame.Guesses, guess)

					if hits == currentGame.StringLength {
						currentGame.GameOver = true
						currentGame.Won = true
						fmt.Printf("正解！答えは %s でした\n", currentGame.CorrectString)
						resp = GameResponse{
							Success:       true,
							Message:       "🎉 おめでとうございます！正解です！",
							GameSession:   currentGame,
							TimeRemaining: currentGame.getRemainingTime(),
							HintDisplay:   currentGame.getHintDisplay(),
						}
					} else {
						resp = GameResponse{
							Success:       true,
							Message:       "間違いです。もう一度試してください。",
							GameSession:   currentGame,
							TimeRemaining: currentGame.getRemainingTime(),
							HintDisplay:   currentGame.getHintDisplay(),
						}
					}
				}
			}
		}

	case "status":
		if currentGame == nil {
			resp = GameResponse{Success: true, Message: "ゲームが開始されていません"}
		} else if currentGame.isTimeUp() && !currentGame.GameOver {
			currentGame.GameOver = true
			currentGame.Won = false
			fmt.Printf("ステータス確認で時間切れ検出: 正解は %s でした\n", currentGame.CorrectString)
			resp = GameResponse{
				Success:       false,
				Message:       fmt.Sprintf("時間切れです！正解は: %s", currentGame.CorrectString),
				GameSession:   currentGame,
				TimeRemaining: 0,
			}
		} else {
			resp = GameResponse{
				Success:       true,
				GameSession:   currentGame,
				TimeRemaining: currentGame.getRemainingTime(),
				HintDisplay:   currentGame.getHintDisplay(),
			}
		}

	default:
		resp = GameResponse{Success: false, Message: "不明なアクション"}
	}

	json.NewEncoder(w).Encode(resp)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Hit and Blow Game</title>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
        }
        .game-setup {
            margin-bottom: 20px;
        }
        .setup-row {
            display: flex;
            align-items: center;
            gap: 15px;
            margin: 10px 0;
        }
        .timer {
            font-size: 18px;
            font-weight: bold;
            text-align: center;
            margin: 10px 0;
            color: #d32f2f;
        }
        .hint-display {
            font-size: 20px;
            font-weight: bold;
            text-align: center;
            margin: 15px 0;
            padding: 10px;
            background-color: #fff3e0;
            border-radius: 5px;
            font-family: monospace;
            color: #e65100;
        }
        .status {
            text-align: center;
            margin: 15px 0;
            padding: 10px;
            border-radius: 5px;
            font-weight: bold;
        }
        .success { background-color: #e8f5e8; color: #2e7d32; }
        .error { background-color: #ffebee; color: #c62828; }
        .info { background-color: #e3f2fd; color: #1565c0; }
        .input-group {
            display: flex;
            gap: 10px;
            margin: 20px 0;
        }
        input, select, button {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
        }
        input[type="text"] {
            flex: 1;
        }
        input[type="checkbox"] {
            transform: scale(1.2);
        }
        button {
            background-color: #1976d2;
            color: white;
            border: none;
            cursor: pointer;
            min-width: 100px;
        }
        button:hover {
            background-color: #1565c0;
        }
        button:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }
        .guesses {
            margin-top: 30px;
        }
        .guess-item {
            display: flex;
            justify-content: space-between;
            padding: 8px;
            margin: 5px 0;
            background-color: #f8f9fa;
            border-radius: 5px;
        }
        .guess-input {
            font-family: monospace;
            font-weight: bold;
        }
        .guess-result {
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎯 Hit and Blow Game</h1>
        
        <div class="game-setup">
            <div class="setup-row">
                <label>文字数を選択: 
                    <select id="lengthSelect">
                        <option value="3">3文字</option>
                        <option value="4" selected>4文字</option>
                        <option value="5">5文字</option>
                        <option value="6">6文字</option>
                        <option value="7">7文字</option>
                        <option value="8">8文字</option>
                    </select>
                </label>
            </div>
            <div class="setup-row">
                <label>
                    <input type="checkbox" id="hintsEnabled"> ヒントを有効にする
                </label>
            </div>
            <div class="setup-row">
                <button id="startBtn" onclick="startGame()">新しいゲーム開始</button>
            </div>
        </div>

        <div class="timer" id="timer">ゲームを開始してください</div>
        <div class="hint-display" id="hintDisplay" style="display: none;"></div>
        
        <div class="status info" id="status">「新しいゲーム開始」ボタンを押してください</div>

        <div class="input-group">
            <input type="text" id="guessInput" placeholder="推測を入力..." disabled>
            <button id="guessBtn" onclick="makeGuess()" disabled>推測</button>
        </div>

        <div class="guesses" id="guesses" style="display: none;">
            <h3>推測履歴</h3>
            <div id="guessList"></div>
        </div>
    </div>

    <script>
        let gameTimer;
        let statusCheckTimer;

        document.getElementById('guessInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                makeGuess();
            }
        });

        function startGame() {
            const length = parseInt(document.getElementById('lengthSelect').value);
            const hintsEnabled = document.getElementById('hintsEnabled').checked;
            
            fetch('/api', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({action: 'start', stringLength: length, hintsEnabled: hintsEnabled})
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    document.getElementById('guessInput').disabled = false;
                    document.getElementById('guessBtn').disabled = false;
                    document.getElementById('startBtn').disabled = true;
                    document.getElementById('lengthSelect').disabled = true;
                    document.getElementById('hintsEnabled').disabled = true;
                    document.getElementById('guessInput').focus();
                    document.getElementById('guesses').style.display = 'none';
                    document.getElementById('guessList').innerHTML = '';
                    
                    updateStatus(data.message, 'success');
                    updateHintDisplay(data.hintDisplay, hintsEnabled);
                    startTimer();
                }
            });
        }

        function makeGuess() {
            const guess = document.getElementById('guessInput').value.toLowerCase();
            
            fetch('/api', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({action: 'guess', guess: guess})
            })
            .then(response => response.json())
            .then(data => {
                updateStatus(data.message, data.success ? 'success' : 'error');
                updateHintDisplay(data.hintDisplay, document.getElementById('hintsEnabled').checked);
                
                if (data.gameSession && data.gameSession.guesses.length > 0) {
                    updateGuessList(data.gameSession.guesses);
                    document.getElementById('guesses').style.display = 'block';
                }
                
                if (data.gameSession && data.gameSession.gameOver) {
                    endGame();
                }
                
                document.getElementById('guessInput').value = '';
            });
        }

        function updateGuessList(guesses) {
            const guessList = document.getElementById('guessList');
            guessList.innerHTML = '';
            
            guesses.forEach(guess => {
                const div = document.createElement('div');
                div.className = 'guess-item';
                div.innerHTML = 
                    '<span class="guess-input">' + guess.input + '</span>' +
                    '<span class="guess-result">' + guess.hits + ' Hit(s), ' + guess.blows + ' Blow(s)</span>';
                guessList.appendChild(div);
            });
        }

        function updateStatus(message, type) {
            const status = document.getElementById('status');
            status.textContent = message;
            status.className = 'status ' + type;
        }

        function updateHintDisplay(hintText, hintsEnabled) {
            const hintDisplay = document.getElementById('hintDisplay');
            if (hintsEnabled && hintText) {
                hintDisplay.textContent = 'ヒント: ' + hintText;
                hintDisplay.style.display = 'block';
            } else {
                hintDisplay.style.display = 'none';
            }
        }

        function startTimer() {
            let timeLeft = 30;
            
            gameTimer = setInterval(() => {
                timeLeft--;
                document.getElementById('timer').textContent = '残り時間: ' + timeLeft.toString().padStart(2, '0') + '秒';
                
                if (timeLeft <= 0) {
                    clearInterval(gameTimer);
                    checkGameStatus();
                }
            }, 1000);
            
            statusCheckTimer = setInterval(checkGameStatus, 1000);
        }

        function checkGameStatus() {
            fetch('/api', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({action: 'status'})
            })
            .then(response => response.json())
            .then(data => {
                if (data.timeRemaining !== undefined) {
                    document.getElementById('timer').textContent = '残り時間: ' + data.timeRemaining.toString().padStart(2, '0') + '秒';
                }
                
                updateHintDisplay(data.hintDisplay, document.getElementById('hintsEnabled').checked);
                
                if (data.gameSession && data.gameSession.gameOver) {
                    if (!data.success && data.message) {
                        updateStatus(data.message, 'error');
                    }
                    endGame();
                } else if (!data.success && data.message) {
                    updateStatus(data.message, 'error');
                    endGame();
                }
            });
        }

        function endGame() {
            clearInterval(gameTimer);
            clearInterval(statusCheckTimer);
            document.getElementById('guessInput').disabled = true;
            document.getElementById('guessBtn').disabled = true;
            document.getElementById('startBtn').disabled = false;
            document.getElementById('lengthSelect').disabled = false;
            document.getElementById('hintsEnabled').disabled = false;
        }
    </script>
</body>
</html>`

	t, _ := template.New("index").Parse(tmpl)
	t.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api", handleAPI)

	fmt.Println("Hit and Blow Game サーバーを開始します...")
	fmt.Println("ブラウザで http://localhost:8081 にアクセスしてください")
	
	log.Fatal(http.ListenAndServe(":8081", nil))
}