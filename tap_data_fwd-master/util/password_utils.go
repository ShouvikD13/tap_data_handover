package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

type PasswordUtilManger struct {
	LM *LoggerManager
}

func (PUM *PasswordUtilManger) FngenerateNewPassword(oldPassword string) (string, error) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters := "abcdefghijklmnopqrstuvwxy"
	numbers := "123456789"
	specialChars := "@/#$%&*"

	passwordLength := CRNT_PSSWD_LEN
	passwordLength -= 3

	var password strings.Builder

	for i := 0; i < passwordLength-2; i++ {
		char, err := PUM.FnrandomVal(letters)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	char, err := PUM.FnrandomVal(lowerLetters)
	if err != nil {
		return "", err
	}
	password.WriteByte(char)

	char, err = PUM.FnrandomVal(specialChars)
	if err != nil {
		return "", err
	}
	password.WriteByte(char)

	for i := 0; i < 2; i++ {
		char, err = PUM.FnrandomVal(numbers)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	shuffledPassword := PUM.FnshuffleString(password.String())

	if shuffledPassword == oldPassword {
		return PUM.FngenerateNewPassword(oldPassword) // Recursively generate a new password
	}

	return shuffledPassword, nil
}

func (PUM *PasswordUtilManger) FnrandomVal(charSet string) (byte, error) {
	max := big.NewInt(int64(len(charSet)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return charSet[n.Int64()], nil
}

func (PUM *PasswordUtilManger) FnshuffleString(s string) string {
	r := []rune(s)
	for i := len(r) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		r[i], r[j.Int64()] = r[j.Int64()], r[i]
	}
	return string(r)
}

func (PUM *PasswordUtilManger) Fnencrypt(input string) string {
	var encrypted strings.Builder
	for i, c := range input {
		if i%2 == 0 {
			encrypted.WriteByte(byte(c) + 2)
		} else {
			encrypted.WriteByte(byte(c) + 1)
		}
	}
	return encrypted.String()
}

func (PUM *PasswordUtilManger) Fndecrypt(encrypted string) string {
	var decrypted strings.Builder
	for i, c := range encrypted {
		if i%2 == 0 {
			decrypted.WriteByte(byte(c) - 2)
		} else {
			decrypted.WriteByte(byte(c) - 1)
		}
	}
	return decrypted.String()
}

func (PUM *PasswordUtilManger) FnwritePasswordChangeToFile(IpPipeID, newPassword string) error {
	filePath := "/mnt/c/Users/devdu/go-workspace/data_fwd_tap/logs/password_changes.log"

	// Attempt to open or create the log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		PUM.LM.LogError("PasswordChange", "Error opening or creating file: %v", err)
		return fmt.Errorf("error opening or creating file: %v", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("Timestamp: %s | IpPipeID: %s | New Password: %s\n", timestamp, IpPipeID, newPassword)

	if _, err := file.WriteString(logEntry); err != nil {
		PUM.LM.LogError("PasswordChange", "Error writing to file: %v", err)
		return fmt.Errorf("error writing to file: %v", err)
	}

	// Log success
	PUM.LM.LogInfo("PasswordChange", "Successfully wrote password change to file for IpPipeID: %s", IpPipeID)

	return nil
}

func (PUM *PasswordUtilManger) CopyAndFormatPassword(dest []byte, destLen int, src string) {

	for i := 0; i < len(src) && i < destLen; i++ {
		dest[i] = src[i]
	}

	for i := len(src); i < destLen; i++ {
		dest[i] = ' '
	}
}
