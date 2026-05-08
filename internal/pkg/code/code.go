package code

import (
	"crypto/rand"
	"math/big"
	"fmt"
)


func GenerateSixDigitCode() (string, error){
	max := big.NewInt(1000000) 
	
	n, err := rand.Int(rand.Reader, max)
	
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n), nil
}