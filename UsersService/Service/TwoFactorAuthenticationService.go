package Service

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"image/png"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// TwoFactorAuthenticationService representa el servicio para la autenticación de dos factores.
type TwoFactorAuthenticationService struct{}

// NewTwoFactorAuthenticationService crea una nueva instancia del servicio de autenticación de dos factores.
func NewTwoFactorAuthenticationService() *TwoFactorAuthenticationService {
	return &TwoFactorAuthenticationService{}
}

func (s *TwoFactorAuthenticationService) GenerateSecretKey(mail string) string {

	secretKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "IAM",
		AccountName: mail,
	})
	if err != nil {
		fmt.Println("Error generating secret key:", err)
		return ""
	}
	return secretKey.Secret()
}

// VerifyCode verifica el código TOTP proporcionado por el usuario
func (s *TwoFactorAuthenticationService) VerifyCode(userCode, secretKey string) (bool, error) {
	_, err := base32.StdEncoding.DecodeString(secretKey)
	if err != nil {
		return false, fmt.Errorf("Invalid secret key: %v", err)
	}

	return totp.Validate(userCode, secretKey), nil
}

func (s *TwoFactorAuthenticationService) GenerateTotpUrl(secretKey, issuer, accountName string) string {
	decodedSecret, err := base32.StdEncoding.DecodeString(secretKey)
	if err != nil {
		fmt.Println("Error decoding secret key:", err)
		return ""
	}

	url, err := totp.Generate(totp.GenerateOpts{
		Secret:      decodedSecret,
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		fmt.Println("Error generating TOTP URL:", err)
		return ""
	}
	return url.String()
}

// GenerateQrCode genera un código QR a partir de la URL TOTP
func (s *TwoFactorAuthenticationService) GenerateQrCode(totpUrl string) ([]byte, error) {
	// Genera un código QR utilizando la URL TOTP
	qrCode, err := qrcode.New(totpUrl, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	// Codifica el código QR como PNG
	var pngBytes bytes.Buffer
	err = png.Encode(&pngBytes, qrCode.Image(256))
	if err != nil {
		return nil, err
	}

	return pngBytes.Bytes(), nil
}
