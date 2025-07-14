package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/oklog/ulid"

	"github.com/datazip/olake-frontend/server/internal/models"
)

func ToMapOfInterface(structure any) map[string]interface{} {
	if structure == nil {
		return nil
	}

	data, _ := json.Marshal(structure)

	var output map[string]interface{}
	_ = json.Unmarshal(data, &output)

	return output
}

func RespondJSON(ctx *web.Controller, status int, success bool, message string, data interface{}) {
	ctx.Ctx.Output.SetStatus(status)
	ctx.Data["json"] = models.JSONResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	_ = ctx.ServeJSON()
}

func SuccessResponse(ctx *web.Controller, data interface{}) {
	RespondJSON(ctx, http.StatusOK, true, "success", data)
}

func ErrorResponse(ctx *web.Controller, status int, message string) {
	RespondJSON(ctx, status, false, message, nil)
}

func HandleJSONOK(w http.ResponseWriter, content interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(content)
}

// send a message as response
func HandleResponseMessage(w http.ResponseWriter, statusCode int, content interface{}, message string) {
	body := make(map[string]interface{})

	if content != nil {
		jsonbody, err := json.Marshal(content)
		if err != nil {
			HandleError(w, http.StatusInternalServerError, err)
		}

		if err = json.Unmarshal(jsonbody, &body); err != nil {
			HandleError(w, http.StatusInternalServerError, err)
		}
	}
	body["message"] = message

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

// send error as json response
func HandleErrorAsMessage(w http.ResponseWriter, statusCode int, err error) {
	body := make(map[string]string)
	body["error"] = err.Error()

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// send error as direct text/string
func HandleError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, err)
}

// Handle errors and pass it to /error page
func HandleErrorJS(w http.ResponseWriter, r *http.Request, err error) {
	http.Redirect(w, r, fmt.Sprintf(`/error?msg=%q`, url.QueryEscape(err.Error())), http.StatusPermanentRedirect)
}

// // Encrypt with AES encryption and returns base64 encoded string
// func encryptAES(content, key []byte) (string, error) {
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return "", err
// 	}
// 	cipherText := make([]byte, aes.BlockSize+len(content))
// 	iv := cipherText[:aes.BlockSize]
// 	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
// 		return "", err
// 	}
// 	stream := cipher.NewCFBEncrypter(block, iv)
// 	stream.XORKeyStream(cipherText[aes.BlockSize:], content)
// 	return base64.StdEncoding.EncodeToString(cipherText), err
// }

// // decryptAES decrypts a base64 encoded string using AES encryption
// func decryptAES(secure string, key []byte) ([]byte, error) {
// 	cipherDecoded, err := base64.StdEncoding.DecodeString(secure)
// 	// if DecodeString failed, exit:
// 	if err != nil {
// 		return nil, err
// 	}
// 	// create a new AES cipher with the key and encrypted message
// 	block, err := aes.NewCipher(key)
// 	// if NewCipher failed, exit:
// 	if err != nil {
// 		return nil, err
// 	}
// 	// if the length of the cipherDecoded is less than 16 Bytes:
// 	if len(cipherDecoded) < aes.BlockSize {
// 		logs.Error("cipherDecoded block size is too short!")
// 		return nil, err
// 	}
// 	iv := cipherDecoded[:aes.BlockSize]
// 	cipherDecoded = cipherDecoded[aes.BlockSize:]
// 	// decrypt the message
// 	stream := cipher.NewCFBDecrypter(block, iv)
// 	stream.XORKeyStream(cipherDecoded, cipherDecoded)
// 	return cipherDecoded, nil
// }

func ExistsInArray[T comparable](arr []T, value T) bool {
	for _, elem := range arr {
		if elem == value {
			return true
		}
	}

	return false
}

func ULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)

	t := time.Now()
	newUlid, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		logs.Critical(err)
	}

	return newUlid.String()
}

// ConnectionStatus represents the structure of the connection status JSON
type ConnectionStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// LogMessage represents the structure of the log message JSON
type LogMessage struct {
	ConnectionStatus *ConnectionStatus `json:"connectionStatus,omitempty"`
	Type             string            `json:"type,omitempty"`
	// Add other fields as needed
}

// ExtractAndParseLastLogMessage extracts the JSON from the last log line and parses it
func ExtractAndParseLastLogMessage(output []byte) (*LogMessage, error) {
	// Convert output to string and split into lines
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	lines := strings.Split(outputStr, "\n")

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			lastLine = trimmed
			break
		}
	}

	if lastLine == "" {
		return nil, fmt.Errorf("no log lines found")
	}

	// Extract JSON part (everything after the first "{")
	start := strings.Index(lastLine, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON found in log line")
	}
	jsonStr := lastLine[start:]

	// Parse the JSON
	var msg LogMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &msg, nil
}

// CreateDirectory creates a directory with the specified permissions if it doesn't exist
func CreateDirectory(dirPath string, perm os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
	}
	return nil
}

// WriteFile writes data to a file, creating the directory if necessary
func WriteFile(filePath string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filePath)
	if err := CreateDirectory(dirPath, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, perm); err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}
	return nil
}

// ParseJSONFile parses a JSON file into a map
func ParseJSONFile(filePath string) (map[string]interface{}, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %v", filePath, err)
	}

	return result, nil
}
